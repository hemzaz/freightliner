package copy

// This file contains go:generate directives for creating mocks of copy-specific interfaces.
// The mocks are generated using gomock for better testability.

//go:generate go run github.com/golang/mock/mockgen -source=interfaces.go -destination=../mocks/copy_mocks.go -package=mocks

// Mock generation for segregated copy interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/source_reader_mock.go -package=mocks freightliner/pkg/copy SourceReader
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/destination_writer_mock.go -package=mocks freightliner/pkg/copy DestinationWriter
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/copy_repository_mock.go -package=mocks freightliner/pkg/copy Repository

// Mock generation for copy-specific interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/progress_reporter_mock.go -package=mocks freightliner/pkg/copy ProgressReporter
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/layer_processor_mock.go -package=mocks freightliner/pkg/copy LayerProcessor
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/manifest_processor_mock.go -package=mocks freightliner/pkg/copy ManifestProcessor
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/transfer_optimizer_mock.go -package=mocks freightliner/pkg/copy TransferOptimizer

// Mock generation for context-aware copy interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/contextual_copier_mock.go -package=mocks freightliner/pkg/copy ContextualCopier

// Mock generation for streaming copy interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/streaming_copier_mock.go -package=mocks freightliner/pkg/copy StreamingCopier

// Mock generation for composition interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/basic_copier_mock.go -package=mocks freightliner/pkg/copy BasicCopier
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/enhanced_copier_mock.go -package=mocks freightliner/pkg/copy EnhancedCopier
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/full_copier_mock.go -package=mocks freightliner/pkg/copy FullCopier
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/copy_composer_mock.go -package=mocks freightliner/pkg/copy CopyComposer

// Mock generation for metrics interface
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/metrics_mock.go -package=mocks freightliner/pkg/copy Metrics
