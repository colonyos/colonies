package registry

// HandlerRegistrar defines the interface for handlers that can register themselves
type HandlerRegistrar interface {
	RegisterHandlers(registry *HandlerRegistry) error
}

// GlobalRegistry is the global handler registry instance
var GlobalRegistry = NewHandlerRegistry()

// RegisterHandler is a convenience function to register handlers globally
func RegisterHandler(payloadType string, handler HandlerFunc) error {
	return GlobalRegistry.Register(payloadType, handler)
}