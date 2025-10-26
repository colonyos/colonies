package gin

// HandlerFunc defines the signature for route handlers in our wrapper
type HandlerFunc func(*Context)

// MiddlewareFunc is an alias for HandlerFunc to represent middleware
type MiddlewareFunc = HandlerFunc