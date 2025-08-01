package copy

import (
	"context"
	"io"
	"time"

	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Import types from the shared interfaces package for compatibility
type (
	RepositoryName        = interfaces.RepositoryName
	ImageReferencer       = interfaces.ImageReferencer
	RemoteOptionsProvider = interfaces.RemoteOptionsProvider
	ImageGetter           = interfaces.ImageGetter
	Manifest              = interfaces.Manifest
	ManifestAccessor      = interfaces.ManifestAccessor
	LayerAccessor         = interfaces.LayerAccessor
)

// ===== SEGREGATED COPY INTERFACES =====

// SourceReader defines the interface for reading from a source repository
type SourceReader interface {
	RepositoryName
	ImageReferencer
	RemoteOptionsProvider
	ImageGetter
	ManifestAccessor
}

// DestinationWriter defines the interface for writing to a destination repository
type DestinationWriter interface {
	RepositoryName
	ImageReferencer
	RemoteOptionsProvider

	// PutManifest uploads a manifest with the given tag
	PutManifest(ctx context.Context, tag string, manifest *Manifest) error

	// PutLayer uploads a layer with the given content
	PutLayer(ctx context.Context, digest string, content io.Reader) error
}

// Repository represents a container repository interface needed for copy operations
// This is a local interface that defines exactly what operations the copy package
// requires from a repository, following the Interface Segregation Principle.
// It's intentionally more limited than interfaces.Repository.
type Repository interface {
	SourceReader
	DestinationWriter
}

// ===== COPY-SPECIFIC INTERFACES =====

// ProgressReporter provides progress reporting capabilities
type ProgressReporter interface {
	// ReportProgress reports the progress of a copy operation
	ReportProgress(ctx context.Context, progress *CopyProgress) error

	// ReportError reports an error during copy operation
	ReportError(ctx context.Context, err error) error

	// ReportCompletion reports the completion of a copy operation
	ReportCompletion(ctx context.Context, result *CopyResult) error
}

// CopyProgress represents the progress of a copy operation
type CopyProgress struct {
	SourceRef        string
	DestinationRef   string
	Stage            CopyStage
	BytesTransferred int64
	TotalBytes       int64
	LayersCompleted  int
	TotalLayers      int
	StartTime        time.Time
	CurrentTime      time.Time
}

// CopyStage represents the current stage of the copy operation
type CopyStage string

const (
	StageInitializing    CopyStage = "initializing"
	StageFetchingSource  CopyStage = "fetching_source"
	StageCopyingLayers   CopyStage = "copying_layers"
	StagePushingManifest CopyStage = "pushing_manifest"
	StageCompleted       CopyStage = "completed"
	StageFailed          CopyStage = "failed"
)

// LayerProcessor provides layer processing capabilities
type LayerProcessor interface {
	// ProcessLayer processes a layer during copy (e.g., compression, deduplication)
	ProcessLayer(ctx context.Context, layer v1.Layer) (v1.Layer, error)

	// ShouldSkipLayer determines if a layer should be skipped
	ShouldSkipLayer(ctx context.Context, digest string) (bool, error)
}

// ManifestProcessor provides manifest processing capabilities
type ManifestProcessor interface {
	// ProcessManifest processes a manifest during copy (e.g., modification, validation)
	ProcessManifest(ctx context.Context, manifest *Manifest) (*Manifest, error)

	// ValidateManifest validates a manifest before copy
	ValidateManifest(ctx context.Context, manifest *Manifest) error
}

// TransferOptimizer provides transfer optimization capabilities
type TransferOptimizer interface {
	// OptimizeTransfer optimizes the transfer strategy based on source and destination
	OptimizeTransfer(ctx context.Context, sourceRef, destRef name.Reference) (*TransferStrategy, error)

	// GetTransferStrategy returns the current transfer strategy
	GetTransferStrategy() *TransferStrategy
}

// TransferStrategy defines how a copy operation should be performed
type TransferStrategy struct {
	UseParallelTransfer bool
	MaxConcurrentLayers int
	UseCompression      bool
	UseDeltaTransfer    bool
	ChunkSize           int64
	RetryAttempts       int
	RetryDelay          time.Duration
}

// ===== CONTEXT-AWARE COPY INTERFACES =====

// ContextualCopier provides context-aware copy operations
type ContextualCopier interface {
	// CopyWithContext copies an image with full context support
	CopyWithContext(ctx context.Context, options *CopyOptionsWithContext) (*CopyResult, error)

	// CopyBatchWithContext copies multiple images with context support
	CopyBatchWithContext(ctx context.Context, requests []*CopyRequest) ([]*CopyResult, error)
}

// CopyOptionsWithContext extends copy options with context-aware features
type CopyOptionsWithContext struct {
	Source             name.Reference
	Destination        name.Reference
	SourceOptions      []remote.Option
	DestinationOptions []remote.Option
	DryRun             bool
	ForceOverwrite     bool

	// Context-aware options
	ProgressReporter  ProgressReporter
	LayerProcessor    LayerProcessor
	ManifestProcessor ManifestProcessor
	TransferOptimizer TransferOptimizer
	MaxRetries        int
	Timeout           time.Duration
}

// CopyRequest represents a single copy request in a batch operation
type CopyRequest struct {
	Options  *CopyOptionsWithContext
	Priority int
	Tags     []string
}

// ===== STREAMING COPY INTERFACES =====

// StreamingCopier provides streaming copy capabilities for large operations
type StreamingCopier interface {
	// StreamCopy copies images from a stream of copy requests
	StreamCopy(ctx context.Context, requests <-chan *CopyRequest) (<-chan *CopyResult, <-chan error)

	// StreamCopyWithBuffer copies images with buffering for performance
	StreamCopyWithBuffer(ctx context.Context, requests <-chan *CopyRequest, bufferSize int) (<-chan *CopyResult, <-chan error)
}

// ===== COMPOSITION INTERFACES =====

// BasicCopier provides basic copy functionality
type BasicCopier interface {
	SourceReader
	DestinationWriter
}

// EnhancedCopier provides enhanced copy functionality with processing
type EnhancedCopier interface {
	BasicCopier
	LayerProcessor
	ManifestProcessor
}

// FullCopier provides all copy functionality
type FullCopier interface {
	EnhancedCopier
	ContextualCopier
	StreamingCopier
	TransferOptimizer
}

// CopyComposer provides composition of copy behaviors
type CopyComposer interface {
	// AsSourceReader returns a source reader view
	AsSourceReader() SourceReader

	// AsDestinationWriter returns a destination writer view
	AsDestinationWriter() DestinationWriter

	// AsLayerProcessor returns a layer processor view
	AsLayerProcessor() LayerProcessor

	// AsManifestProcessor returns a manifest processor view
	AsManifestProcessor() ManifestProcessor

	// AsTransferOptimizer returns a transfer optimizer view
	AsTransferOptimizer() TransferOptimizer

	// AsContextualCopier returns a contextual copier view
	AsContextualCopier() ContextualCopier

	// AsStreamingCopier returns a streaming copier view
	AsStreamingCopier() StreamingCopier
}
