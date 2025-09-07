package backends

// BackendType represents the type of backend to create
type BackendType string

const (
	// GinBackendType represents the Gin web framework backend
	GinBackendType BackendType = "gin"
)

// BackendFactory provides factory methods for creating backends
type BackendFactory struct{}

// NewBackendFactory creates a new backend factory
func NewBackendFactory() *BackendFactory {
	return &BackendFactory{}
}

// CreateBackend creates a backend of the specified type
func (f *BackendFactory) CreateBackend(backendType BackendType) Backend {
	switch backendType {
	case GinBackendType:
		return NewGinBackend()
	default:
		return NewGinBackend() // Default to Gin
	}
}

// CreateCORSBackend creates a CORS-enabled backend of the specified type
func (f *BackendFactory) CreateCORSBackend(backendType BackendType) CORSBackend {
	switch backendType {
	case GinBackendType:
		return NewGinCORSBackend()
	default:
		return NewGinCORSBackend() // Default to Gin
	}
}

// GetAvailableBackends returns a list of available backend types
func (f *BackendFactory) GetAvailableBackends() []BackendType {
	return []BackendType{GinBackendType}
}