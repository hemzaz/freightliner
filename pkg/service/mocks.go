package service

// This file contains go:generate directives for creating mocks of service-specific interfaces.
// The mocks are generated using gomock for better testability.

//go:generate go run github.com/golang/mock/mockgen -source=interfaces.go -destination=../mocks/service_mocks.go -package=mocks

// Mock generation for service-specific interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/repository_creator_mock.go -package=mocks freightliner/pkg/service RepositoryCreator
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/replication_service_mock.go -package=mocks freightliner/pkg/service ReplicationService
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/monitoring_service_mock.go -package=mocks freightliner/pkg/service MonitoringService
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/health_service_mock.go -package=mocks freightliner/pkg/service HealthService

// Mock generation for composition interfaces
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/basic_service_mock.go -package=mocks freightliner/pkg/service BasicService
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/enhanced_service_mock.go -package=mocks freightliner/pkg/service EnhancedService
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/full_service_mock.go -package=mocks freightliner/pkg/service FullService
//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/service_composer_mock.go -package=mocks freightliner/pkg/service ServiceComposer
