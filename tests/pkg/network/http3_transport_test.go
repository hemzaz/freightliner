package network

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"freightliner/pkg/network"
)

func TestHTTP3Transport_NewTransport(t *testing.T) {
	transport := network.NewHTTP3Transport(nil)
	if transport == nil {
		t.Fatal("expected transport to be created")
	}

	stats := transport.GetStats()
	if stats.HTTP3Requests != 0 {
		t.Errorf("expected zero requests, got %d", stats.HTTP3Requests)
	}
}

func TestHTTP3Transport_DoRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != "test response" {
		t.Errorf("expected 'test response', got '%s'", string(body))
	}
}

func TestHTTP3Transport_StreamDownload(t *testing.T) {
	data := []byte("large test data for streaming")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	var buf bytes.Buffer
	ctx := context.Background()
	n, err := transport.StreamDownload(ctx, server.URL, &buf)
	if err != nil {
		t.Fatalf("stream download failed: %v", err)
	}

	if n != int64(len(data)) {
		t.Errorf("expected %d bytes, got %d", len(data), n)
	}

	if !bytes.Equal(buf.Bytes(), data) {
		t.Error("downloaded data doesn't match")
	}
}

func TestHTTP3Transport_ParallelDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	urls := []string{server.URL, server.URL, server.URL}
	writers := []io.Writer{&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}}

	ctx := context.Background()
	err := transport.ParallelDownload(ctx, urls, writers)
	if err != nil {
		t.Fatalf("parallel download failed: %v", err)
	}
}

func TestHTTP3Transport_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var buf bytes.Buffer
	_, err := transport.StreamDownload(ctx, server.URL, &buf)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestHTTP3Transport_Stats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	// Make some requests
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		transport.Do(req)
	}

	stats := transport.GetStats()
	totalRequests := stats.HTTP3Requests + stats.HTTP2Requests + stats.HTTP1Requests
	if totalRequests == 0 {
		t.Error("expected non-zero total requests")
	}
}

func BenchmarkHTTP3Transport_SingleRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("benchmark data"))
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, err := transport.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkHTTP3Transport_ParallelDownload(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(make([]byte, 1024)) // 1 KB
	}))
	defer server.Close()

	transport := network.NewHTTP3Transport(nil)
	defer transport.Close()

	urls := make([]string, 10)
	for i := range urls {
		urls[i] = server.URL
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writers := make([]io.Writer, 10)
		for j := range writers {
			writers[j] = io.Discard
		}

		ctx := context.Background()
		if err := transport.ParallelDownload(ctx, urls, writers); err != nil {
			b.Fatal(err)
		}
	}
}
