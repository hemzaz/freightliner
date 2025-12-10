package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"freightliner/pkg/distributed"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/storage"
)

// Example: Running a 3-node distributed cluster

func main() {
	// Parse command line flags
	nodeID := flag.String("node-id", "node-1", "Node ID")
	bindAddr := flag.String("bind-addr", "127.0.0.1:7000", "Raft bind address")
	grpcAddr := flag.String("grpc-addr", "127.0.0.1:7001", "gRPC bind address")
	dataDir := flag.String("data-dir", "/tmp/freightliner", "Data directory")
	bootstrap := flag.Bool("bootstrap", false, "Bootstrap cluster")
	joinAddr := flag.String("join", "", "Existing node address to join")
	flag.Parse()

	logger := log.NewBasicLogger(log.InfoLevel)

	logger.WithFields(map[string]interface{}{
		"node_id":   *nodeID,
		"bind_addr": *bindAddr,
		"grpc_addr": *grpcAddr,
		"bootstrap": *bootstrap,
	}).Info("Starting Freightliner distributed node")

	// Create Raft coordinator
	raftConfig := distributed.RaftConfig{
		NodeID:           *nodeID,
		BindAddr:         *bindAddr,
		DataDir:          fmt.Sprintf("%s/%s/raft", *dataDir, *nodeID),
		Bootstrap:        *bootstrap,
		Logger:           logger,
		HeartbeatTimeout: 1 * time.Second,
		ElectionTimeout:  3 * time.Second,
	}

	coordinator, err := distributed.NewRaftCoordinator(raftConfig)
	if err != nil {
		logger.Error("Failed to create Raft coordinator", err)
		os.Exit(1)
	}
	defer coordinator.Shutdown()

	// Wait for leader election
	logger.Info("Waiting for leader election...")
	if err := coordinator.WaitForLeader(10 * time.Second); err != nil {
		logger.Error("Failed to elect leader", err)
		os.Exit(1)
	}

	isLeader := coordinator.IsLeader()
	leaderAddr := coordinator.GetLeader()
	logger.WithFields(map[string]interface{}{
		"is_leader":   isLeader,
		"leader_addr": leaderAddr,
	}).Info("Raft cluster ready")

	// Create work stealing scheduler
	scheduler := distributed.NewWorkStealingScheduler(
		*nodeID,
		1000,  // Local queue capacity
		10000, // Global queue capacity
		logger,
	)
	defer scheduler.Stop()

	// Create content-addressable storage
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger:       logger,
		GCInterval:   1 * time.Hour,
		EnableCache:  true,
		MaxCacheSize: 1 * 1024 * 1024 * 1024, // 1GB cache
	})
	defer cas.Stop()

	// Create distributed cache
	cache := distributed.NewDistributedCache(distributed.CacheConfig{
		Logger:       logger,
		VirtualNodes: 150,
		Replication:  3,
	})

	// Add this node to distributed cache
	cacheNode := distributed.NewCacheNode(
		*nodeID,
		*grpcAddr,
		10*1024*1024*1024, // 10GB capacity
		nil,               // Local node, no client needed
	)
	if err := cache.AddNode(cacheNode); err != nil {
		logger.Error("Failed to add cache node", err)
		os.Exit(1)
	}

	// Create gRPC mesh
	meshConfig := distributed.MeshConfig{
		NodeID:  *nodeID,
		Address: *grpcAddr,
		Logger:  logger,
	}

	mesh, err := distributed.NewGRPCMesh(meshConfig)
	if err != nil {
		logger.Error("Failed to create gRPC mesh", err)
		os.Exit(1)
	}
	defer mesh.Stop()

	if err := mesh.Start(); err != nil {
		logger.Error("Failed to start gRPC mesh", err)
		os.Exit(1)
	}

	// Join existing cluster if specified
	if *joinAddr != "" {
		logger.WithFields(map[string]interface{}{
			"join_addr": *joinAddr,
		}).Info("Joining existing cluster")

		// Add peer to scheduler
		scheduler.AddPeer(&distributed.Peer{
			ID:      "existing-node",
			Address: *joinAddr,
		})

		// Connect via gRPC
		if err := mesh.ConnectToNode("existing-node", *joinAddr); err != nil {
			logger.Error("Failed to connect to existing node", err)
		}
	}

	// Start example workload
	go runExampleWorkload(coordinator, scheduler, cas, cache, logger)

	// Print cluster status periodically
	go printClusterStatus(coordinator, scheduler, cas, cache, logger)

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down gracefully...")
}

func runExampleWorkload(
	coordinator *distributed.RaftCoordinator,
	scheduler *distributed.WorkStealingScheduler,
	cas *storage.ContentAddressableStore,
	cache *distributed.DistributedCache,
	logger log.Logger,
) {
	ctx := context.Background()

	// Wait a bit for cluster to stabilize
	time.Sleep(5 * time.Second)

	logger.Info("Starting example workload")

	// Example 1: Store blob in CAS
	blobData := []byte("Example container layer data")
	digest, err := cas.Store(ctx, blobData)
	if err != nil {
		logger.Error("Failed to store blob", err)
		return
	}

	logger.WithFields(map[string]interface{}{
		"digest": digest.String(),
		"size":   len(blobData),
	}).Info("Stored blob in CAS")

	// Example 2: Cache manifest
	manifestData := []byte(`{"schemaVersion": 2, "config": {...}}`)
	if err := cache.Set(ctx, "manifest:nginx:latest", manifestData, 3600); err != nil {
		logger.Error("Failed to cache manifest", err)
		return
	}

	logger.Info("Cached manifest")

	// Example 3: Create job in Raft
	if coordinator.IsLeader() {
		job := &distributed.JobState{
			ID:         "example-job-1",
			Status:     "pending",
			NodeID:     "node-1",
			StartTime:  time.Now(),
			UpdateTime: time.Now(),
		}

		if err := coordinator.CreateJob(ctx, job); err != nil {
			logger.Error("Failed to create job", err)
			return
		}

		logger.WithFields(map[string]interface{}{
			"job_id": job.ID,
		}).Info("Created job in Raft")
	}

	// Example 4: Schedule work
	for i := 0; i < 5; i++ {
		job := &distributed.Job{
			ID:       fmt.Sprintf("replication-%d", i),
			Priority: 10,
			Task: func(ctx context.Context) error {
				logger.WithFields(map[string]interface{}{
					"job_id": fmt.Sprintf("replication-%d", i),
				}).Info("Executing replication task")

				// Simulate work
				time.Sleep(2 * time.Second)
				return nil
			},
		}

		if err := scheduler.Schedule(job); err != nil {
			logger.Error("Failed to schedule job", err)
		}
	}

	logger.Info("Scheduled 5 replication jobs")

	// Example 5: Retrieve cached manifest
	cachedData, err := cache.Get(ctx, "manifest:nginx:latest")
	if err != nil {
		logger.Error("Failed to retrieve cached manifest", err)
		return
	}

	logger.WithFields(map[string]interface{}{
		"size": len(cachedData),
	}).Info("Retrieved manifest from cache")

	// Example 6: Retrieve blob from CAS
	retrievedBlob, err := cas.Get(ctx, digest)
	if err != nil {
		logger.Error("Failed to retrieve blob", err)
		return
	}

	logger.WithFields(map[string]interface{}{
		"size":    len(retrievedBlob),
		"matches": string(retrievedBlob) == string(blobData),
	}).Info("Retrieved blob from CAS")

	logger.Info("Example workload completed")
}

func printClusterStatus(
	coordinator *distributed.RaftCoordinator,
	scheduler *distributed.WorkStealingScheduler,
	cas *storage.ContentAddressableStore,
	cache *distributed.DistributedCache,
	logger log.Logger,
) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Raft stats
		raftStats := coordinator.Stats()
		isLeader := coordinator.IsLeader()
		leaderAddr := coordinator.GetLeader()

		// Scheduler metrics
		schedMetrics := scheduler.GetMetrics()
		queueDepth := scheduler.GetQueueDepth()

		// CAS stats
		casStats := cas.GetStats()

		// Cache stats
		cacheStats := cache.GetStats()

		logger.WithFields(map[string]interface{}{
			"raft_state":     raftStats["state"],
			"is_leader":      isLeader,
			"leader":         leaderAddr,
			"jobs_scheduled": schedMetrics.JobsScheduled.Load(),
			"jobs_stolen":    schedMetrics.JobsStolen.Load(),
			"queue_depth":    queueDepth,
			"cas_blobs":      casStats["blob_count"],
			"cas_dedup_rate": casStats["dedup_rate"],
			"cache_hit_rate": cacheStats["hit_rate"],
			"cache_nodes":    cacheStats["nodes"],
		}).Info("Cluster status")
	}
}
