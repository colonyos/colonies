package gin

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/colonyos/colonies/pkg/service/backend"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GinBackend implements the Backend interface using Gin
type GinBackend struct {
	engine *gin.Engine
	server *http.Server
	name   string
	mutex  sync.RWMutex
}

// GinContext wraps gin.Context to implement our Context interface
type GinContext struct {
	ginCtx *gin.Context
	req    *GinRequest
	resp   *GinResponse
}

// GinRequest wraps gin.Context for request operations
type GinRequest struct {
	ctx *gin.Context
}

// GinResponse wraps gin.Context for response operations
type GinResponse struct {
	ctx *gin.Context
}

// NewGinBackend creates a new Gin backend
func NewGinBackend(config backend.BackendConfig) (backend.Backend, error) {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	
	engine := gin.New()
	
	// Add basic middleware
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	
	return &GinBackend{
		engine: engine,
		name:   config.Name,
	}, nil
}

// Type returns the backend type
func (g *GinBackend) Type() backend.BackendType {
	return backend.HTTPBackend
}

// Name returns the backend name
func (g *GinBackend) Name() string {
	return g.name
}

// Start starts the Gin server
func (g *GinBackend) Start(addr string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.server = &http.Server{
		Addr:    addr,
		Handler: g.engine,
	}
	
	log.WithFields(log.Fields{
		"backend": "gin",
		"address": addr,
	}).Info("Starting Gin backend")
	
	return g.server.ListenAndServe()
}

// Stop stops the Gin server
func (g *GinBackend) Stop() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	if g.server != nil {
		log.WithFields(log.Fields{"backend": "gin"}).Info("Stopping Gin backend")
		return g.server.Shutdown(context.Background())
	}
	
	return nil
}

// Handle registers a handler for a specific HTTP method and path
func (g *GinBackend) Handle(method, path string, handler backend.HandlerFunc) {
	ginHandler := func(c *gin.Context) {
		ctx := &GinContext{
			ginCtx: c,
			req:    &GinRequest{ctx: c},
			resp:   &GinResponse{ctx: c},
		}
		handler(ctx)
	}
	
	g.engine.Handle(method, path, ginHandler)
}

// Use adds middleware to the Gin engine
func (g *GinBackend) Use(middleware backend.MiddlewareFunc) {
	ginMiddleware := func(c *gin.Context) {
		ctx := &GinContext{
			ginCtx: c,
			req:    &GinRequest{ctx: c},
			resp:   &GinResponse{ctx: c},
		}
		
		wrappedHandler := middleware(func(ctx backend.Context) {
			c.Next()
		})
		
		wrappedHandler(ctx)
	}
	
	g.engine.Use(ginMiddleware)
}

// SetCORS configures CORS for the Gin backend
func (g *GinBackend) SetCORS(config backend.CORSConfig) {
	corsConfig := cors.Config{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     config.AllowMethods,
		AllowHeaders:     config.AllowHeaders,
		AllowCredentials: config.AllowCredentials,
	}
	
	g.engine.Use(cors.New(corsConfig))
}

// Context interface implementation
func (gc *GinContext) Request() backend.Request {
	return gc.req
}

func (gc *GinContext) Response() backend.Response {
	return gc.resp
}

func (gc *GinContext) Set(key string, value interface{}) {
	gc.ginCtx.Set(key, value)
}

func (gc *GinContext) Get(key string) (interface{}, bool) {
	return gc.ginCtx.Get(key)
}

func (gc *GinContext) Abort() {
	gc.ginCtx.Abort()
}

func (gc *GinContext) IsAborted() bool {
	return gc.ginCtx.IsAborted()
}

func (gc *GinContext) Context() context.Context {
	return gc.ginCtx.Request.Context()
}

// Request interface implementation
func (gr *GinRequest) GetBody() ([]byte, error) {
	if gr.ctx.Request.Body == nil {
		return []byte{}, nil
	}
	
	body, err := io.ReadAll(gr.ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	
	// Restore the body for potential subsequent reads
	gr.ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	
	return body, nil
}

func (gr *GinRequest) GetHeader(key string) string {
	return gr.ctx.GetHeader(key)
}

func (gr *GinRequest) GetMethod() string {
	return gr.ctx.Request.Method
}

func (gr *GinRequest) GetPath() string {
	return gr.ctx.Request.URL.Path
}

func (gr *GinRequest) GetRemoteAddr() string {
	return gr.ctx.ClientIP()
}

func (gr *GinRequest) GetQuery(key string) string {
	return gr.ctx.Query(key)
}

func (gr *GinRequest) GetParam(key string) string {
	return gr.ctx.Param(key)
}

// Response interface implementation
func (gr *GinResponse) SetStatus(code int) {
	gr.ctx.Status(code)
}

func (gr *GinResponse) SetHeader(key, value string) {
	gr.ctx.Header(key, value)
}

func (gr *GinResponse) Write(data []byte) error {
	_, err := gr.ctx.Writer.Write(data)
	return err
}

func (gr *GinResponse) WriteJSON(data interface{}) error {
	gr.ctx.JSON(http.StatusOK, data)
	return nil
}

func (gr *GinResponse) WriteString(data string) error {
	_, err := gr.ctx.Writer.WriteString(data)
	return err
}

func (gr *GinResponse) GetStatus() int {
	return gr.ctx.Writer.Status()
}

// Factory function for registration
func CreateGinBackend(config backend.BackendConfig) (backend.Backend, error) {
	return NewGinBackend(config)
}