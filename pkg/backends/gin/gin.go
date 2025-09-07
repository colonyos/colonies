// Package gin provides a wrapper around the Gin web framework
// This package exposes commonly used Gin functionality through a clean interface
// while maintaining compatibility with the underlying gin.Engine and gin.Context
package gin

import (
	"github.com/gin-gonic/gin"
)

// Default creates a new Engine instance with default middleware (Logger and Recovery)
func Default() *Engine {
	return NewEngineWithGin(gin.Default())
}

// New creates a new blank Engine instance without any middleware
func New() *Engine {
	return NewEngineWithGin(gin.New())
}

// SetMode sets the Gin mode (debug, release, test)
func SetMode(value string) {
	gin.SetMode(value)
}

// Mode returns the current Gin mode
func Mode() string {
	return gin.Mode()
}

// IsDebugging returns true if the framework is running in debug mode
func IsDebugging() bool {
	return gin.IsDebugging()
}

// Recovery middleware recovers from any panics and writes a 500 if there was one
func Recovery() HandlerFunc {
	ginRecovery := gin.Recovery()
	return func(c *Context) {
		ginRecovery(c.ginContext)
	}
}

// Logger returns a gin.HandlerFunc for logging
func Logger() HandlerFunc {
	ginLogger := gin.Logger()
	return func(c *Context) {
		ginLogger(c.ginContext)
	}
}

// LoggerWithFormatter returns a gin.HandlerFunc for logging with custom formatter
func LoggerWithFormatter(f gin.LogFormatter) HandlerFunc {
	ginLogger := gin.LoggerWithFormatter(f)
	return func(c *Context) {
		ginLogger(c.ginContext)
	}
}

// LoggerWithWriter returns a gin.HandlerFunc for logging with custom writer
func LoggerWithWriter(out gin.LogFormatter, notlogged ...string) HandlerFunc {
	ginLogger := gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: out,
		SkipPaths: notlogged,
	})
	return func(c *Context) {
		ginLogger(c.ginContext)
	}
}

// BasicAuth returns a Basic HTTP Authorization middleware
func BasicAuth(accounts gin.Accounts) HandlerFunc {
	ginBasicAuth := gin.BasicAuth(accounts)
	return func(c *Context) {
		ginBasicAuth(c.ginContext)
	}
}