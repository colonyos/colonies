package gin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEngine(t *testing.T) {
	// Create a new engine
	engine := Default()

	// Add a test route
	engine.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "test response")
	})

	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	w := httptest.NewRecorder()

	// Serve the request
	engine.Handler().ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	expected := "test response"
	if w.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, w.Body.String())
	}
}

func TestContext(t *testing.T) {
	engine := Default()

	// Test JSON response
	engine.POST("/json", func(c *Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	// Test with JSON request
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
}

func TestServer(t *testing.T) {
	engine := Default()
	engine.GET("/health", func(c *Context) {
		c.String(http.StatusOK, "OK")
	})

	// Test server creation
	server := NewServer(8080, engine)
	if server == nil {
		t.Fatal("Failed to create server")
	}

	// Test address setting
	expectedAddr := ":8080"
	if server.GetAddr() != expectedAddr {
		t.Errorf("Expected address %q, got %q", expectedAddr, server.GetAddr())
	}

	// Test address change
	newAddr := ":9090"
	server.SetAddr(newAddr)
	if server.GetAddr() != newAddr {
		t.Errorf("Expected address %q, got %q", newAddr, server.GetAddr())
	}
}

func TestMiddleware(t *testing.T) {
	engine := New() // Use New() to get blank engine without default middleware

	// Add custom middleware
	engine.Use(func(c *Context) {
		c.Header("X-Custom-Header", "test-value")
		c.Next()
	})

	engine.GET("/middleware-test", func(c *Context) {
		c.String(http.StatusOK, "middleware works")
	})

	req, err := http.NewRequest("GET", "/middleware-test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	engine.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	customHeader := w.Header().Get("X-Custom-Header")
	if customHeader != "test-value" {
		t.Errorf("Expected header value %q, got %q", "test-value", customHeader)
	}
}