package distributed

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"sync"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// GRPCMesh manages gRPC connections between cluster nodes
type GRPCMesh struct {
	nodeID      string
	address     string
	server      *grpc.Server
	clients     map[string]*GRPCClient
	logger      log.Logger
	mu          sync.RWMutex
	tlsConfig   *tls.Config
	interceptor *MeshInterceptor
}

// GRPCClient represents a connection to a remote node
type GRPCClient struct {
	NodeID   string
	Address  string
	conn     *grpc.ClientConn
	client   ClusterServiceClient
	healthy  bool
	lastPing time.Time
	mu       sync.RWMutex
}

// MeshInterceptor provides request/response interception
type MeshInterceptor struct {
	logger  log.Logger
	metrics *MeshMetrics
}

// MeshMetrics tracks mesh performance
type MeshMetrics struct {
	RequestsTotal    uint64
	RequestsSuccess  uint64
	RequestsFailed   uint64
	AvgLatency       time.Duration
	ConnectionsTotal int
	mu               sync.RWMutex
}

// ClusterServiceClient is a placeholder for the generated gRPC client
// In production, this would be generated from the proto file
type ClusterServiceClient interface {
	SubmitJob(ctx context.Context, req *SubmitJobRequest) (*SubmitJobResponse, error)
	GetJobStatus(ctx context.Context, req *GetJobStatusRequest) (*JobStatusResponse, error)
	StealWork(ctx context.Context, req *StealWorkRequest) (*StealWorkResponse, error)
	GetBlob(ctx context.Context, req *GetBlobRequest) (BlobStream, error)
	PutBlob(ctx context.Context) (BlobUploadStream, error)
	CacheGet(ctx context.Context, req *CacheGetRequest) (*CacheGetResponse, error)
	CacheSet(ctx context.Context, req *CacheSetRequest) (*CacheSetResponse, error)
	HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error)
}

// BlobStream represents a streaming blob download
type BlobStream interface {
	Recv() (*BlobChunk, error)
}

// BlobUploadStream represents a streaming blob upload
type BlobUploadStream interface {
	Send(*BlobChunk) error
	CloseAndRecv() (*PutBlobResponse, error)
}

// Request/Response types (would be generated from proto)
type SubmitJobRequest struct {
	JobID                 string
	SourceRegistry        string
	SourceRepository      string
	DestinationRegistry   string
	DestinationRepository string
	Tags                  []string
	Priority              int
}

type SubmitJobResponse struct {
	JobID        string
	Status       string
	AssignedNode string
}

type GetJobStatusRequest struct {
	JobID string
}

type JobStatusResponse struct {
	JobID      string
	Status     string
	NodeID     string
	StartTime  time.Time
	UpdateTime time.Time
	Progress   *JobProgress
}

type JobProgress struct {
	TotalTags        int64
	CompletedTags    int64
	FailedTags       int64
	BytesTransferred int64
}

type StealWorkRequest struct {
	NodeID  string
	MaxJobs int
}

type StealWorkResponse struct {
	Jobs []*Job
}

type GetBlobRequest struct {
	Digest string
	Offset int64
	Size   int64
}

type BlobChunk struct {
	Digest    string
	Data      []byte
	Offset    int64
	TotalSize int64
	IsLast    bool
}

type PutBlobResponse struct {
	Digest       string
	Size         int64
	Deduplicated bool
}

type CacheGetRequest struct {
	Key string
}

type CacheGetResponse struct {
	Value []byte
	Found bool
}

type CacheSetRequest struct {
	Key   string
	Value []byte
	TTL   int64
}

type CacheSetResponse struct {
	Success bool
}

type HealthCheckRequest struct {
	NodeID string
}

type HealthCheckResponse struct {
	Status    string
	Timestamp time.Time
	Health    *NodeHealth
}

type NodeHealth struct {
	CPUUsage      float64
	MemoryUsage   float64
	ActiveJobs    int
	QueueSize     int
	IsLeader      bool
	LeaderAddress string
}

// MeshConfig holds gRPC mesh configuration
type MeshConfig struct {
	NodeID    string
	Address   string
	TLSConfig *tls.Config
	Logger    log.Logger
}

// NewGRPCMesh creates a new gRPC service mesh
func NewGRPCMesh(config MeshConfig) (*GRPCMesh, error) {
	if config.Logger == nil {
		config.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	mesh := &GRPCMesh{
		nodeID:    config.NodeID,
		address:   config.Address,
		clients:   make(map[string]*GRPCClient),
		logger:    config.Logger,
		tlsConfig: config.TLSConfig,
		interceptor: &MeshInterceptor{
			logger:  config.Logger,
			metrics: &MeshMetrics{},
		},
	}

	// Create gRPC server
	serverOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(100 * 1024 * 1024), // 100MB
		grpc.MaxSendMsgSize(100 * 1024 * 1024), // 100MB
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 10 * time.Second,
		}),
		grpc.UnaryInterceptor(mesh.interceptor.UnaryServerInterceptor),
		grpc.StreamInterceptor(mesh.interceptor.StreamServerInterceptor),
	}

	if config.TLSConfig != nil {
		serverOpts = append(serverOpts, grpc.Creds(credentials.NewTLS(config.TLSConfig)))
	}

	mesh.server = grpc.NewServer(serverOpts...)

	return mesh, nil
}

// Start starts the gRPC server
func (gm *GRPCMesh) Start() error {
	// In production, this would register the ClusterService implementation
	// RegisterClusterServiceServer(gm.server, &clusterServiceImpl{mesh: gm})

	gm.logger.WithFields(map[string]interface{}{
		"node_id": gm.nodeID,
		"address": gm.address,
	}).Info("Starting gRPC mesh server")

	// This is a placeholder - actual implementation would start the server
	return nil
}

// ConnectToNode establishes a connection to a remote node
func (gm *GRPCMesh) ConnectToNode(nodeID, address string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if _, exists := gm.clients[nodeID]; exists {
		return errors.AlreadyExistsf("already connected to node: %s", nodeID)
	}

	// Create connection options
	dialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100*1024*1024),
			grpc.MaxCallSendMsgSize(100*1024*1024),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithUnaryInterceptor(gm.interceptor.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(gm.interceptor.StreamClientInterceptor),
	}

	if gm.tlsConfig != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(gm.tlsConfig)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Establish connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, dialOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to connect to node")
	}

	// Create client
	client := &GRPCClient{
		NodeID:   nodeID,
		Address:  address,
		conn:     conn,
		healthy:  true,
		lastPing: time.Now(),
	}

	gm.clients[nodeID] = client

	gm.logger.WithFields(map[string]interface{}{
		"node_id": nodeID,
		"address": address,
	}).Info("Connected to node")

	// Start health checking
	go gm.healthCheckNode(nodeID)

	return nil
}

// DisconnectFromNode closes connection to a remote node
func (gm *GRPCMesh) DisconnectFromNode(nodeID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	client, exists := gm.clients[nodeID]
	if !exists {
		return errors.NotFoundf("not connected to node: %s", nodeID)
	}

	if err := client.conn.Close(); err != nil {
		gm.logger.WithFields(map[string]interface{}{
			"node_id": nodeID,
			"error":   err.Error(),
		}).Warn("Error closing connection")
	}

	delete(gm.clients, nodeID)

	gm.logger.WithFields(map[string]interface{}{
		"node_id": nodeID,
	}).Info("Disconnected from node")

	return nil
}

// GetClient returns a client for a specific node
func (gm *GRPCMesh) GetClient(nodeID string) (*GRPCClient, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	client, exists := gm.clients[nodeID]
	if !exists {
		return nil, errors.NotFoundf("not connected to node: %s", nodeID)
	}

	if !client.IsHealthy() {
		return nil, errors.New("node is unhealthy")
	}

	return client, nil
}

// BroadcastToAll sends a request to all connected nodes
func (gm *GRPCMesh) BroadcastToAll(ctx context.Context, fn func(*GRPCClient) error) []error {
	gm.mu.RLock()
	clients := make([]*GRPCClient, 0, len(gm.clients))
	for _, client := range gm.clients {
		clients = append(clients, client)
	}
	gm.mu.RUnlock()

	errorsCh := make(chan error, len(clients))
	var wg sync.WaitGroup

	for _, client := range clients {
		wg.Add(1)
		go func(c *GRPCClient) {
			defer wg.Done()
			if err := fn(c); err != nil {
				errorsCh <- err
			}
		}(client)
	}

	wg.Wait()
	close(errorsCh)

	var errs []error
	for err := range errorsCh {
		errs = append(errs, err)
	}

	return errs
}

// GetMetrics returns mesh metrics
func (gm *GRPCMesh) GetMetrics() *MeshMetrics {
	return gm.interceptor.metrics
}

// Stop stops the gRPC server and closes all connections
func (gm *GRPCMesh) Stop() error {
	// Close all client connections
	gm.mu.Lock()
	for nodeID := range gm.clients {
		gm.DisconnectFromNode(nodeID)
	}
	gm.mu.Unlock()

	// Stop server
	gm.server.GracefulStop()

	gm.logger.Info("gRPC mesh stopped")

	return nil
}

// healthCheckNode performs periodic health checks
func (gm *GRPCMesh) healthCheckNode(nodeID string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		gm.mu.RLock()
		client, exists := gm.clients[nodeID]
		gm.mu.RUnlock()

		if !exists {
			return
		}

		req := &HealthCheckRequest{NodeID: gm.nodeID}

		// This would call the actual gRPC method
		// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// resp, err := client.client.HealthCheck(ctx, req)
		// cancel()

		// Placeholder for health check logic
		var err error
		_ = req

		client.mu.Lock()
		if err != nil {
			client.healthy = false
			gm.logger.WithFields(map[string]interface{}{
				"node_id": nodeID,
				"error":   err.Error(),
			}).Warn("Health check failed")
		} else {
			client.healthy = true
			client.lastPing = time.Now()
		}
		client.mu.Unlock()
	}
}

// GRPCClient methods

// IsHealthy returns the health status
func (gc *GRPCClient) IsHealthy() bool {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.healthy
}

// GetLastPing returns the last successful ping time
func (gc *GRPCClient) GetLastPing() time.Time {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.lastPing
}

// Interceptor methods

// UnaryServerInterceptor intercepts unary RPC calls on server side
func (mi *MeshInterceptor) UnaryServerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	mi.metrics.mu.Lock()
	mi.metrics.RequestsTotal++
	mi.metrics.mu.Unlock()

	resp, err := handler(ctx, req)

	duration := time.Since(start)

	mi.metrics.mu.Lock()
	if err != nil {
		mi.metrics.RequestsFailed++
	} else {
		mi.metrics.RequestsSuccess++
	}
	mi.metrics.AvgLatency = duration
	mi.metrics.mu.Unlock()

	mi.logger.WithFields(map[string]interface{}{
		"method":   info.FullMethod,
		"duration": duration,
		"error":    err != nil,
	}).Debug("gRPC call completed")

	return resp, err
}

// StreamServerInterceptor intercepts stream RPC calls on server side
func (mi *MeshInterceptor) StreamServerInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()

	mi.metrics.mu.Lock()
	mi.metrics.RequestsTotal++
	mi.metrics.mu.Unlock()

	err := handler(srv, ss)

	duration := time.Since(start)

	mi.metrics.mu.Lock()
	if err != nil && err != io.EOF {
		mi.metrics.RequestsFailed++
	} else {
		mi.metrics.RequestsSuccess++
	}
	mi.metrics.AvgLatency = duration
	mi.metrics.mu.Unlock()

	return err
}

// UnaryClientInterceptor intercepts unary RPC calls on client side
func (mi *MeshInterceptor) UnaryClientInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	start := time.Now()

	err := invoker(ctx, method, req, reply, cc, opts...)

	duration := time.Since(start)

	mi.logger.WithFields(map[string]interface{}{
		"method":   method,
		"duration": duration,
		"error":    err != nil,
	}).Debug("gRPC client call completed")

	return err
}

// StreamClientInterceptor intercepts stream RPC calls on client side
func (mi *MeshInterceptor) StreamClientInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	return streamer(ctx, desc, cc, method, opts...)
}

// Helper function to format gRPC errors
func formatGRPCError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("gRPC error: %v", err)
}
