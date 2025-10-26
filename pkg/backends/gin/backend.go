package gin

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/gin-contrib/cors"
	ginframework "github.com/gin-gonic/gin"
)

// Backend implements the complete Backend interface for Gin
type Backend struct{}

// NewBackend creates a new Gin backend implementation
func NewBackend() backends.Backend {
	return &Backend{}
}

func (g *Backend) NewEngine() backends.Engine {
	return NewEngineAdapter(New())
}

func (g *Backend) NewEngineWithDefaults() backends.Engine {
	return NewEngineAdapter(Default())
}

func (g *Backend) NewServer(port int, engine backends.Engine) backends.Server {
	ginEngineAdapter := engine.(*EngineAdapter)
	ginServer := NewServer(port, ginEngineAdapter.engine)
	return NewServerAdapter(ginServer)
}

func (g *Backend) NewServerWithAddr(addr string, engine backends.Engine) backends.Server {
	ginEngineAdapter := engine.(*EngineAdapter)
	ginServer := NewServerWithAddr(addr, ginEngineAdapter.engine)
	return NewServerAdapter(ginServer)
}

func (g *Backend) SetMode(mode string) {
	ginframework.SetMode(mode)
}

func (g *Backend) GetMode() string {
	return ginframework.Mode()
}

func (g *Backend) Logger() backends.MiddlewareFunc {
	return func(c backends.Context) {
		adapter := c.(*ContextAdapter)
		ginframework.Logger()(adapter.ginContext)
	}
}

func (g *Backend) Recovery() backends.MiddlewareFunc {
	return func(c backends.Context) {
		adapter := c.(*ContextAdapter)
		ginframework.Recovery()(adapter.ginContext)
	}
}

// CORSBackend implements CORSBackend interface
type CORSBackend struct {
	*Backend
}

// NewCORSBackend creates a new Gin CORS backend implementation
func NewCORSBackend() backends.CORSBackend {
	return &CORSBackend{
		Backend: &Backend{},
	}
}

func (g *CORSBackend) CORS() backends.MiddlewareFunc {
	corsMiddleware := cors.Default()
	return func(c backends.Context) {
		adapter := c.(*ContextAdapter)
		corsMiddleware(adapter.ginContext)
	}
}

func (g *CORSBackend) CORSWithConfig(config backends.CORSConfig) backends.MiddlewareFunc {
	// Convert generic CORSConfig to gin-contrib/cors Config
	corsConfig := cors.Config{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     config.AllowMethods,
		AllowHeaders:     config.AllowHeaders,
		ExposeHeaders:    config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	}
	corsMiddleware := cors.New(corsConfig)
	return func(c backends.Context) {
		adapter := c.(*ContextAdapter)
		corsMiddleware(adapter.ginContext)
	}
}

// ContextAdapter adapts gin.Context to the generic Context interface
type ContextAdapter struct {
	ginContext *ginframework.Context
}

// NewContextAdapter creates a new adapter for gin.Context
func NewContextAdapter(ginContext *ginframework.Context) backends.Context {
	return &ContextAdapter{ginContext: ginContext}
}

// Implement Context interface methods by delegating to gin.Context

func (g *ContextAdapter) String(code int, format string, values ...interface{}) {
	g.ginContext.String(code, format, values...)
}

func (g *ContextAdapter) JSON(code int, obj interface{}) {
	g.ginContext.JSON(code, obj)
}

func (g *ContextAdapter) XML(code int, obj interface{}) {
	g.ginContext.XML(code, obj)
}

func (g *ContextAdapter) Data(code int, contentType string, data []byte) {
	g.ginContext.Data(code, contentType, data)
}

func (g *ContextAdapter) Status(code int) {
	g.ginContext.Status(code)
}

func (g *ContextAdapter) Request() *http.Request {
	return g.ginContext.Request
}

func (g *ContextAdapter) ReadBody() ([]byte, error) {
	return io.ReadAll(g.ginContext.Request.Body)
}

func (g *ContextAdapter) GetHeader(key string) string {
	return g.ginContext.GetHeader(key)
}

func (g *ContextAdapter) Header(key, value string) {
	g.ginContext.Header(key, value)
}

func (g *ContextAdapter) Param(key string) string {
	return g.ginContext.Param(key)
}

func (g *ContextAdapter) Query(key string) string {
	return g.ginContext.Query(key)
}

func (g *ContextAdapter) DefaultQuery(key, defaultValue string) string {
	return g.ginContext.DefaultQuery(key, defaultValue)
}

func (g *ContextAdapter) PostForm(key string) string {
	return g.ginContext.PostForm(key)
}

func (g *ContextAdapter) DefaultPostForm(key, defaultValue string) string {
	return g.ginContext.DefaultPostForm(key, defaultValue)
}

func (g *ContextAdapter) Bind(obj interface{}) error {
	return g.ginContext.Bind(obj)
}

func (g *ContextAdapter) ShouldBind(obj interface{}) error {
	return g.ginContext.ShouldBind(obj)
}

func (g *ContextAdapter) BindJSON(obj interface{}) error {
	return g.ginContext.BindJSON(obj)
}

func (g *ContextAdapter) ShouldBindJSON(obj interface{}) error {
	return g.ginContext.ShouldBindJSON(obj)
}

func (g *ContextAdapter) Set(key string, value interface{}) {
	g.ginContext.Set(key, value)
}

func (g *ContextAdapter) Get(key string) (value interface{}, exists bool) {
	return g.ginContext.Get(key)
}

func (g *ContextAdapter) GetString(key string) string {
	return g.ginContext.GetString(key)
}

func (g *ContextAdapter) GetBool(key string) bool {
	return g.ginContext.GetBool(key)
}

func (g *ContextAdapter) GetInt(key string) int {
	return g.ginContext.GetInt(key)
}

func (g *ContextAdapter) GetInt64(key string) int64 {
	return g.ginContext.GetInt64(key)
}

func (g *ContextAdapter) GetFloat64(key string) float64 {
	return g.ginContext.GetFloat64(key)
}

func (g *ContextAdapter) Abort() {
	g.ginContext.Abort()
}

func (g *ContextAdapter) AbortWithStatus(code int) {
	g.ginContext.AbortWithStatus(code)
}

func (g *ContextAdapter) AbortWithStatusJSON(code int, jsonObj interface{}) {
	g.ginContext.AbortWithStatusJSON(code, jsonObj)
}

func (g *ContextAdapter) IsAborted() bool {
	return g.ginContext.IsAborted()
}

func (g *ContextAdapter) Next() {
	g.ginContext.Next()
}

// GinContext returns the underlying raw gin.Context - needed for realtime handler
func (g *ContextAdapter) GinContext() *ginframework.Context {
	return g.ginContext
}

// EngineAdapter adapts gin.Engine to the generic Engine interface
type EngineAdapter struct {
	engine *Engine
}

// NewEngineAdapter creates a new adapter for gin.Engine
func NewEngineAdapter(ginEngine *Engine) backends.Engine {
	return &EngineAdapter{engine: ginEngine}
}

func (g *EngineAdapter) GET(relativePath string, handlers ...backends.HandlerFunc) {
	g.engine.GinEngine().GET(relativePath, g.convertHandlers(handlers...)...)
}

func (g *EngineAdapter) POST(relativePath string, handlers ...backends.HandlerFunc) {
	g.engine.GinEngine().POST(relativePath, g.convertHandlers(handlers...)...)
}

func (g *EngineAdapter) PUT(relativePath string, handlers ...backends.HandlerFunc) {
	g.engine.GinEngine().PUT(relativePath, g.convertHandlers(handlers...)...)
}

func (g *EngineAdapter) DELETE(relativePath string, handlers ...backends.HandlerFunc) {
	g.engine.GinEngine().DELETE(relativePath, g.convertHandlers(handlers...)...)
}

func (g *EngineAdapter) PATCH(relativePath string, handlers ...backends.HandlerFunc) {
	g.engine.GinEngine().PATCH(relativePath, g.convertHandlers(handlers...)...)
}

func (g *EngineAdapter) Use(middleware ...backends.HandlerFunc) {
	g.engine.GinEngine().Use(g.convertHandlers(middleware...)...)
}

func (g *EngineAdapter) Handler() http.Handler {
	return g.engine.Handler()
}

// convertHandlers converts generic HandlerFunc directly to gin framework HandlerFunc
func (g *EngineAdapter) convertHandlers(handlers ...backends.HandlerFunc) []ginframework.HandlerFunc {
	ginHandlers := make([]ginframework.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		ginHandlers[i] = func(c *ginframework.Context) {
			adapter := NewContextAdapter(c)
			handler(adapter)
		}
	}
	return ginHandlers
}

// ServerAdapter adapts gin.Server to the generic Server interface
type ServerAdapter struct {
	ginServer *Server
	engine    backends.Engine
}

// NewServerAdapter creates a new adapter for gin.Server
func NewServerAdapter(ginServer *Server) backends.Server {
	return &ServerAdapter{
		ginServer: ginServer,
		engine:    NewEngineAdapter(ginServer.Engine()),
	}
}

func (g *ServerAdapter) ListenAndServe() error {
	return g.ginServer.ListenAndServe()
}

func (g *ServerAdapter) ListenAndServeTLS(certFile, keyFile string) error {
	return g.ginServer.ListenAndServeTLS(certFile, keyFile)
}

func (g *ServerAdapter) Shutdown(ctx context.Context) error {
	return g.ginServer.Shutdown(ctx)
}

func (g *ServerAdapter) ShutdownWithTimeout(timeout time.Duration) error {
	return g.ginServer.ShutdownWithTimeout(timeout)
}

func (g *ServerAdapter) SetAddr(addr string) {
	g.ginServer.SetAddr(addr)
}

func (g *ServerAdapter) GetAddr() string {
	return g.ginServer.GetAddr()
}

func (g *ServerAdapter) SetReadTimeout(timeout time.Duration) {
	g.ginServer.SetReadTimeout(timeout)
}

func (g *ServerAdapter) SetWriteTimeout(timeout time.Duration) {
	g.ginServer.SetWriteTimeout(timeout)
}

func (g *ServerAdapter) SetIdleTimeout(timeout time.Duration) {
	g.ginServer.SetIdleTimeout(timeout)
}

func (g *ServerAdapter) SetReadHeaderTimeout(timeout time.Duration) {
	g.ginServer.SetReadHeaderTimeout(timeout)
}

func (g *ServerAdapter) Engine() backends.Engine {
	return g.engine
}

func (g *ServerAdapter) HTTPServer() *http.Server {
	return g.ginServer.HTTPServer()
}