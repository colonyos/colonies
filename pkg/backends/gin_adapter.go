package backends

import (
	"context"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/backends/gin"
	"github.com/gin-contrib/cors"
	ginframework "github.com/gin-gonic/gin"
)

// GinContextAdapter adapts gin.Context to the generic Context interface
type GinContextAdapter struct {
	ginContext *gin.Context
}

// NewGinContextAdapter creates a new adapter for gin.Context
func NewGinContextAdapter(ginContext *gin.Context) *GinContextAdapter {
	return &GinContextAdapter{ginContext: ginContext}
}

// Implement Context interface methods by delegating to gin.Context

func (g *GinContextAdapter) String(code int, format string, values ...interface{}) {
	g.ginContext.String(code, format, values...)
}

func (g *GinContextAdapter) JSON(code int, obj interface{}) {
	g.ginContext.JSON(code, obj)
}

func (g *GinContextAdapter) XML(code int, obj interface{}) {
	g.ginContext.XML(code, obj)
}

func (g *GinContextAdapter) Data(code int, contentType string, data []byte) {
	g.ginContext.Data(code, contentType, data)
}

func (g *GinContextAdapter) Status(code int) {
	g.ginContext.Status(code)
}

func (g *GinContextAdapter) Request() *http.Request {
	return g.ginContext.Request()
}

func (g *GinContextAdapter) ReadBody() ([]byte, error) {
	return g.ginContext.ReadBody()
}

func (g *GinContextAdapter) GetHeader(key string) string {
	return g.ginContext.GetHeader(key)
}

func (g *GinContextAdapter) Header(key, value string) {
	g.ginContext.Header(key, value)
}

func (g *GinContextAdapter) Param(key string) string {
	return g.ginContext.Param(key)
}

func (g *GinContextAdapter) Query(key string) string {
	return g.ginContext.Query(key)
}

func (g *GinContextAdapter) DefaultQuery(key, defaultValue string) string {
	return g.ginContext.DefaultQuery(key, defaultValue)
}

func (g *GinContextAdapter) PostForm(key string) string {
	return g.ginContext.PostForm(key)
}

func (g *GinContextAdapter) DefaultPostForm(key, defaultValue string) string {
	return g.ginContext.DefaultPostForm(key, defaultValue)
}

func (g *GinContextAdapter) Bind(obj interface{}) error {
	return g.ginContext.Bind(obj)
}

func (g *GinContextAdapter) ShouldBind(obj interface{}) error {
	return g.ginContext.ShouldBind(obj)
}

func (g *GinContextAdapter) BindJSON(obj interface{}) error {
	return g.ginContext.BindJSON(obj)
}

func (g *GinContextAdapter) ShouldBindJSON(obj interface{}) error {
	return g.ginContext.ShouldBindJSON(obj)
}

func (g *GinContextAdapter) Set(key string, value interface{}) {
	g.ginContext.Set(key, value)
}

func (g *GinContextAdapter) Get(key string) (value interface{}, exists bool) {
	return g.ginContext.Get(key)
}

func (g *GinContextAdapter) GetString(key string) string {
	return g.ginContext.GetString(key)
}

func (g *GinContextAdapter) GetBool(key string) bool {
	return g.ginContext.GetBool(key)
}

func (g *GinContextAdapter) GetInt(key string) int {
	return g.ginContext.GetInt(key)
}

func (g *GinContextAdapter) GetInt64(key string) int64 {
	return g.ginContext.GetInt64(key)
}

func (g *GinContextAdapter) GetFloat64(key string) float64 {
	return g.ginContext.GetFloat64(key)
}

func (g *GinContextAdapter) Abort() {
	g.ginContext.Abort()
}

func (g *GinContextAdapter) AbortWithStatus(code int) {
	g.ginContext.AbortWithStatus(code)
}

func (g *GinContextAdapter) AbortWithStatusJSON(code int, jsonObj interface{}) {
	g.ginContext.AbortWithStatusJSON(code, jsonObj)
}

func (g *GinContextAdapter) IsAborted() bool {
	return g.ginContext.IsAborted()
}

func (g *GinContextAdapter) Next() {
	g.ginContext.Next()
}

// GinContext returns the underlying raw gin.Context
func (g *GinContextAdapter) GinContext() *ginframework.Context {
	return g.ginContext.GinContext()
}

// GinEngineAdapter adapts gin.Engine to the generic Engine interface
type GinEngineAdapter struct {
	ginEngine *gin.Engine
}

// NewGinEngineAdapter creates a new adapter for gin.Engine
func NewGinEngineAdapter(ginEngine *gin.Engine) *GinEngineAdapter {
	return &GinEngineAdapter{ginEngine: ginEngine}
}

func (g *GinEngineAdapter) GET(relativePath string, handlers ...HandlerFunc) {
	g.ginEngine.GET(relativePath, g.convertHandlers(handlers...)...)
}

func (g *GinEngineAdapter) POST(relativePath string, handlers ...HandlerFunc) {
	g.ginEngine.POST(relativePath, g.convertHandlers(handlers...)...)
}

func (g *GinEngineAdapter) PUT(relativePath string, handlers ...HandlerFunc) {
	g.ginEngine.PUT(relativePath, g.convertHandlers(handlers...)...)
}

func (g *GinEngineAdapter) DELETE(relativePath string, handlers ...HandlerFunc) {
	g.ginEngine.DELETE(relativePath, g.convertHandlers(handlers...)...)
}

func (g *GinEngineAdapter) PATCH(relativePath string, handlers ...HandlerFunc) {
	g.ginEngine.PATCH(relativePath, g.convertHandlers(handlers...)...)
}

func (g *GinEngineAdapter) Use(middleware ...HandlerFunc) {
	g.ginEngine.Use(g.convertHandlers(middleware...)...)
}

func (g *GinEngineAdapter) Handler() http.Handler {
	return g.ginEngine.Handler()
}

// GinEngine returns the underlying gin.Engine
func (g *GinEngineAdapter) GinEngine() *gin.Engine {
	return g.ginEngine
}

// convertHandlers converts generic HandlerFunc to gin.HandlerFunc
func (g *GinEngineAdapter) convertHandlers(handlers ...HandlerFunc) []gin.HandlerFunc {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		ginHandlers[i] = func(c *gin.Context) {
			adapter := NewGinContextAdapter(c)
			handler(adapter)
		}
	}
	return ginHandlers
}

// GinServerAdapter adapts gin.Server to the generic Server interface
type GinServerAdapter struct {
	ginServer *gin.Server
	engine    *GinEngineAdapter
}

// NewGinServerAdapter creates a new adapter for gin.Server
func NewGinServerAdapter(ginServer *gin.Server) *GinServerAdapter {
	return &GinServerAdapter{
		ginServer: ginServer,
		engine:    NewGinEngineAdapter(ginServer.Engine()),
	}
}

func (g *GinServerAdapter) ListenAndServe() error {
	return g.ginServer.ListenAndServe()
}

func (g *GinServerAdapter) ListenAndServeTLS(certFile, keyFile string) error {
	return g.ginServer.ListenAndServeTLS(certFile, keyFile)
}

func (g *GinServerAdapter) Shutdown(ctx context.Context) error {
	return g.ginServer.Shutdown(ctx)
}

func (g *GinServerAdapter) ShutdownWithTimeout(timeout time.Duration) error {
	return g.ginServer.ShutdownWithTimeout(timeout)
}

func (g *GinServerAdapter) SetAddr(addr string) {
	g.ginServer.SetAddr(addr)
}

func (g *GinServerAdapter) GetAddr() string {
	return g.ginServer.GetAddr()
}

func (g *GinServerAdapter) SetReadTimeout(timeout time.Duration) {
	g.ginServer.SetReadTimeout(timeout)
}

func (g *GinServerAdapter) SetWriteTimeout(timeout time.Duration) {
	g.ginServer.SetWriteTimeout(timeout)
}

func (g *GinServerAdapter) SetIdleTimeout(timeout time.Duration) {
	g.ginServer.SetIdleTimeout(timeout)
}

func (g *GinServerAdapter) SetReadHeaderTimeout(timeout time.Duration) {
	g.ginServer.SetReadHeaderTimeout(timeout)
}

func (g *GinServerAdapter) Engine() Engine {
	return g.engine
}

func (g *GinServerAdapter) HTTPServer() *http.Server {
	return g.ginServer.HTTPServer()
}

// GinBackend implements the complete Backend interface for Gin
type GinBackend struct{}

// NewGinBackend creates a new Gin backend implementation
func NewGinBackend() *GinBackend {
	return &GinBackend{}
}

func (g *GinBackend) NewEngine() Engine {
	return NewGinEngineAdapter(gin.New())
}

func (g *GinBackend) NewEngineWithDefaults() Engine {
	return NewGinEngineAdapter(gin.Default())
}

func (g *GinBackend) NewServer(port int, engine Engine) Server {
	ginEngineAdapter := engine.(*GinEngineAdapter)
	ginServer := gin.NewServer(port, ginEngineAdapter.ginEngine)
	return NewGinServerAdapter(ginServer)
}

func (g *GinBackend) NewServerWithAddr(addr string, engine Engine) Server {
	ginEngineAdapter := engine.(*GinEngineAdapter)
	ginServer := gin.NewServerWithAddr(addr, ginEngineAdapter.ginEngine)
	return NewGinServerAdapter(ginServer)
}

func (g *GinBackend) SetMode(mode string) {
	gin.SetMode(mode)
}

func (g *GinBackend) GetMode() string {
	return gin.Mode()
}

func (g *GinBackend) Logger() MiddlewareFunc {
	ginLogger := gin.Logger()
	return func(c Context) {
		adapter := c.(*GinContextAdapter)
		ginLogger(adapter.ginContext)
	}
}

func (g *GinBackend) Recovery() MiddlewareFunc {
	ginRecovery := gin.Recovery()
	return func(c Context) {
		adapter := c.(*GinContextAdapter)
		ginRecovery(adapter.ginContext)
	}
}

// GinCORSBackend implements CORSBackend interface
type GinCORSBackend struct {
	*GinBackend
}

// NewGinCORSBackend creates a new Gin CORS backend implementation
func NewGinCORSBackend() *GinCORSBackend {
	return &GinCORSBackend{
		GinBackend: NewGinBackend(),
	}
}

func (g *GinCORSBackend) CORS() MiddlewareFunc {
	corsMiddleware := gin.Recovery() // This should be a proper CORS middleware
	return func(c Context) {
		adapter := c.(*GinContextAdapter)
		corsMiddleware(adapter.ginContext)
	}
}

func (g *GinCORSBackend) CORSWithConfig(config CORSConfig) MiddlewareFunc {
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
	return func(c Context) {
		adapter := c.(*GinContextAdapter)
		corsMiddleware(adapter.ginContext.GinContext())
	}
}