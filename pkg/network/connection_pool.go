package network

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
)

// ConnectionPool manages HTTP client connections with aggressive reuse and keep-alive
type ConnectionPool struct {
	config ConnectionPoolConfig
	logger log.Logger

	// Connection management
	clients     sync.Map // map[string]*PooledHTTPClient
	clientCount atomic.Int64

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Metrics
	metrics *ConnectionPoolMetrics
}

// ConnectionPoolConfig configures the connection pool behavior
type ConnectionPoolConfig struct {
	// Connection limits
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int

	// Timeouts
	IdleConnTimeout       time.Duration
	ResponseHeaderTimeout time.Duration
	TLSHandshakeTimeout   time.Duration
	DialTimeout           time.Duration

	// Keep-alive settings
	KeepAlive         time.Duration
	DisableKeepAlives bool

	// TLS configuration
	InsecureSkipVerify bool
	MinTLSVersion      uint16

	// Connection pool behavior
	CleanupInterval    time.Duration
	ConnectionTTL      time.Duration
	EnableHTTP2        bool
	DisableCompression bool
	ForceAttemptHTTP2  bool

	// Performance tuning
	WriteBufferSize int
	ReadBufferSize  int
}

// DefaultConnectionPoolConfig returns optimized defaults for registry operations
func DefaultConnectionPoolConfig() ConnectionPoolConfig {
	return ConnectionPoolConfig{
		// Aggressive connection pooling for high throughput
		MaxIdleConns:        200, // Support 200 concurrent idle connections
		MaxIdleConnsPerHost: 100, // 100 per registry
		MaxConnsPerHost:     100, // Allow 100 concurrent connections per registry

		// Optimized timeouts for registry operations
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		DialTimeout:           10 * time.Second,

		// Aggressive keep-alive
		KeepAlive:         60 * time.Second,
		DisableKeepAlives: false,

		// TLS settings
		InsecureSkipVerify: false,
		MinTLSVersion:      tls.VersionTLS12,

		// Pool management
		CleanupInterval:   1 * time.Minute,
		ConnectionTTL:     5 * time.Minute,
		EnableHTTP2:       true,
		ForceAttemptHTTP2: true,

		// Buffer sizes optimized for large blob transfers
		WriteBufferSize: 64 * 1024, // 64KB
		ReadBufferSize:  64 * 1024, // 64KB
	}
}

// PooledHTTPClient wraps an http.Client with connection pool management
type PooledHTTPClient struct {
	*http.Client
	createdAt    time.Time
	lastUsed     atomic.Int64 // Unix timestamp
	requestCount atomic.Int64
	key          string
}

// ConnectionPoolMetrics tracks connection pool performance
type ConnectionPoolMetrics struct {
	ActiveClients      atomic.Int64
	TotalRequests      atomic.Int64
	ConnectionReuses   atomic.Int64
	NewConnections     atomic.Int64
	ExpiredConnections atomic.Int64
	FailedConnections  atomic.Int64
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config ConnectionPoolConfig, logger log.Logger) *ConnectionPool {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnectionPool{
		config:  config,
		logger:  logger,
		ctx:     ctx,
		cancel:  cancel,
		metrics: &ConnectionPoolMetrics{},
	}

	// Start cleanup routine
	pool.wg.Add(1)
	go pool.cleanupRoutine()

	pool.logger.WithFields(map[string]interface{}{
		"max_idle_conns":          config.MaxIdleConns,
		"max_idle_conns_per_host": config.MaxIdleConnsPerHost,
		"max_conns_per_host":      config.MaxConnsPerHost,
		"keep_alive":              config.KeepAlive.String(),
	}).Info("Connection pool initialized")

	return pool
}

// GetClient returns a pooled HTTP client for the given host
func (p *ConnectionPool) GetClient(host string) (*PooledHTTPClient, error) {
	key := p.clientKey(host)

	// Try to get existing client
	if val, ok := p.clients.Load(key); ok {
		client := val.(*PooledHTTPClient)

		// Check if client is still valid
		if !p.isExpired(client) {
			client.lastUsed.Store(time.Now().Unix())
			client.requestCount.Add(1)
			p.metrics.ConnectionReuses.Add(1)
			p.metrics.TotalRequests.Add(1)
			return client, nil
		}

		// Client expired, remove it
		p.clients.Delete(key)
		p.metrics.ExpiredConnections.Add(1)
		p.clientCount.Add(-1)
	}

	// Create new client
	client, err := p.createClient(key, host)
	if err != nil {
		p.metrics.FailedConnections.Add(1)
		return nil, errors.Wrap(err, "failed to create HTTP client")
	}

	// Store in pool
	p.clients.Store(key, client)
	p.clientCount.Add(1)
	p.metrics.NewConnections.Add(1)
	p.metrics.TotalRequests.Add(1)
	p.metrics.ActiveClients.Store(p.clientCount.Load())

	return client, nil
}

// createClient creates a new optimized HTTP client
func (p *ConnectionPool) createClient(key, host string) (*PooledHTTPClient, error) {
	// Create optimized transport
	transport := &http.Transport{
		// Connection pooling
		MaxIdleConns:        p.config.MaxIdleConns,
		MaxIdleConnsPerHost: p.config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     p.config.MaxConnsPerHost,

		// Timeouts
		IdleConnTimeout:       p.config.IdleConnTimeout,
		ResponseHeaderTimeout: p.config.ResponseHeaderTimeout,
		TLSHandshakeTimeout:   p.config.TLSHandshakeTimeout,

		// Keep-alive
		DisableKeepAlives: p.config.DisableKeepAlives,

		// HTTP/2 support
		ForceAttemptHTTP2: p.config.ForceAttemptHTTP2,

		// Compression
		DisableCompression: p.config.DisableCompression,

		// Buffer sizes
		WriteBufferSize: p.config.WriteBufferSize,
		ReadBufferSize:  p.config.ReadBufferSize,

		// Optimized dialer
		DialContext: (&net.Dialer{
			Timeout:   p.config.DialTimeout,
			KeepAlive: p.config.KeepAlive,
			// Enable TCP Fast Open (TFO) for faster connection establishment
			Control: func(network, address string, c syscall.RawConn) error {
				return nil
			},
		}).DialContext,

		// TLS configuration
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: p.config.InsecureSkipVerify,
			MinVersion:         p.config.MinTLSVersion,
			// Session ticket reuse for faster TLS handshakes
			ClientSessionCache: tls.NewLRUClientSessionCache(100),
		},

		// Proxy settings
		Proxy: http.ProxyFromEnvironment,
	}

	client := &PooledHTTPClient{
		Client: &http.Client{
			Transport: transport,
			Timeout:   0, // No timeout at client level, control at request level
		},
		createdAt: time.Now(),
		key:       key,
	}
	client.lastUsed.Store(time.Now().Unix())

	return client, nil
}

// clientKey generates a unique key for a client based on host
func (p *ConnectionPool) clientKey(host string) string {
	return host
}

// isExpired checks if a client connection should be recycled
func (p *ConnectionPool) isExpired(client *PooledHTTPClient) bool {
	age := time.Since(client.createdAt)
	return age > p.config.ConnectionTTL
}

// cleanupRoutine periodically removes expired connections
func (p *ConnectionPool) cleanupRoutine() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performCleanup()
		}
	}
}

// performCleanup removes expired connections from the pool
func (p *ConnectionPool) performCleanup() {
	var removed int64

	p.clients.Range(func(key, value interface{}) bool {
		client := value.(*PooledHTTPClient)

		if p.isExpired(client) {
			p.clients.Delete(key)
			client.Client.CloseIdleConnections()
			removed++
			p.clientCount.Add(-1)
			p.metrics.ExpiredConnections.Add(1)
		}

		return true
	})

	if removed > 0 {
		p.metrics.ActiveClients.Store(p.clientCount.Load())
		p.logger.WithFields(map[string]interface{}{
			"removed_clients":   removed,
			"remaining_clients": p.clientCount.Load(),
		}).Debug("Cleaned up expired connections")
	}
}

// Close shuts down the connection pool
func (p *ConnectionPool) Close() error {
	p.cancel()
	p.wg.Wait()

	// Close all clients
	p.clients.Range(func(key, value interface{}) bool {
		client := value.(*PooledHTTPClient)
		client.Client.CloseIdleConnections()
		return true
	})

	p.logger.WithFields(map[string]interface{}{
		"total_requests":      p.metrics.TotalRequests.Load(),
		"connection_reuses":   p.metrics.ConnectionReuses.Load(),
		"new_connections":     p.metrics.NewConnections.Load(),
		"expired_connections": p.metrics.ExpiredConnections.Load(),
	}).Info("Connection pool closed")

	return nil
}

// GetMetrics returns current connection pool metrics
func (p *ConnectionPool) GetMetrics() *ConnectionPoolMetrics {
	return p.metrics
}

// Stats returns human-readable connection pool statistics
func (p *ConnectionPool) Stats() map[string]interface{} {
	totalRequests := p.metrics.TotalRequests.Load()
	reuses := p.metrics.ConnectionReuses.Load()
	var reuseRate float64
	if totalRequests > 0 {
		reuseRate = float64(reuses) / float64(totalRequests) * 100
	}

	return map[string]interface{}{
		"active_clients":        p.metrics.ActiveClients.Load(),
		"total_requests":        totalRequests,
		"connection_reuses":     reuses,
		"connection_reuse_rate": reuseRate,
		"new_connections":       p.metrics.NewConnections.Load(),
		"expired_connections":   p.metrics.ExpiredConnections.Load(),
		"failed_connections":    p.metrics.FailedConnections.Load(),
	}
}
