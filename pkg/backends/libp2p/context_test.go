package libp2p

import (
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/stretchr/testify/assert"
)

func TestStreamContextImplementsInterface(t *testing.T) {
	// Create a mock stream context
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Verify it implements backends.Context interface
	var _ backends.Context = ctx
}

func TestStreamContextGetHeader(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// LibP2P doesn't have headers, should return empty string
	header := ctx.GetHeader("Content-Type")
	assert.Empty(t, header)

	header = ctx.GetHeader("Authorization")
	assert.Empty(t, header)
}

func TestStreamContextQuery(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// LibP2P doesn't have query params, should return empty string
	query := ctx.Query("key")
	assert.Empty(t, query)
}

func TestStreamContextDefaultQuery(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Should return default value
	query := ctx.DefaultQuery("key", "default")
	assert.Equal(t, "default", query)

	query = ctx.DefaultQuery("another", "value123")
	assert.Equal(t, "value123", query)
}

func TestStreamContextParam(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// LibP2P doesn't have URL params, should return empty string
	param := ctx.Param("id")
	assert.Empty(t, param)
}

func TestStreamContextPostForm(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// LibP2P doesn't have POST forms, should return empty string
	form := ctx.PostForm("username")
	assert.Empty(t, form)
}

func TestStreamContextDefaultPostForm(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Should return default value
	form := ctx.DefaultPostForm("username", "guest")
	assert.Equal(t, "guest", form)

	form = ctx.DefaultPostForm("role", "user")
	assert.Equal(t, "user", form)
}

func TestStreamContextSetGet(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Test setting and getting values
	ctx.Set("key1", "value1")
	ctx.Set("key2", 42)
	ctx.Set("key3", true)

	// Get existing keys
	val, exists := ctx.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", val)

	val, exists = ctx.Get("key2")
	assert.True(t, exists)
	assert.Equal(t, 42, val)

	val, exists = ctx.Get("key3")
	assert.True(t, exists)
	assert.Equal(t, true, val)

	// Get non-existent key
	val, exists = ctx.Get("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, val)
}

func TestStreamContextGetString(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Set string value
	ctx.Set("name", "Alice")
	assert.Equal(t, "Alice", ctx.GetString("name"))

	// Set non-string value
	ctx.Set("count", 42)
	assert.Equal(t, "", ctx.GetString("count"))

	// Get non-existent key
	assert.Equal(t, "", ctx.GetString("missing"))
}

func TestStreamContextGetBool(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Set bool value
	ctx.Set("active", true)
	assert.True(t, ctx.GetBool("active"))

	ctx.Set("inactive", false)
	assert.False(t, ctx.GetBool("inactive"))

	// Set non-bool value
	ctx.Set("text", "true")
	assert.False(t, ctx.GetBool("text"))

	// Get non-existent key
	assert.False(t, ctx.GetBool("missing"))
}

func TestStreamContextGetInt(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Set int value
	ctx.Set("age", 25)
	assert.Equal(t, 25, ctx.GetInt("age"))

	ctx.Set("zero", 0)
	assert.Equal(t, 0, ctx.GetInt("zero"))

	// Set non-int value
	ctx.Set("text", "123")
	assert.Equal(t, 0, ctx.GetInt("text"))

	// Get non-existent key
	assert.Equal(t, 0, ctx.GetInt("missing"))
}

func TestStreamContextGetInt64(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Set int64 value
	ctx.Set("bignum", int64(1234567890))
	assert.Equal(t, int64(1234567890), ctx.GetInt64("bignum"))

	ctx.Set("zero", int64(0))
	assert.Equal(t, int64(0), ctx.GetInt64("zero"))

	// Set non-int64 value
	ctx.Set("text", "123")
	assert.Equal(t, int64(0), ctx.GetInt64("text"))

	// Get non-existent key
	assert.Equal(t, int64(0), ctx.GetInt64("missing"))
}

func TestStreamContextGetFloat64(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Set float64 value
	ctx.Set("price", 19.99)
	assert.Equal(t, 19.99, ctx.GetFloat64("price"))

	ctx.Set("zero", 0.0)
	assert.Equal(t, 0.0, ctx.GetFloat64("zero"))

	// Set non-float64 value
	ctx.Set("text", "3.14")
	assert.Equal(t, 0.0, ctx.GetFloat64("text"))

	// Get non-existent key
	assert.Equal(t, 0.0, ctx.GetFloat64("missing"))
}

func TestStreamContextAbort(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Initially not aborted
	assert.False(t, ctx.IsAborted())

	// Abort
	ctx.Abort()
	assert.True(t, ctx.IsAborted())
}

func TestStreamContextAbortWithStatus(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Initially not aborted
	assert.False(t, ctx.IsAborted())

	// Abort with status
	ctx.AbortWithStatus(400)
	assert.True(t, ctx.IsAborted())
}

func TestStreamContextStatus(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Status is a no-op for libp2p, should not panic
	assert.NotPanics(t, func() {
		ctx.Status(200)
		ctx.Status(404)
		ctx.Status(500)
	})
}

func TestStreamContextHeader(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Header is a no-op for libp2p, should not panic
	assert.NotPanics(t, func() {
		ctx.Header("Content-Type", "application/json")
		ctx.Header("Authorization", "Bearer token")
	})
}

func TestStreamContextRequest(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// LibP2P doesn't have HTTP requests
	req := ctx.Request()
	assert.Nil(t, req)
}

func TestStreamContextNext(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Next is a no-op for libp2p, should not panic
	assert.NotPanics(t, func() {
		ctx.Next()
	})
}

func TestStreamContextGetPeerID(t *testing.T) {
	ctx := &StreamContext{
		peerID: "12D3KooWBrsnBU9rZ8ZBaniVexPfdLmYyF34doTRtSJ7XqfC3JfM",
		store:  make(map[string]interface{}),
	}

	assert.Equal(t, "12D3KooWBrsnBU9rZ8ZBaniVexPfdLmYyF34doTRtSJ7XqfC3JfM", ctx.GetPeerID())
}

func TestStreamContextGetStream(t *testing.T) {
	ctx := &StreamContext{
		stream: nil, // No mock stream
		store:  make(map[string]interface{}),
	}

	// Should return nil when no stream is set
	stream := ctx.GetStream()
	assert.Nil(t, stream)
}

func TestStreamContextGetPubSub(t *testing.T) {
	ctx := &StreamContext{
		pubsub: nil, // No mock pubsub
		store:  make(map[string]interface{}),
	}

	// Should return nil when no pubsub is set
	pubsub := ctx.GetPubSub()
	assert.Nil(t, pubsub)
}

func TestStreamContextMultipleValues(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Set multiple values of different types
	ctx.Set("string", "hello")
	ctx.Set("int", 42)
	ctx.Set("int64", int64(9999))
	ctx.Set("float64", 3.14)
	ctx.Set("bool", true)

	// Verify all values
	assert.Equal(t, "hello", ctx.GetString("string"))
	assert.Equal(t, 42, ctx.GetInt("int"))
	assert.Equal(t, int64(9999), ctx.GetInt64("int64"))
	assert.Equal(t, 3.14, ctx.GetFloat64("float64"))
	assert.True(t, ctx.GetBool("bool"))
}

func TestStreamContextOverwriteValue(t *testing.T) {
	ctx := &StreamContext{
		store: make(map[string]interface{}),
	}

	// Set initial value
	ctx.Set("key", "value1")
	assert.Equal(t, "value1", ctx.GetString("key"))

	// Overwrite with new value
	ctx.Set("key", "value2")
	assert.Equal(t, "value2", ctx.GetString("key"))

	// Overwrite with different type
	ctx.Set("key", 123)
	assert.Equal(t, "", ctx.GetString("key")) // No longer a string
	assert.Equal(t, 123, ctx.GetInt("key"))   // Now an int
}
