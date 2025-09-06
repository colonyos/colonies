package backend

import (
	"fmt"
)

// BackendFactory interface for creating backends
type BackendFactory interface {
	CreateBackend(config BackendConfig) (Backend, error)
	SupportedTypes() []BackendType
}

// DefaultBackendFactory implements the BackendFactory interface
type DefaultBackendFactory struct {
	creators map[BackendType]func(BackendConfig) (Backend, error)
}

// NewBackendFactory creates a new backend factory
func NewBackendFactory() *DefaultBackendFactory {
	return &DefaultBackendFactory{
		creators: make(map[BackendType]func(BackendConfig) (Backend, error)),
	}
}

// RegisterCreator registers a backend creator function
func (f *DefaultBackendFactory) RegisterCreator(backendType BackendType, creator func(BackendConfig) (Backend, error)) {
	f.creators[backendType] = creator
}

// CreateBackend creates a backend based on configuration
func (f *DefaultBackendFactory) CreateBackend(config BackendConfig) (Backend, error) {
	creator, exists := f.creators[config.Type]
	if !exists {
		return nil, fmt.Errorf("unsupported backend type: %s", config.Type)
	}
	return creator(config)
}

// SupportedTypes returns all supported backend types
func (f *DefaultBackendFactory) SupportedTypes() []BackendType {
	types := make([]BackendType, 0, len(f.creators))
	for backendType := range f.creators {
		types = append(types, backendType)
	}
	return types
}