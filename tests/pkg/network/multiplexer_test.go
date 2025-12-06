package network

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"freightliner/pkg/network"
)

func TestMultiplexer_DownloadLayers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("layer data"))
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	multiplexer := network.NewStreamMultiplexer(transport, nil)

	layers := []network.LayerDescriptor{
		{URL: server.URL, Digest: "sha256:abc123", Size: 10, Writer: &bytes.Buffer{}, Priority: 1},
		{URL: server.URL, Digest: "sha256:def456", Size: 10, Writer: &bytes.Buffer{}, Priority: 2},
		{URL: server.URL, Digest: "sha256:ghi789", Size: 10, Writer: &bytes.Buffer{}, Priority: 3},
	}

	ctx := context.Background()
	err := multiplexer.DownloadLayers(ctx, layers)
	if err != nil {
		t.Fatalf("download layers failed: %v", err)
	}

	stats := multiplexer.GetStats()
	if stats.CompletedLayers != 3 {
		t.Errorf("expected 3 completed layers, got %d", stats.CompletedLayers)
	}

	if stats.FailedLayers != 0 {
		t.Errorf("expected 0 failed layers, got %d", stats.FailedLayers)
	}
}

func TestMultiplexer_EmptyLayers(t *testing.T) {
	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	multiplexer := network.NewStreamMultiplexer(transport, nil)

	ctx := context.Background()
	err := multiplexer.DownloadLayers(ctx, []network.LayerDescriptor{})
	if err != nil {
		t.Fatalf("empty layers should not error: %v", err)
	}
}

func TestMultiplexer_Priority(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	multiplexer := network.NewStreamMultiplexer(transport, nil)

	// Layers with different priorities
	layers := []network.LayerDescriptor{
		{URL: server.URL, Digest: "low", Size: 4, Writer: &bytes.Buffer{}, Priority: 1},
		{URL: server.URL, Digest: "high", Size: 4, Writer: &bytes.Buffer{}, Priority: 10},
		{URL: server.URL, Digest: "medium", Size: 4, Writer: &bytes.Buffer{}, Priority: 5},
	}

	ctx := context.Background()
	err := multiplexer.DownloadLayers(ctx, layers)
	if err != nil {
		t.Fatalf("download with priority failed: %v", err)
	}
}

func TestMultiplexer_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	multiplexer := network.NewStreamMultiplexer(transport, nil)

	layers := []network.LayerDescriptor{
		{URL: server.URL, Digest: "test", Size: 0, Writer: &bytes.Buffer{}, Priority: 1},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := multiplexer.DownloadLayers(ctx, layers)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestMultiplexer_Stats(t *testing.T) {
	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	multiplexer := network.NewStreamMultiplexer(transport, nil)

	// Initial stats
	stats := multiplexer.GetStats()
	if stats.TotalLayers != 0 {
		t.Errorf("expected 0 total layers, got %d", stats.TotalLayers)
	}

	// Reset stats
	multiplexer.Reset()
	stats = multiplexer.GetStats()
	if stats.TotalLayers != 0 || stats.CompletedLayers != 0 {
		t.Error("reset should clear all stats")
	}
}

func TestBatchMultiplexer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("batch data"))
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	batchMux := network.NewBatchMultiplexer(transport, 2)
	defer batchMux.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	batchMux.Start(ctx)

	layers := []network.LayerDescriptor{
		{URL: server.URL, Digest: "batch1", Size: 10, Writer: &bytes.Buffer{}, Priority: 1},
		{URL: server.URL, Digest: "batch2", Size: 10, Writer: &bytes.Buffer{}, Priority: 1},
	}

	batch := network.StreamBatch{
		Layers:   layers,
		Priority: 1,
		Context:  ctx,
	}

	err := batchMux.SubmitBatch(batch)
	if err != nil {
		t.Fatalf("submit batch failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // Give time for processing
}

func BenchmarkMultiplexer_Sequential(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(make([]byte, 1024)) // 1 KB
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	multiplexer := network.NewStreamMultiplexer(transport, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layers := []network.LayerDescriptor{
			{URL: server.URL, Digest: "l1", Size: 1024, Writer: &bytes.Buffer{}, Priority: 1},
			{URL: server.URL, Digest: "l2", Size: 1024, Writer: &bytes.Buffer{}, Priority: 1},
			{URL: server.URL, Digest: "l3", Size: 1024, Writer: &bytes.Buffer{}, Priority: 1},
		}

		ctx := context.Background()
		if err := multiplexer.DownloadLayers(ctx, layers); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMultiplexer_Parallel(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(make([]byte, 1024)) // 1 KB
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	config := &network.MultiplexerConfig{
		MaxStreams:     100,
		StreamTimeout:  30 * time.Second,
		RetryAttempts:  3,
		BufferSize:     64 * 1024,
		EnablePriority: true,
	}
	multiplexer := network.NewStreamMultiplexer(transport, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create 10 layers to download in parallel
		layers := make([]network.LayerDescriptor, 10)
		for j := range layers {
			layers[j] = network.LayerDescriptor{
				URL:      server.URL,
				Digest:   "layer",
				Size:     1024,
				Writer:   &bytes.Buffer{},
				Priority: 1,
			}
		}

		ctx := context.Background()
		if err := multiplexer.DownloadLayers(ctx, layers); err != nil {
			b.Fatal(err)
		}
	}
}
