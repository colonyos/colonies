package backends

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBackendFactory(t *testing.T) {
	factory := NewBackendFactory()

	// Test creating a Gin backend
	backend := factory.CreateBackend(GinBackendType)
	if backend == nil {
		t.Fatal("Failed to create Gin backend")
	}

	// Test creating CORS backend
	corsBackend := factory.CreateCORSBackend(GinBackendType)
	if corsBackend == nil {
		t.Fatal("Failed to create Gin CORS backend")
	}

	// Test available backends
	backends := factory.GetAvailableBackends()
	if len(backends) == 0 {
		t.Fatal("No backends available")
	}

	found := false
	for _, b := range backends {
		if b == GinBackendType {
			found = true
			break
		}
	}
	if !found {
		t.Error("Gin backend not found in available backends")
	}
}

func TestGinBackendInterface(t *testing.T) {
	factory := NewBackendFactory()
	backend := factory.CreateBackend(GinBackendType)

	// Test engine creation
	engine := backend.NewEngineWithDefaults()
	if engine == nil {
		t.Fatal("Failed to create engine")
	}

	// Test route registration
	engine.GET("/test", func(c Context) {
		c.String(http.StatusOK, "test response")
	})

	// Test server creation
	server := backend.NewServer(8080, engine)
	if server == nil {
		t.Fatal("Failed to create server")
	}

	// Test server address
	expectedAddr := ":8080"
	if server.GetAddr() != expectedAddr {
		t.Errorf("Expected address %q, got %q", expectedAddr, server.GetAddr())
	}
}

func TestGinEngineAdapter(t *testing.T) {
	factory := NewBackendFactory()
	backend := factory.CreateBackend(GinBackendType)
	engine := backend.NewEngineWithDefaults()

	// Test adding routes
	engine.GET("/get", func(c Context) {
		c.String(http.StatusOK, "get response")
	})

	engine.POST("/post", func(c Context) {
		c.JSON(http.StatusOK, map[string]string{"method": "POST"})
	})

	engine.PUT("/put", func(c Context) {
		c.String(http.StatusOK, "put response")
	})

	engine.DELETE("/delete", func(c Context) {
		c.String(http.StatusOK, "delete response")
	})

	engine.PATCH("/patch", func(c Context) {
		c.String(http.StatusOK, "patch response")
	})

	// Test GET request first, then add middleware for next test
	req, err := http.NewRequest("GET", "/get", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "get response" {
		t.Errorf("Expected body %q, got %q", "get response", w.Body.String())
	}

	// Test middleware on a new engine with middleware applied before routes
	middlewareEngine := backend.NewEngineWithDefaults()
	middlewareEngine.Use(func(c Context) {
		c.Header("X-Test-Middleware", "active")
		c.Next()
	})
	middlewareEngine.GET("/middleware", func(c Context) {
		c.String(http.StatusOK, "middleware test")
	})

	req2, err := http.NewRequest("GET", "/middleware", nil)
	if err != nil {
		t.Fatalf("Failed to create middleware request: %v", err)
	}

	w2 := httptest.NewRecorder()
	middlewareEngine.Handler().ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w2.Code)
	}

	// Check middleware header
	if w2.Header().Get("X-Test-Middleware") != "active" {
		t.Error("Middleware header not set")
	}
}

func TestGinContextAdapter(t *testing.T) {
	factory := NewBackendFactory()
	backend := factory.CreateBackend(GinBackendType)
	engine := backend.NewEngineWithDefaults()

	// Test JSON binding
	engine.POST("/json", func(c Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	// Test query parameters
	engine.GET("/query", func(c Context) {
		name := c.Query("name")
		age := c.DefaultQuery("age", "unknown")
		c.JSON(http.StatusOK, map[string]string{
			"name": name,
			"age":  age,
		})
	})

	// Test URL parameters
	engine.GET("/user/:id", func(c Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, map[string]string{"id": id})
	})

	// Test context storage
	engine.GET("/storage", func(c Context) {
		c.Set("test_key", "test_value")
		value := c.GetString("test_key")
		c.String(http.StatusOK, value)
	})

	// Test JSON request
	jsonData := `{"message": "hello"}`
	req, err := http.NewRequest("POST", "/json", bytes.NewBufferString(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Test query parameters
	req, err = http.NewRequest("GET", "/query?name=test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w = httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Test context storage
	req, err = http.NewRequest("GET", "/storage", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w = httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "test_value" {
		t.Errorf("Expected body %q, got %q", "test_value", w.Body.String())
	}
}

func TestBackendMethods(t *testing.T) {
	factory := NewBackendFactory()
	backend := factory.CreateBackend(GinBackendType)

	// Test mode setting (this affects the underlying Gin framework)
	originalMode := backend.GetMode()
	backend.SetMode("release")
	
	// Note: In real usage, you'd typically set mode before creating engines
	// Here we're just testing that the methods work
	
	// Test middleware creation
	logger := backend.Logger()
	if logger == nil {
		t.Error("Logger middleware is nil")
	}

	recovery := backend.Recovery()
	if recovery == nil {
		t.Error("Recovery middleware is nil")
	}

	// Restore original mode
	backend.SetMode(originalMode)
}