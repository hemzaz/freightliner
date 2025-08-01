package interfaces

// This file contains go:generate directives for creating mocks of all interfaces.
// The mocks are generated using gomock for better testability.

//go:generate go run github.com/golang/mock/mockgen -source=repository.go -destination=../mocks/repository_mocks.go -package=mocks
//go:generate go run github.com/golang/mock/mockgen -source=auth.go -destination=../mocks/auth_mocks.go -package=mocks
//go:generate go run github.com/golang/mock/mockgen -source=client.go -destination=../mocks/client_mocks.go -package=mocks

// Mock generation for segregated repository interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/reader_mock.go -package=mocks freightliner/pkg/interfaces Reader
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/writer_mock.go -package=mocks freightliner/pkg/interfaces Writer
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/image_provider_mock.go -package=mocks freightliner/pkg/interfaces ImageProvider
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/metadata_provider_mock.go -package=mocks freightliner/pkg/interfaces MetadataProvider
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/content_provider_mock.go -package=mocks freightliner/pkg/interfaces ContentProvider
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/content_manager_mock.go -package=mocks freightliner/pkg/interfaces ContentManager

// Mock generation for composition interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/repository_composer_mock.go -package=mocks freightliner/pkg/interfaces RepositoryComposer
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/read_write_repository_mock.go -package=mocks freightliner/pkg/interfaces ReadWriteRepository
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/image_repository_mock.go -package=mocks freightliner/pkg/interfaces ImageRepository
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/full_repository_mock.go -package=mocks freightliner/pkg/interfaces FullRepository

// Mock generation for context-aware interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/contextual_tag_lister_mock.go -package=mocks freightliner/pkg/interfaces ContextualTagLister
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/contextual_manifest_manager_mock.go -package=mocks freightliner/pkg/interfaces ContextualManifestManager
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/contextual_layer_accessor_mock.go -package=mocks freightliner/pkg/interfaces ContextualLayerAccessor
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/streaming_repository_mock.go -package=mocks freightliner/pkg/interfaces StreamingRepository

// Mock generation for authentication interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/token_provider_mock.go -package=mocks freightliner/pkg/interfaces TokenProvider
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/header_provider_mock.go -package=mocks freightliner/pkg/interfaces HeaderProvider
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/authenticator_provider_mock.go -package=mocks freightliner/pkg/interfaces AuthenticatorProvider
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/token_manager_mock.go -package=mocks freightliner/pkg/interfaces TokenManager
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/caching_authenticator_mock.go -package=mocks freightliner/pkg/interfaces CachingAuthenticator
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/multi_registry_authenticator_mock.go -package=mocks freightliner/pkg/interfaces MultiRegistryAuthenticator

// Mock generation for client interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/repository_lister_mock.go -package=mocks freightliner/pkg/interfaces RepositoryLister
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/repository_provider_mock.go -package=mocks freightliner/pkg/interfaces RepositoryProvider
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/registry_info_mock.go -package=mocks freightliner/pkg/interfaces RegistryInfo
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/paginated_repository_lister_mock.go -package=mocks freightliner/pkg/interfaces PaginatedRepositoryLister
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/caching_repository_provider_mock.go -package=mocks freightliner/pkg/interfaces CachingRepositoryProvider
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/batch_repository_provider_mock.go -package=mocks freightliner/pkg/interfaces BatchRepositoryProvider
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/health_checker_mock.go -package=mocks freightliner/pkg/interfaces HealthChecker

// Mock generation for streaming interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/streaming_repository_lister_mock.go -package=mocks freightliner/pkg/interfaces StreamingRepositoryLister

// Mock generation for multi-registry interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/multi_registry_client_mock.go -package=mocks freightliner/pkg/interfaces MultiRegistryClient
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/federated_client_mock.go -package=mocks freightliner/pkg/interfaces FederatedClient

// Mock generation for composition interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/basic_client_mock.go -package=mocks freightliner/pkg/interfaces BasicClient
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/enhanced_client_mock.go -package=mocks freightliner/pkg/interfaces EnhancedClient
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/caching_client_mock.go -package=mocks freightliner/pkg/interfaces CachingClient
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/full_client_mock.go -package=mocks freightliner/pkg/interfaces FullClient
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/client_composer_mock.go -package=mocks freightliner/pkg/interfaces ClientComposer

// Mock generation for legacy interfaces (for backward compatibility)
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/repository_mock.go -package=mocks freightliner/pkg/interfaces Repository
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/registry_client_mock.go -package=mocks freightliner/pkg/interfaces RegistryClient
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/registry_authenticator_mock.go -package=mocks freightliner/pkg/interfaces RegistryAuthenticator
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/registry_provider_mock.go -package=mocks freightliner/pkg/interfaces RegistryProvider
