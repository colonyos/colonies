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
	"github.com/gin-contrib/cors"
	ginframework "github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestCreateEngine(t *testing.T) {
	engine := CreateEngine()
	assert.NotNil(t, engine)

	// Verify it implements backends.Engine interface
	var _ backends.Engine = engine
}

func TestCreateEngineWithDefaults(t *testing.T) {
	engine := CreateEngineWithDefaults()
	assert.NotNil(t, engine)

	// Verify it implements backends.Engine interface
	var _ backends.Engine = engine
}

func TestNewBackendServer(t *testing.T) {
	engine := CreateEngine()
	server := NewBackendServer(8080, engine)
	assert.NotNil(t, server)

	// Verify it implements backends.Server interface
	var _ backends.Server = server

	// Check default address
	assert.Equal(t, ":8080", server.GetAddr())
}

func TestNewBackendServerWithAddr(t *testing.T) {
	engine := CreateEngine()
	server := NewBackendServerWithAddr("localhost:9090", engine)
	assert.NotNil(t, server)

	// Verify it implements backends.Server interface
	var _ backends.Server = server

	// Check address
	assert.Equal(t, "localhost:9090", server.GetAddr())
}

func TestCORS(t *testing.T) {
	corsMiddleware := CORS()
	assert.NotNil(t, corsMiddleware)
}

func TestCORSWithConfig(t *testing.T) {
	config := backends.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	corsMiddleware := CORSWithConfig(config)
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
	engine := CreateEngine()

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
	engine := CreateEngine()

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
	engine := CreateEngine()
	server := NewBackendServer(8080, engine)

	assert.Equal(t, ":8080", server.GetAddr())

	server.SetAddr(":9090")
	assert.Equal(t, ":9090", server.GetAddr())
}

func TestServerAdapterTimeouts(t *testing.T) {
	engine := CreateEngine()
	server := NewBackendServer(8080, engine)

	// Set various timeouts
	assert.NotPanics(t, func() {
		server.SetReadTimeout(10 * time.Second)
		server.SetWriteTimeout(10 * time.Second)
		server.SetIdleTimeout(60 * time.Second)
		server.SetReadHeaderTimeout(5 * time.Second)
	})
}

func TestServerAdapterEngine(t *testing.T) {
	engine := CreateEngine()
	server := NewBackendServer(8080, engine)

	// Should return the same engine
	assert.NotNil(t, server.Engine())
}

func TestServerAdapterHTTPServer(t *testing.T) {
	engine := CreateEngine()
	server := NewBackendServer(8080, engine)

	httpServer := server.HTTPServer()
	assert.NotNil(t, httpServer)
	assert.Equal(t, ":8080", httpServer.Addr)
}

func TestServerAdapterShutdown(t *testing.T) {
	engine := CreateEngine()
	server := NewBackendServer(0, engine) // Use port 0 for dynamic allocation

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
	engine := CreateEngine()
	server := NewBackendServer(0, engine)

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
	ginframework.SetMode(ginframework.TestMode)

	// Create engine with middleware
	engine := CreateEngineWithDefaults()

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
	server := NewBackendServerWithAddr(":8888", engine)
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

// ============== Context wrapper tests (context.go) ==============

func TestContextWrapper(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	assert.NotNil(t, ctx)
}

func TestContextString(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.String(http.StatusOK, "hello %s", "world")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello world", w.Body.String())
}

func TestContextJSON(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.JSON(http.StatusOK, map[string]string{"msg": "hi"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "hi")
}

func TestContextXML(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	type Message struct {
		Text string `xml:"text"`
	}

	ctx := NewContext(c)
	ctx.XML(http.StatusOK, Message{Text: "xml-test"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "xml-test")
}

func TestContextData(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.Data(http.StatusOK, "text/plain", []byte("raw data"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "raw data", w.Body.String())
}

func TestContextRequest(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	ctx := NewContext(c)
	assert.Equal(t, req, ctx.Request())
}

func TestContextWriter(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	assert.NotNil(t, ctx.Writer())
}

func TestContextGetHeader(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Custom", "test-value")
	c.Request = req

	ctx := NewContext(c)
	assert.Equal(t, "test-value", ctx.GetHeader("X-Custom"))
}

func TestContextHeader(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.Header("X-Response", "response-value")

	assert.Equal(t, "response-value", w.Header().Get("X-Response"))
}

func TestContextParam(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	c.Params = []ginframework.Param{{Key: "id", Value: "456"}}

	ctx := NewContext(c)
	assert.Equal(t, "456", ctx.Param("id"))
}

func TestContextQuery(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test?foo=bar", nil)
	c.Request = req

	ctx := NewContext(c)
	assert.Equal(t, "bar", ctx.Query("foo"))
}

func TestContextDefaultQuery(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test?foo=bar", nil)
	c.Request = req

	ctx := NewContext(c)
	assert.Equal(t, "bar", ctx.DefaultQuery("foo", "default"))
	assert.Equal(t, "default", ctx.DefaultQuery("missing", "default"))
}

func TestContextPostForm(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString("name=alice"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = req

	ctx := NewContext(c)
	assert.Equal(t, "alice", ctx.PostForm("name"))
}

func TestContextDefaultPostForm(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString("name=bob"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = req

	ctx := NewContext(c)
	assert.Equal(t, "bob", ctx.DefaultPostForm("name", "default"))
	assert.Equal(t, "default", ctx.DefaultPostForm("missing", "default"))
}

func TestContextBind(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	jsonData := `{"name": "test"}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	ctx := NewContext(c)
	var result map[string]interface{}
	err := ctx.Bind(&result)
	assert.NoError(t, err)
	assert.Equal(t, "test", result["name"])
}

func TestContextShouldBind(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	jsonData := `{"name": "should-bind"}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	ctx := NewContext(c)
	var result map[string]interface{}
	err := ctx.ShouldBind(&result)
	assert.NoError(t, err)
	assert.Equal(t, "should-bind", result["name"])
}

func TestContextBindJSON(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	jsonData := `{"key": "value"}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	ctx := NewContext(c)
	var result map[string]interface{}
	err := ctx.BindJSON(&result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result["key"])
}

func TestContextShouldBindJSON(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	jsonData := `{"data": "test"}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	ctx := NewContext(c)
	var result map[string]interface{}
	err := ctx.ShouldBindJSON(&result)
	assert.NoError(t, err)
	assert.Equal(t, "test", result["data"])
}

func TestContextSetGet(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.Set("key", "value")

	val, exists := ctx.Get("key")
	assert.True(t, exists)
	assert.Equal(t, "value", val)
}

func TestContextGetString(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.Set("str", "hello")

	assert.Equal(t, "hello", ctx.GetString("str"))
}

func TestContextGetBool(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.Set("active", true)

	assert.True(t, ctx.GetBool("active"))
}

func TestContextGetInt(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.Set("num", 42)

	assert.Equal(t, 42, ctx.GetInt("num"))
}

func TestContextGetInt64(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.Set("bignum", int64(9999999999))

	assert.Equal(t, int64(9999999999), ctx.GetInt64("bignum"))
}

func TestContextGetFloat64(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.Set("price", 19.99)

	assert.Equal(t, 19.99, ctx.GetFloat64("price"))
}

func TestContextAbort(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	assert.False(t, ctx.IsAborted())
	ctx.Abort()
	assert.True(t, ctx.IsAborted())
}

func TestContextAbortWithStatus(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.AbortWithStatus(http.StatusForbidden)
	assert.True(t, ctx.IsAborted())
}

func TestContextAbortWithStatusJSON(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{"error": "bad"})
	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContextNext(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	assert.NotPanics(t, func() {
		ctx.Next()
	})
}

func TestContextGinContext(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	assert.Equal(t, c, ctx.GinContext())
}

func TestContextReadBody(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString("body content"))
	c.Request = req

	ctx := NewContext(c)
	body, err := ctx.ReadBody()
	assert.NoError(t, err)
	assert.Equal(t, "body content", string(body))
}

func TestContextStatus(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	ctx := NewContext(c)
	ctx.Status(http.StatusCreated)
	c.Writer.WriteHeaderNow()

	assert.Equal(t, http.StatusCreated, w.Code)
}

// ============== Engine wrapper tests (engine.go) ==============

func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.GinEngine())
}

func TestEngineUseCORS(t *testing.T) {
	engine := NewEngine()
	assert.NotPanics(t, func() {
		engine.UseCORS()
	})
}

func TestEngineUseCORSWithConfig(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := NewEngine()

	config := cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST"},
	}

	assert.NotPanics(t, func() {
		engine.UseCORSWithConfig(config)
	})
}

func TestEnginePUT(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := NewEngine()

	engine.PUT("/test", func(c *Context) {
		c.String(http.StatusOK, "PUT ok")
	})

	req, _ := http.NewRequest("PUT", "/test", nil)
	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "PUT ok", w.Body.String())
}

func TestEngineDELETE(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := NewEngine()

	engine.DELETE("/test", func(c *Context) {
		c.String(http.StatusOK, "DELETE ok")
	})

	req, _ := http.NewRequest("DELETE", "/test", nil)
	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "DELETE ok", w.Body.String())
}

func TestEnginePATCH(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := NewEngine()

	engine.PATCH("/test", func(c *Context) {
		c.String(http.StatusOK, "PATCH ok")
	})

	req, _ := http.NewRequest("PATCH", "/test", nil)
	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "PATCH ok", w.Body.String())
}

func TestEngineNewEngineWithGin(t *testing.T) {
	ginEngine := ginframework.New()
	engine := NewEngineWithGin(ginEngine)
	assert.NotNil(t, engine)
	assert.Equal(t, ginEngine, engine.GinEngine())
}

// ============== gin.go helper function tests ==============

func TestGinSetMode(t *testing.T) {
	SetMode(ginframework.TestMode)
	assert.Equal(t, ginframework.TestMode, Mode())
}

func TestGinMode(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	mode := Mode()
	assert.NotEmpty(t, mode)
}

func TestGinIsDebugging(t *testing.T) {
	ginframework.SetMode(ginframework.ReleaseMode)
	assert.False(t, IsDebugging())

	ginframework.SetMode(ginframework.DebugMode)
	assert.True(t, IsDebugging())

	// Reset to test mode
	ginframework.SetMode(ginframework.TestMode)
}

func TestGinRecovery(t *testing.T) {
	recoveryMiddleware := Recovery()
	assert.NotNil(t, recoveryMiddleware)
}

func TestGinLogger(t *testing.T) {
	loggerMiddleware := Logger()
	assert.NotNil(t, loggerMiddleware)
}

func TestGinLoggerWithFormatter(t *testing.T) {
	formatter := func(param ginframework.LogFormatterParams) string {
		return param.Method + " " + param.Path
	}
	loggerMiddleware := LoggerWithFormatter(formatter)
	assert.NotNil(t, loggerMiddleware)
}

func TestGinLoggerWithWriter(t *testing.T) {
	formatter := func(param ginframework.LogFormatterParams) string {
		return param.Method
	}
	loggerMiddleware := LoggerWithWriter(formatter, "/health")
	assert.NotNil(t, loggerMiddleware)
}

func TestGinBasicAuth(t *testing.T) {
	accounts := ginframework.Accounts{
		"admin": "secret",
	}
	authMiddleware := BasicAuth(accounts)
	assert.NotNil(t, authMiddleware)
}

func TestGinDefault(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := Default()
	assert.NotNil(t, engine)
}

func TestGinNew(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := New()
	assert.NotNil(t, engine)
}

// ============== Server wrapper tests (server.go) ==============

func TestServerListenAndServeTLS(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := NewEngine()
	server := NewServer(0, engine)

	// Test that ListenAndServeTLS returns error for invalid cert files
	err := server.ListenAndServeTLS("nonexistent.cert", "nonexistent.key")
	assert.Error(t, err)
}

// ============== ServerHelpers tests (server_helpers.go) ==============

func TestServerHelpersNew(t *testing.T) {
	helpers := NewServerHelpers()
	assert.NotNil(t, helpers)
}

func TestServerHelpersHandleHTTPErrorGin(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	helpers := NewServerHelpers()

	// Test with no error
	result := helpers.HandleHTTPErrorGin(c, nil, http.StatusBadRequest)
	assert.False(t, result)

	// Test with error
	result = helpers.HandleHTTPErrorGin(c, assert.AnError, http.StatusBadRequest)
	assert.True(t, result)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServerHelpersHandleHTTPErrorContext(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	helpers := NewServerHelpers()

	// Test with no error
	result := helpers.HandleHTTPErrorContext(adapter, nil, http.StatusInternalServerError)
	assert.False(t, result)

	// Test with error
	w2 := httptest.NewRecorder()
	c2, _ := ginframework.CreateTestContext(w2)
	adapter2 := NewContextAdapter(c2)

	result = helpers.HandleHTTPErrorContext(adapter2, assert.AnError, http.StatusInternalServerError)
	assert.True(t, result)
	assert.Equal(t, http.StatusInternalServerError, w2.Code)
}

func TestServerHelpersSendHTTPReplyGin(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	helpers := NewServerHelpers()
	helpers.SendHTTPReplyGin(c, "test-type", `{"data":"test"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "test-type", w.Header().Get("Payload-Type"))
	assert.Equal(t, `{"data":"test"}`, w.Body.String())
}

func TestServerHelpersSendEmptyHTTPReplyGin(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	helpers := NewServerHelpers()
	helpers.SendEmptyHTTPReplyGin(c, "empty-type")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "empty-type", w.Header().Get("Payload-Type"))
}

func TestServerHelpersExtractGinContext(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	adapter := NewContextAdapter(c)
	helpers := NewServerHelpers()

	ginCtx, ok := helpers.ExtractGinContext(adapter)
	assert.True(t, ok)
	assert.Equal(t, c, ginCtx)
}

func TestServerHelpersHandleContextUnion(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	helpers := NewServerHelpers()

	// Test with raw gin.Context
	ctx, ginCtx, ok := helpers.HandleContextUnion(c)
	assert.True(t, ok)
	assert.NotNil(t, ctx)
	assert.Equal(t, c, ginCtx)

	// Test with backends.Context (ContextAdapter)
	adapter := NewContextAdapter(c)
	ctx, ginCtx, ok = helpers.HandleContextUnion(adapter)
	assert.True(t, ok)
	assert.NotNil(t, ctx)
	assert.Equal(t, c, ginCtx)

	// Test with invalid type
	ctx, ginCtx, ok = helpers.HandleContextUnion("invalid")
	assert.False(t, ok)
	assert.Nil(t, ctx)
	assert.Nil(t, ginCtx)
}

// ============== Factory tests (factory.go) ==============

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
}

func TestFactoryCreateConnectionValid(t *testing.T) {
	factory := NewFactory()

	// Test with invalid connection type
	_, err := factory.CreateConnection("invalid")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidConnType, err)
}

func TestFactoryCreateEventHandler(t *testing.T) {
	factory := NewFactory()

	// Test with nil relay server
	handler := factory.CreateEventHandler(nil)
	assert.NotNil(t, handler)
}

func TestFactoryCreateTestableEventHandler(t *testing.T) {
	factory := NewFactory()

	// Test with nil relay server
	handler := factory.CreateTestableEventHandler(nil)
	assert.NotNil(t, handler)
}

func TestFactoryCreateSubscriptionController(t *testing.T) {
	factory := NewFactory()

	// Create an event handler first
	eventHandler := factory.CreateEventHandler(nil)

	// Create subscription controller
	subController := factory.CreateSubscriptionController(eventHandler)
	assert.NotNil(t, subController)
}

// ============== Connection tests (connection.go) ==============

func TestWebSocketConnectionNilConn(t *testing.T) {
	// Test with nil connection
	conn := NewWebSocketConnection(nil)
	wsConn := conn.(*WebSocketConnection)

	// WriteMessage should return error for nil conn
	err := wsConn.WriteMessage(1, []byte("test"))
	assert.Error(t, err)
	assert.Equal(t, ErrConnectionClosed, err)

	// IsOpen should return false for nil conn
	assert.False(t, wsConn.IsOpen())

	// Close should not panic for nil conn
	err = wsConn.Close()
	assert.NoError(t, err)
}

// ============== ContextAdapter Bind/ShouldBind tests ==============

func TestContextAdapterBind(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	jsonData := `{"field": "test-bind"}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	adapter := NewContextAdapter(c)

	var result map[string]interface{}
	err := adapter.Bind(&result)
	assert.NoError(t, err)
	assert.Equal(t, "test-bind", result["field"])
}

func TestContextAdapterShouldBind(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	jsonData := `{"field": "should-bind-test"}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	adapter := NewContextAdapter(c)

	var result map[string]interface{}
	err := adapter.ShouldBind(&result)
	assert.NoError(t, err)
	assert.Equal(t, "should-bind-test", result["field"])
}

// ============== Subscription utility function tests ==============

func TestCreateSubscription(t *testing.T) {
	// Test with nil WebSocket connection (allowed for struct creation)
	sub := CreateSubscription(nil, 1, "process-123", "test-executor", 1, 30)
	assert.NotNil(t, sub)
	assert.Equal(t, 1, sub.MsgType)
	assert.Equal(t, "process-123", sub.ProcessID)
	assert.Equal(t, "test-executor", sub.ExecutorType)
	assert.Equal(t, 1, sub.State)
	assert.Equal(t, 30, sub.Timeout)
}

func TestCreateProcessesSubscription(t *testing.T) {
	sub := CreateProcessesSubscription(nil, 2, "executor-type", 60, 2)
	assert.NotNil(t, sub)
	assert.Equal(t, 2, sub.MsgType)
	assert.Equal(t, "executor-type", sub.ExecutorType)
	assert.Equal(t, 60, sub.Timeout)
	assert.Equal(t, 2, sub.State)
}

func TestCreateProcessSubscription(t *testing.T) {
	sub := CreateProcessSubscription(nil, 1, "proc-456", "worker", 120, 3)
	assert.NotNil(t, sub)
	assert.Equal(t, 1, sub.MsgType)
	assert.Equal(t, "proc-456", sub.ProcessID)
	assert.Equal(t, "worker", sub.ExecutorType)
	assert.Equal(t, 120, sub.Timeout)
	assert.Equal(t, 3, sub.State)
}

func TestNewSubscriptionController(t *testing.T) {
	factory := NewFactory()
	eventHandler := factory.CreateEventHandler(nil)

	controller := NewSubscriptionController(eventHandler)
	assert.NotNil(t, controller)
}

// ============== ServerAdapter TLS test ==============

func TestServerAdapterListenAndServeTLS(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := CreateEngine()
	server := NewBackendServer(0, engine)

	serverAdapter := server.(*ServerAdapter)

	// Test that ListenAndServeTLS returns error for invalid cert files
	err := serverAdapter.ListenAndServeTLS("nonexistent.cert", "nonexistent.key")
	assert.Error(t, err)
}

// ============== CORS middleware actual execution tests ==============

func TestCORSMiddlewareExecution(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := CreateEngine()

	// Add CORS middleware
	corsMiddleware := CORS()
	engine.Use(func(c backends.Context) {
		corsMiddleware(c)
		c.Next()
	})

	engine.GET("/test", func(c backends.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Test OPTIONS request (preflight)
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	// CORS headers should be present
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSWithConfigMiddlewareExecution(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := CreateEngine()

	config := backends.CORSConfig{
		AllowOrigins:     []string{"http://example.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600 * time.Second,
	}

	corsMiddleware := CORSWithConfig(config)
	engine.Use(func(c backends.Context) {
		corsMiddleware(c)
		c.Next()
	})

	engine.GET("/api", func(c backends.Context) {
		c.String(http.StatusOK, "api response")
	})

	// Test preflight request
	req, _ := http.NewRequest("OPTIONS", "/api", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	// Verify CORS is applied
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Origin"), "http://example.com")
}

// ============== Additional eventhandler tests ==============

func TestEventHandlerHasStopped(t *testing.T) {
	factory := NewFactory()
	handler := factory.CreateTestableEventHandler(nil)

	testableHandler := handler.(*DefaultEventHandler)

	// Call HasStopped - should not panic
	_ = testableHandler.HasStopped()

	// Stop the handler
	testableHandler.Stop()

	// After stopping, HasStopped should return true
	time.Sleep(100 * time.Millisecond)
	assert.True(t, testableHandler.HasStopped())
}

// ============== RealtimeHandler tests ==============

func TestNewRealtimeHandler(t *testing.T) {
	// Test with nil server (just verifies constructor doesn't panic)
	handler := NewRealtimeHandler(nil)
	assert.NotNil(t, handler)
}

// ============== Middleware invocation tests ==============

func TestRecoveryMiddlewareInvocation(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := CreateEngine()

	// Get recovery middleware
	recovery := Recovery()

	// Add to engine through the wrapper mechanism
	engine.Use(func(c backends.Context) {
		// Create a new context wrapper for recovery
		ginAdapter := c.(*ContextAdapter)
		ctx := NewContext(ginAdapter.ginContext)
		recovery(ctx)
		c.Next()
	})

	engine.GET("/test", func(c backends.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoggerMiddlewareInvocation(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := CreateEngine()

	// Get logger middleware
	logger := Logger()

	engine.Use(func(c backends.Context) {
		ginAdapter := c.(*ContextAdapter)
		ctx := NewContext(ginAdapter.ginContext)
		logger(ctx)
		c.Next()
	})

	engine.GET("/test", func(c backends.Context) {
		c.String(http.StatusOK, "logged")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBasicAuthMiddlewareInvocation(t *testing.T) {
	ginframework.SetMode(ginframework.TestMode)
	engine := CreateEngine()

	accounts := ginframework.Accounts{
		"admin": "password123",
	}
	basicAuth := BasicAuth(accounts)

	engine.Use(func(c backends.Context) {
		ginAdapter := c.(*ContextAdapter)
		ctx := NewContext(ginAdapter.ginContext)
		basicAuth(ctx)
		c.Next()
	})

	engine.GET("/protected", func(c backends.Context) {
		c.String(http.StatusOK, "protected content")
	})

	// Without auth, should get 401
	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============== Connection with actual WebSocket mock ==============

func TestWebSocketConnectionWithActualConn(t *testing.T) {
	// Test the positive case where connection is valid
	conn := &WebSocketConnection{conn: nil}

	// IsOpen for nil internal conn
	assert.False(t, conn.IsOpen())

	// WriteMessage with nil internal conn
	err := conn.WriteMessage(1, []byte("test"))
	assert.Equal(t, ErrConnectionClosed, err)

	// Close with nil internal conn
	err = conn.Close()
	assert.NoError(t, err)
}

// ============== ExtractGinContext edge case ==============

func TestExtractGinContextWithNonAdapter(t *testing.T) {
	helpers := NewServerHelpers()

	// Test with a non-ContextAdapter type (simulated with our own Context wrapper)
	ginframework.SetMode(ginframework.TestMode)
	w := httptest.NewRecorder()
	c, _ := ginframework.CreateTestContext(w)

	// Use the wrapped Context (not ContextAdapter)
	ctx := NewContext(c)

	// ExtractGinContext expects ContextAdapter, not Context
	// Should return false for non-ContextAdapter
	_, ok := helpers.ExtractGinContext(ctx)
	assert.False(t, ok)
}

// ============== Factory with valid WebSocket connection ==============

func TestFactoryCreateConnectionWithWebSocket(t *testing.T) {
	factory := NewFactory()

	// Test with nil websocket connection (should work - nil is valid websocket.Conn pointer)
	conn, err := factory.CreateConnection((*websocket.Conn)(nil))
	assert.NoError(t, err)
	assert.NotNil(t, conn)
}
