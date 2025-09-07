package gin

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Engine wraps gin.Engine with additional functionality
type Engine struct {
	ginEngine *gin.Engine
}

// NewEngine creates a new Engine wrapper
func NewEngine() *Engine {
	ginEngine := gin.Default()
	return &Engine{
		ginEngine: ginEngine,
	}
}

// NewEngineWithGin creates a new Engine wrapper with an existing gin.Engine
func NewEngineWithGin(ginEngine *gin.Engine) *Engine {
	return &Engine{
		ginEngine: ginEngine,
	}
}

// UseCORS adds CORS middleware with default configuration
func (e *Engine) UseCORS() {
	e.ginEngine.Use(cors.Default())
}

// UseCORSWithConfig adds CORS middleware with custom configuration
func (e *Engine) UseCORSWithConfig(config cors.Config) {
	e.ginEngine.Use(cors.New(config))
}

// POST adds a POST route handler
func (e *Engine) POST(relativePath string, handlers ...HandlerFunc) {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		ginHandlers[i] = convertToGinHandler(handler)
	}
	e.ginEngine.POST(relativePath, ginHandlers...)
}

// GET adds a GET route handler
func (e *Engine) GET(relativePath string, handlers ...HandlerFunc) {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		ginHandlers[i] = convertToGinHandler(handler)
	}
	e.ginEngine.GET(relativePath, ginHandlers...)
}

// PUT adds a PUT route handler
func (e *Engine) PUT(relativePath string, handlers ...HandlerFunc) {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		ginHandlers[i] = convertToGinHandler(handler)
	}
	e.ginEngine.PUT(relativePath, ginHandlers...)
}

// DELETE adds a DELETE route handler
func (e *Engine) DELETE(relativePath string, handlers ...HandlerFunc) {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		ginHandlers[i] = convertToGinHandler(handler)
	}
	e.ginEngine.DELETE(relativePath, ginHandlers...)
}

// PATCH adds a PATCH route handler
func (e *Engine) PATCH(relativePath string, handlers ...HandlerFunc) {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		ginHandlers[i] = convertToGinHandler(handler)
	}
	e.ginEngine.PATCH(relativePath, ginHandlers...)
}

// Use adds middleware to the engine
func (e *Engine) Use(middleware ...HandlerFunc) {
	for _, mw := range middleware {
		e.ginEngine.Use(convertToGinHandler(mw))
	}
}

// Handler returns the underlying gin.Engine as http.Handler
func (e *Engine) Handler() http.Handler {
	return e.ginEngine
}

// GinEngine returns the underlying gin.Engine
func (e *Engine) GinEngine() *gin.Engine {
	return e.ginEngine
}

// convertToGinHandler converts our HandlerFunc to gin.HandlerFunc
func convertToGinHandler(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &Context{ginContext: c}
		handler(ctx)
	}
}