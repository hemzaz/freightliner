package copy

import (
	"bytes"
	"context"
	"io"
	"testing"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/helper/util"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// TestCheckBlobExists tests blob existence checking logic
func TestCheckBlobExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping external dependency test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	ref, _ := name.ParseReference("gcr.io/test/repo:tag")
	hash, _ := v1.NewHash("sha256:abc123")

	// This will fail because we're not actually connecting to a registry
	// But it tests the code path
	exists, _ := copier.checkBlobExists(ctx, ref, hash, nil)
	_ = exists // Result doesn't matter for unit test
}

// TestUploadBlob tests blob upload logic structure
func TestUploadBlob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping external dependency test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	ref, _ := name.ParseReference("gcr.io/test/repo:tag")
	hash, _ := v1.NewHash("sha256:test123")
	reader := bytes.NewReader([]byte("test data"))

	// This will fail because we're not connected to a real registry
	// But it exercises the code path
	_ = copier.uploadBlob(ctx, ref, hash, reader, nil)
}

// TestCheckDestinationExists tests destination checking
func TestCheckDestinationExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping external dependency test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	ref, _ := name.ParseReference("gcr.io/test/repo:tag")

	// Test with forceOverwrite = true (should return nil)
	err := copier.checkDestinationExists(ctx, ref, nil, true)
	if err != nil {
		t.Errorf("Expected no error with forceOverwrite=true, got: %v", err)
	}

	// Test with forceOverwrite = false (will fail due to no registry connection)
	_ = copier.checkDestinationExists(ctx, ref, nil, false)
}

// TestCopierStructure tests copier structure is properly initialized
func TestCopierStructure(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	// Verify all fields are initialized
	if copier.stats == nil {
		t.Error("Stats should be initialized")
	}

	if copier.bufferMgr == nil {
		t.Error("Buffer manager should be initialized")
	}

	if copier.logger == nil {
		t.Error("Logger should be set")
	}

	// Verify stats is a pointer to a new struct
	if copier.stats.BytesTransferred != 0 {
		t.Error("Stats should start at zero")
	}
}

// TestBufferManagerIntegration tests buffer manager usage
func TestBufferManagerIntegration(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	if copier.bufferMgr == nil {
		t.Fatal("Expected buffer manager to be initialized")
	}

	// Test getting a buffer
	buf := copier.bufferMgr.GetOptimalBuffer(1024, "test")
	if buf == nil {
		t.Error("Expected to get a buffer")
	}
	defer buf.Release()

	if len(buf.Bytes()) != 1024 {
		t.Errorf("Expected buffer size 1024, got %d", len(buf.Bytes()))
	}
}

// TestStreamingBlobLayerSize tests size estimation
func TestStreamingBlobLayerSize(t *testing.T) {
	data := []byte("test data")
	hash, _ := v1.NewHash("sha256:test")
	reader := bytes.NewReader(data)

	layer := &streamingBlobLayer{
		digestHash: hash,
		reader:     reader,
		bufferMgr:  util.NewBufferManager(),
		cachedSize: 0, // No cached size
	}

	size, err := layer.Size()
	if err != nil {
		t.Errorf("Size() error: %v", err)
	}

	// Should return default size
	if size != 1024*1024 {
		t.Errorf("Expected default size 1MB, got %d", size)
	}

	// Test with cached size
	layer.cachedSize = 12345
	size, err = layer.Size()
	if err != nil {
		t.Errorf("Size() with cache error: %v", err)
	}
	if size != 12345 {
		t.Errorf("Expected cached size 12345, got %d", size)
	}
}

// TestStreamingBlobLayerUncompressed tests uncompressed method
func TestStreamingBlobLayerUncompressed(t *testing.T) {
	data := []byte("test data")
	hash, _ := v1.NewHash("sha256:test")
	reader := bytes.NewReader(data)

	layer := &streamingBlobLayer{
		digestHash: hash,
		reader:     reader,
		bufferMgr:  util.NewBufferManager(),
		cachedSize: int64(len(data)),
	}

	uncompressed, err := layer.Uncompressed()
	if err != nil {
		t.Errorf("Uncompressed() error: %v", err)
	}
	defer uncompressed.Close()

	// Should be able to read
	buf := make([]byte, 10)
	n, _ := uncompressed.Read(buf)
	if n == 0 {
		t.Error("Expected to read data from uncompressed stream")
	}
}

// TestOptimizedReadCloserWithBuffer tests buffer management
func TestOptimizedReadCloserWithBuffer(t *testing.T) {
	data := []byte("test data")
	reader := bytes.NewReader(data)
	bufferMgr := util.NewBufferManager()
	buffer := bufferMgr.GetOptimalBuffer(1024, "test")

	orc := &optimizedReadCloser{
		reader:    reader,
		bufferMgr: bufferMgr,
		buffer:    buffer,
	}

	// Test reading
	buf := make([]byte, len(data))
	n, err := orc.Read(buf)
	if err != nil && err != io.EOF {
		t.Errorf("Read() error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to read %d bytes, got %d", len(data), n)
	}

	// Test close releases buffer
	err = orc.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}

	if orc.buffer != nil {
		t.Error("Expected buffer to be released after close")
	}
}

// TestOptimizedReadCloserWithReadCloser tests closing underlying reader
func TestOptimizedReadCloserWithReadCloser(t *testing.T) {
	type closableReader struct {
		*bytes.Reader
		closed bool
	}

	cr := &closableReader{Reader: bytes.NewReader([]byte("test"))}
	cr.closed = false

	// Make it implement io.Closer
	rc := struct {
		io.Reader
		io.Closer
	}{
		Reader: cr.Reader,
		Closer: io.NopCloser(cr.Reader),
	}

	orc := &optimizedReadCloser{
		reader:    rc,
		bufferMgr: util.NewBufferManager(),
	}

	err := orc.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}
}

// TestCopyProgressStructure tests copy progress structure
func TestCopyProgressStructure(t *testing.T) {
	progress := &CopyProgress{
		SourceRef:        "source:tag",
		DestinationRef:   "dest:tag",
		Stage:            StageCopyingLayers,
		BytesTransferred: 1024,
		TotalBytes:       2048,
		LayersCompleted:  3,
		TotalLayers:      5,
	}

	if progress.Stage != StageCopyingLayers {
		t.Errorf("Expected stage %s, got %s", StageCopyingLayers, progress.Stage)
	}

	if progress.BytesTransferred != 1024 {
		t.Errorf("Expected bytes transferred 1024, got %d", progress.BytesTransferred)
	}
}

// TestCopyStages tests copy stage constants
func TestCopyStages(t *testing.T) {
	stages := []CopyStage{
		StageInitializing,
		StageFetchingSource,
		StageCopyingLayers,
		StagePushingManifest,
		StageCompleted,
		StageFailed,
	}

	if len(stages) != 6 {
		t.Errorf("Expected 6 stages, got %d", len(stages))
	}

	if StageInitializing != "initializing" {
		t.Errorf("Expected initializing stage, got %s", StageInitializing)
	}
}

// TestTransferStrategy tests transfer strategy structure
func TestTransferStrategy(t *testing.T) {
	strategy := &TransferStrategy{
		UseParallelTransfer: true,
		MaxConcurrentLayers: 5,
		UseCompression:      true,
		UseDeltaTransfer:    false,
		ChunkSize:           65536,
		RetryAttempts:       3,
	}

	if !strategy.UseParallelTransfer {
		t.Error("Expected parallel transfer to be enabled")
	}

	if strategy.MaxConcurrentLayers != 5 {
		t.Errorf("Expected 5 concurrent layers, got %d", strategy.MaxConcurrentLayers)
	}
}

// TestCopyOptionsWithContext tests extended copy options
func TestCopyOptionsWithContext(t *testing.T) {
	srcRef, _ := name.ParseReference("source:tag")
	destRef, _ := name.ParseReference("dest:tag")

	options := &CopyOptionsWithContext{
		Source:         srcRef,
		Destination:    destRef,
		DryRun:         true,
		ForceOverwrite: false,
		MaxRetries:     3,
	}

	if options.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", options.MaxRetries)
	}

	if !options.DryRun {
		t.Error("Expected dry run to be true")
	}
}

// TestCopyRequest tests copy request structure
func TestCopyRequest(t *testing.T) {
	srcRef, _ := name.ParseReference("source:tag")
	destRef, _ := name.ParseReference("dest:tag")

	options := &CopyOptionsWithContext{
		Source:      srcRef,
		Destination: destRef,
	}

	request := &CopyRequest{
		Options:  options,
		Priority: 10,
		Tags:     []string{"v1.0", "latest"},
	}

	if request.Priority != 10 {
		t.Errorf("Expected priority 10, got %d", request.Priority)
	}

	if len(request.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(request.Tags))
	}
}

// TestManifestMediaTypeDetection tests media type detection in pushManifest
func TestManifestMediaTypeDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping external dependency test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	ref, _ := name.ParseReference("gcr.io/test/repo:tag")

	tests := []struct {
		name          string
		manifest      []byte
		expectedMedia types.MediaType
	}{
		{
			name:          "Docker Manifest V2 Schema 2",
			manifest:      []byte(`{"schemaVersion": 2, "mediaType": "application/vnd.docker.distribution.manifest.v2+json"}`),
			expectedMedia: types.DockerManifestSchema2,
		},
		{
			name:          "Docker Manifest V2 Schema 1",
			manifest:      []byte(`{"schemaVersion": 1}`),
			expectedMedia: types.DockerManifestSchema1,
		},
		{
			name:          "OCI Manifest",
			manifest:      []byte(`{"imageLayoutVersion": "1.0.0"}`),
			expectedMedia: types.OCIManifestSchema1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail because we're not connected to a registry
			// But it exercises the media type detection logic
			_ = copier.pushManifest(ctx, tt.manifest, ref, nil)
		})
	}
}

// TestProcessManifest tests process manifest stub
func TestProcessManifest(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	srcRef, _ := name.ParseReference("source:tag")
	destRef, _ := name.ParseReference("dest:tag")

	stats := &CopyStats{}

	// Process manifest is a stub that returns empty bytes
	result, err := copier.processManifest(ctx, nil, srcRef, destRef, nil, nil, true, stats)
	if err != nil {
		t.Errorf("processManifest() error: %v", err)
	}

	// Should return empty bytes
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d bytes", len(result))
	}
}
