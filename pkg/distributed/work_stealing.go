package distributed

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
)

// WorkStealingScheduler implements work stealing for dynamic load balancing
type WorkStealingScheduler struct {
	localQueue  *lockFreeDeque
	globalQueue *ConcurrentQueue
	peers       []*Peer
	nodeID      string
	logger      log.Logger
	mu          sync.RWMutex
	metrics     *SchedulerMetrics
	stopped     atomic.Bool
	stopCh      chan struct{}
}

// SchedulerMetrics tracks scheduler performance
type SchedulerMetrics struct {
	JobsScheduled  atomic.Uint64
	JobsStolen     atomic.Uint64
	LocalHits      atomic.Uint64
	GlobalHits     atomic.Uint64
	StealAttempts  atomic.Uint64
	StealSuccesses atomic.Uint64
	AvgQueueDepth  atomic.Uint64
}

// Peer represents a remote scheduler node
type Peer struct {
	ID       string
	Address  string
	queue    *lockFreeDeque
	capacity int64
	client   PeerClient
	mu       sync.RWMutex
}

// PeerClient defines the interface for peer communication
type PeerClient interface {
	GetQueueSize(ctx context.Context) (int, error)
	StealJob(ctx context.Context) (*Job, error)
	SubmitJob(ctx context.Context, job *Job) error
	GetCapacity(ctx context.Context) (int, error)
}

// Job represents a replication job for scheduling
type Job struct {
	ID         string
	Priority   int
	SubmitTime time.Time
	Deadline   time.Time
	Task       func(ctx context.Context) error
	Metadata   map[string]string
	RetryCount int
	MaxRetries int
}

// lockFreeDeque implements a lock-free double-ended queue
type lockFreeDeque struct {
	head   atomic.Pointer[dequeNode]
	tail   atomic.Pointer[dequeNode]
	size   atomic.Int64
	maxCap int64
}

type dequeNode struct {
	job  *Job
	next atomic.Pointer[dequeNode]
	prev atomic.Pointer[dequeNode]
}

// ConcurrentQueue implements a thread-safe global queue
type ConcurrentQueue struct {
	items  []*Job
	mu     sync.RWMutex
	cond   *sync.Cond
	maxCap int
}

// NewWorkStealingScheduler creates a new work stealing scheduler
func NewWorkStealingScheduler(nodeID string, localCap, globalCap int, logger log.Logger) *WorkStealingScheduler {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	ws := &WorkStealingScheduler{
		localQueue:  newLockFreeDeque(localCap),
		globalQueue: NewConcurrentQueue(globalCap),
		peers:       make([]*Peer, 0),
		nodeID:      nodeID,
		logger:      logger,
		metrics:     &SchedulerMetrics{},
		stopCh:      make(chan struct{}),
	}

	// Start background worker for queue management
	go ws.manageQueues()

	return ws
}

// newLockFreeDeque creates a new lock-free deque
func newLockFreeDeque(capacity int) *lockFreeDeque {
	d := &lockFreeDeque{
		maxCap: int64(capacity),
	}
	// Initialize with sentinel nodes
	sentinel := &dequeNode{}
	d.head.Store(sentinel)
	d.tail.Store(sentinel)
	return d
}

// PushBack adds a job to the back of the deque
func (d *lockFreeDeque) PushBack(job *Job) bool {
	if d.size.Load() >= d.maxCap {
		return false
	}

	newNode := &dequeNode{job: job}

	for {
		tail := d.tail.Load()
		if d.tail.CompareAndSwap(tail, newNode) {
			tail.next.Store(newNode)
			newNode.prev.Store(tail)
			d.size.Add(1)
			return true
		}
	}
}

// PopFront removes and returns a job from the front
func (d *lockFreeDeque) PopFront() *Job {
	for {
		head := d.head.Load()
		next := head.next.Load()

		if next == nil {
			return nil
		}

		if d.head.CompareAndSwap(head, next) {
			d.size.Add(-1)
			return next.job
		}
	}
}

// PopBack removes and returns a job from the back (for stealing)
func (d *lockFreeDeque) PopBack() *Job {
	for {
		tail := d.tail.Load()
		prev := tail.prev.Load()

		if prev == nil || prev == d.head.Load() {
			return nil
		}

		if d.tail.CompareAndSwap(tail, prev) {
			prev.next.Store(nil)
			d.size.Add(-1)
			return tail.job
		}
	}
}

// Len returns the current size
func (d *lockFreeDeque) Len() int64 {
	return d.size.Load()
}

// Cap returns the capacity
func (d *lockFreeDeque) Cap() int64 {
	return d.maxCap
}

// NewConcurrentQueue creates a new concurrent queue
func NewConcurrentQueue(capacity int) *ConcurrentQueue {
	q := &ConcurrentQueue{
		items:  make([]*Job, 0, capacity),
		maxCap: capacity,
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Push adds a job to the queue
func (q *ConcurrentQueue) Push(job *Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= q.maxCap {
		return errors.New("global queue is full")
	}

	q.items = append(q.items, job)
	q.cond.Signal()
	return nil
}

// Pop removes and returns a job from the queue
func (q *ConcurrentQueue) Pop() *Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	job := q.items[0]
	q.items = q.items[1:]
	return job
}

// WaitPop waits for a job to be available
func (q *ConcurrentQueue) WaitPop(timeout time.Duration) *Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	deadline := time.Now().Add(timeout)
	for len(q.items) == 0 {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil
		}

		// Wait with timeout
		go func() {
			time.Sleep(remaining)
			q.cond.Broadcast()
		}()
		q.cond.Wait()
	}

	if len(q.items) == 0 {
		return nil
	}

	job := q.items[0]
	q.items = q.items[1:]
	return job
}

// Len returns the queue size
func (q *ConcurrentQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

// Schedule schedules a job using work stealing algorithm
func (ws *WorkStealingScheduler) Schedule(job *Job) error {
	ws.metrics.JobsScheduled.Add(1)

	// Try local queue first (fast path)
	if ws.localQueue.PushBack(job) {
		ws.metrics.LocalHits.Add(1)
		ws.logger.WithFields(map[string]interface{}{
			"job_id": job.ID,
			"queue":  "local",
		}).Debug("Job scheduled to local queue")
		return nil
	}

	// Try to find underutilized peer
	ws.mu.RLock()
	peers := ws.peers
	ws.mu.RUnlock()

	for _, peer := range peers {
		peerSize, err := peer.GetQueueSize()
		if err != nil {
			continue
		}

		capacity, err := peer.GetCapacity()
		if err != nil {
			continue
		}

		// If peer is less than 50% utilized, send job there
		if peerSize < capacity/2 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := peer.client.SubmitJob(ctx, job); err == nil {
				ws.logger.WithFields(map[string]interface{}{
					"job_id":  job.ID,
					"peer_id": peer.ID,
				}).Debug("Job scheduled to peer")
				return nil
			}
		}
	}

	// Fall back to global queue
	if err := ws.globalQueue.Push(job); err != nil {
		return errors.Wrap(err, "failed to schedule job")
	}

	ws.metrics.GlobalHits.Add(1)
	ws.logger.WithFields(map[string]interface{}{
		"job_id": job.ID,
		"queue":  "global",
	}).Debug("Job scheduled to global queue")

	return nil
}

// StealWork attempts to steal work from other nodes
func (ws *WorkStealingScheduler) StealWork(ctx context.Context) *Job {
	ws.metrics.StealAttempts.Add(1)

	// Check local queue first (owner priority)
	if job := ws.localQueue.PopFront(); job != nil {
		ws.metrics.LocalHits.Add(1)
		return job
	}

	// Try to steal from busiest peer
	ws.mu.RLock()
	peers := ws.peers
	ws.mu.RUnlock()

	// Randomize peer order to avoid hotspots
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	var busiestPeer *Peer
	maxSize := 0

	for _, peer := range peers {
		size, err := peer.GetQueueSize()
		if err != nil {
			continue
		}

		if size > maxSize && size > 1 { // Only steal if peer has multiple jobs
			maxSize = size
			busiestPeer = peer
		}
	}

	if busiestPeer != nil {
		stealCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if job, err := busiestPeer.client.StealJob(stealCtx); err == nil && job != nil {
			ws.metrics.JobsStolen.Add(1)
			ws.metrics.StealSuccesses.Add(1)
			ws.logger.WithFields(map[string]interface{}{
				"job_id":  job.ID,
				"peer_id": busiestPeer.ID,
			}).Debug("Successfully stole job from peer")
			return job
		}
	}

	// Check global queue
	if job := ws.globalQueue.Pop(); job != nil {
		ws.metrics.GlobalHits.Add(1)
		return job
	}

	return nil
}

// AddPeer adds a peer to the scheduler
func (ws *WorkStealingScheduler) AddPeer(peer *Peer) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.peers = append(ws.peers, peer)
	ws.logger.WithFields(map[string]interface{}{
		"peer_id":  peer.ID,
		"address":  peer.Address,
		"capacity": peer.capacity,
	}).Info("Added peer to scheduler")
}

// RemovePeer removes a peer from the scheduler
func (ws *WorkStealingScheduler) RemovePeer(peerID string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for i, peer := range ws.peers {
		if peer.ID == peerID {
			ws.peers = append(ws.peers[:i], ws.peers[i+1:]...)
			ws.logger.WithFields(map[string]interface{}{
				"peer_id": peerID,
			}).Info("Removed peer from scheduler")
			return
		}
	}
}

// GetMetrics returns scheduler metrics
func (ws *WorkStealingScheduler) GetMetrics() *SchedulerMetrics {
	return ws.metrics
}

// GetQueueDepth returns current queue depth
func (ws *WorkStealingScheduler) GetQueueDepth() int64 {
	return ws.localQueue.Len()
}

// GetCapacity returns local queue capacity
func (ws *WorkStealingScheduler) GetCapacity() int64 {
	return ws.localQueue.Cap()
}

// Stop stops the scheduler
func (ws *WorkStealingScheduler) Stop() {
	if ws.stopped.CompareAndSwap(false, true) {
		close(ws.stopCh)
	}
}

// manageQueues performs background queue management
func (ws *WorkStealingScheduler) manageQueues() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update average queue depth metric
			depth := uint64(ws.localQueue.Len())
			ws.metrics.AvgQueueDepth.Store(depth)

			// Log metrics periodically
			if depth > 0 {
				ws.logger.WithFields(map[string]interface{}{
					"local_depth":    depth,
					"global_depth":   ws.globalQueue.Len(),
					"jobs_scheduled": ws.metrics.JobsScheduled.Load(),
					"jobs_stolen":    ws.metrics.JobsStolen.Load(),
					"steal_success":  float64(ws.metrics.StealSuccesses.Load()) / float64(ws.metrics.StealAttempts.Load()+1),
				}).Debug("Scheduler metrics")
			}

		case <-ws.stopCh:
			return
		}
	}
}

// GetQueueSize returns the size of the peer's queue
func (p *Peer) GetQueueSize() (int, error) {
	if p.queue != nil {
		return int(p.queue.Len()), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return p.client.GetQueueSize(ctx)
}

// GetCapacity returns the peer's capacity
func (p *Peer) GetCapacity() (int, error) {
	if p.capacity > 0 {
		return int(p.capacity), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	capacity, err := p.client.GetCapacity(ctx)
	if err == nil {
		p.mu.Lock()
		p.capacity = int64(capacity)
		p.mu.Unlock()
	}

	return capacity, err
}
