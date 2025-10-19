package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	ginframework "github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewBackend(t *testing.T) {
	backend := NewBackend()
	assert.NotNil(t, backend)

	// Verify it implements backends.Backend interface
	var _ backends.Backend = backend
}

func TestBackendSetGetMode(t *testing.T) {
	backend := NewBackend()

	// Test setting different modes
	backend.SetMode(ginframework.DebugMode)
	assert.Equal(t, ginframework.DebugMode, backend.GetMode())

	backend.SetMode(ginframework.ReleaseMode)
	assert.Equal(t, ginframework.ReleaseMode, backend.GetMode())

	backend.SetMode(ginframework.TestMode)
	assert.Equal(t, ginframework.TestMode, backend.GetMode())
}

func TestBackendNewEngine(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()
	assert.NotNil(t, engine)

	// Verify it implements backends.Engine interface
	var _ backends.Engine = engine
}

func TestBackendNewEngineWithDefaults(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngineWithDefaults()
	assert.NotNil(t, engine)

	// Verify it implements backends.Engine interface
	var _ backends.Engine = engine
}

func TestBackendNewServer(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()
	server := backend.NewServer(8080, engine)
	assert.NotNil(t, server)

	// Verify it implements backends.Server interface
	var _ backends.Server = server

	// Check default address
	assert.Equal(t, ":8080", server.GetAddr())
}

func TestBackendNewServerWithAddr(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()
	server := backend.NewServerWithAddr("localhost:9090", engine)
	assert.NotNil(t, server)

	// Verify it implements backends.Server interface
	var _ backends.Server = server

	// Check address
	assert.Equal(t, "localhost:9090", server.GetAddr())
}

func TestBackendLogger(t *testing.T) {
	backend := NewBackend()
	logger := backend.Logger()
	assert.NotNil(t, logger)
}

func TestBackendRecovery(t *testing.T) {
	backend := NewBackend()
	recovery := backend.Recovery()
	assert.NotNil(t, recovery)
}

func TestNewCORSBackend(t *testing.T) {
	corsBackend := NewCORSBackend()
	assert.NotNil(t, corsBackend)

	// Verify it implements backends.CORSBackend interface
	var _ backends.CORSBackend = corsBackend
}

func TestCORSBackendCORS(t *testing.T) {
	corsBackend := NewCORSBackend()
	corsMiddleware := corsBackend.CORS()
	assert.NotNil(t, corsMiddleware)
}

func TestCORSBackendCORSWithConfig(t *testing.T) {
	corsBackend := NewCORSBackend()

	config := backends.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	corsMiddleware := corsBackend.CORSWithConfig(config)
	assert.NotNil(t, corsMiddleware)
}

func TestContextAdapterString(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	adapter.String(http.StatusOK, "test message")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test message", w.Body.String())
}

func TestContextAdapterJSON(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	adapter.JSON(http.StatusOK, map[string]string{"message": "hello"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "hello")
}

func TestContextAdapterXML(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	type Message struct {
		Text string `xml:"text"`
	}

	adapter := NewContextAdapter(c)
	adapter.XML(http.StatusOK, Message{Text: "hello"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "hello")
}

func TestContextAdapterData(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	data := []byte("raw data")
	adapter.Data(http.StatusOK, "application/octet-stream", data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "raw data", w.Body.String())
}

func TestContextAdapterStatus(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	adapter.Status(http.StatusNoContent)

	// Need to trigger response write for status to be set
	c.Writer.WriteHeaderNow()

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestContextAdapterRequest(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	adapter := NewContextAdapter(c)
	assert.Equal(t, req, adapter.Request())
}

func TestContextAdapterGetHeader(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	c.Request = req

	adapter := NewContextAdapter(c)
	assert.Equal(t, "Bearer token123", adapter.GetHeader("Authorization"))
}

func TestContextAdapterHeader(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	adapter.Header("X-Custom-Header", "custom-value")

	assert.Equal(t, "custom-value", w.Header().Get("X-Custom-Header"))
}

func TestContextAdapterParam(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	c.Params = []ginframework.Param{
		{Key: "id", Value: "123"},
	}

	adapter := NewContextAdapter(c)
	assert.Equal(t, "123", adapter.Param("id"))
}

func TestContextAdapterQuery(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test?name=alice", nil)
	c.Request = req

	adapter := NewContextAdapter(c)
	assert.Equal(t, "alice", adapter.Query("name"))
}

func TestContextAdapterDefaultQuery(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test?name=alice", nil)
	c.Request = req

	adapter := NewContextAdapter(c)
	assert.Equal(t, "alice", adapter.DefaultQuery("name", "default"))
	assert.Equal(t, "default", adapter.DefaultQuery("missing", "default"))
}

func TestContextAdapterPostForm(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString("username=alice"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = req

	adapter := NewContextAdapter(c)
	assert.Equal(t, "alice", adapter.PostForm("username"))
}

func TestContextAdapterDefaultPostForm(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString("username=alice"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = req

	adapter := NewContextAdapter(c)
	assert.Equal(t, "alice", adapter.DefaultPostForm("username", "default"))
	assert.Equal(t, "default", adapter.DefaultPostForm("missing", "default"))
}

func TestContextAdapterSetGet(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)

	// Set values
	adapter.Set("key1", "value1")
	adapter.Set("key2", 42)
	adapter.Set("key3", true)

	// Get values
	val, exists := adapter.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", val)

	val, exists = adapter.Get("key2")
	assert.True(t, exists)
	assert.Equal(t, 42, val)

	val, exists = adapter.Get("key3")
	assert.True(t, exists)
	assert.Equal(t, true, val)

	// Get non-existent key
	val, exists = adapter.Get("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, val)
}

func TestContextAdapterGetString(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	adapter.Set("name", "Alice")

	assert.Equal(t, "Alice", adapter.GetString("name"))
	assert.Equal(t, "", adapter.GetString("missing"))
}

func TestContextAdapterGetBool(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	adapter.Set("active", true)
	adapter.Set("inactive", false)

	assert.True(t, adapter.GetBool("active"))
	assert.False(t, adapter.GetBool("inactive"))
	assert.False(t, adapter.GetBool("missing"))
}

func TestContextAdapterGetInt(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	adapter.Set("count", 42)

	assert.Equal(t, 42, adapter.GetInt("count"))
	assert.Equal(t, 0, adapter.GetInt("missing"))
}

func TestContextAdapterGetInt64(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	adapter.Set("bignum", int64(1234567890))

	assert.Equal(t, int64(1234567890), adapter.GetInt64("bignum"))
	assert.Equal(t, int64(0), adapter.GetInt64("missing"))
}

func TestContextAdapterGetFloat64(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	adapter.Set("price", 19.99)

	assert.Equal(t, 19.99, adapter.GetFloat64("price"))
	assert.Equal(t, 0.0, adapter.GetFloat64("missing"))
}

func TestContextAdapterAbort(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)

	assert.False(t, adapter.IsAborted())
	adapter.Abort()
	assert.True(t, adapter.IsAborted())
}

func TestContextAdapterAbortWithStatus(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)

	assert.False(t, adapter.IsAborted())
	adapter.AbortWithStatus(http.StatusUnauthorized)
	assert.True(t, adapter.IsAborted())
}

func TestContextAdapterAbortWithStatusJSON(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)

	assert.False(t, adapter.IsAborted())
	adapter.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	assert.True(t, adapter.IsAborted())
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContextAdapterBindJSON(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	jsonData := `{"name": "Alice", "age": 30}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	adapter := NewContextAdapter(c)

	var result map[string]interface{}
	err := adapter.BindJSON(&result)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, float64(30), result["age"])
}

func TestContextAdapterShouldBindJSON(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	jsonData := `{"name": "Bob", "age": 25}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	adapter := NewContextAdapter(c)

	var result map[string]interface{}
	err := adapter.ShouldBindJSON(&result)
	assert.NoError(t, err)
	assert.Equal(t, "Bob", result["name"])
}

func TestContextAdapterNext(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)

	// Next should not panic
	assert.NotPanics(t, func() {
		adapter.Next()
	})
}

func TestContextAdapterGinContext(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)

	// Cast to ContextAdapter to access GinContext method
	contextAdapter, ok := adapter.(*ContextAdapter)
	assert.True(t, ok)

	// Should return the underlying gin context
	assert.Equal(t, c, contextAdapter.GinContext())
}

func TestEngineAdapterHTTPMethods(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()

	// Test all HTTP methods
	engine.GET("/get", func(c backends.Context) {
		c.String(http.StatusOK, "GET response")
	})

	engine.POST("/post", func(c backends.Context) {
		c.String(http.StatusOK, "POST response")
	})

	engine.PUT("/put", func(c backends.Context) {
		c.String(http.StatusOK, "PUT response")
	})

	engine.DELETE("/delete", func(c backends.Context) {
		c.String(http.StatusOK, "DELETE response")
	})

	engine.PATCH("/patch", func(c backends.Context) {
		c.String(http.StatusOK, "PATCH response")
	})

	// Test GET
	req, _ := http.NewRequest("GET", "/get", nil)
	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "GET response", w.Body.String())

	// Test POST
	req, _ = http.NewRequest("POST", "/post", nil)
	w = httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "POST response", w.Body.String())

	// Test PUT
	req, _ = http.NewRequest("PUT", "/put", nil)
	w = httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "PUT response", w.Body.String())

	// Test DELETE
	req, _ = http.NewRequest("DELETE", "/delete", nil)
	w = httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "DELETE response", w.Body.String())

	// Test PATCH
	req, _ = http.NewRequest("PATCH", "/patch", nil)
	w = httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "PATCH response", w.Body.String())
}

func TestEngineAdapterUseMiddleware(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()

	// Add middleware that sets a custom header
	engine.Use(func(c backends.Context) {
		c.Header("X-Middleware", "applied")
		c.Next()
	})

	engine.GET("/test", func(c backends.Context) {
		c.String(http.StatusOK, "test")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "applied", w.Header().Get("X-Middleware"))
}

func TestServerAdapterSetGetAddr(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()
	server := backend.NewServer(8080, engine)

	assert.Equal(t, ":8080", server.GetAddr())

	server.SetAddr(":9090")
	assert.Equal(t, ":9090", server.GetAddr())
}

func TestServerAdapterTimeouts(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()
	server := backend.NewServer(8080, engine)

	// Set various timeouts
	assert.NotPanics(t, func() {
		server.SetReadTimeout(10 * time.Second)
		server.SetWriteTimeout(10 * time.Second)
		server.SetIdleTimeout(60 * time.Second)
		server.SetReadHeaderTimeout(5 * time.Second)
	})
}

func TestServerAdapterEngine(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()
	server := backend.NewServer(8080, engine)

	// Should return the same engine
	assert.NotNil(t, server.Engine())
}

func TestServerAdapterHTTPServer(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()
	server := backend.NewServer(8080, engine)

	httpServer := server.HTTPServer()
	assert.NotNil(t, httpServer)
	assert.Equal(t, ":8080", httpServer.Addr)
}

func TestServerAdapterShutdown(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()
	server := backend.NewServer(0, engine) // Use port 0 for dynamic allocation

	// Start server in background
	go func() {
		server.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown with context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestServerAdapterShutdownWithTimeout(t *testing.T) {
	backend := NewBackend()
	engine := backend.NewEngine()
	server := backend.NewServer(0, engine)

	// Start server in background
	go func() {
		server.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown with timeout
	err := server.ShutdownWithTimeout(5 * time.Second)
	assert.NoError(t, err)
}

func TestContextAdapterReadBody(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	bodyContent := "request body content"
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(bodyContent))
	c.Request = req

	adapter := NewContextAdapter(c)
	body, err := adapter.ReadBody()
	assert.NoError(t, err)
	assert.Equal(t, bodyContent, string(body))
}

func TestFullBackendWorkflow(t *testing.T) {
	// Create a complete workflow test
	backend := NewBackend()
	backend.SetMode(ginframework.TestMode)

	// Create engine with middleware
	engine := backend.NewEngineWithDefaults()

	// Add custom middleware
	engine.Use(func(c backends.Context) {
		c.Header("X-Test", "workflow")
		c.Next()
	})

	// Add routes
	engine.POST("/api/data", func(c backends.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	// Create server
	server := backend.NewServerWithAddr(":8888", engine)
	assert.Equal(t, ":8888", server.GetAddr())

	// Test request
	jsonData := map[string]interface{}{
		"message": "hello",
		"count":   42,
	}
	jsonBytes, _ := json.Marshal(jsonData)

	req, _ := http.NewRequest("POST", "/api/data", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "workflow", w.Header().Get("X-Test"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "hello", response["message"])
	assert.Equal(t, float64(42), response["count"])
}
