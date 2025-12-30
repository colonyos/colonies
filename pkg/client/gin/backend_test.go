package gin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewGinClientBackend(t *testing.T) {
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "localhost",
		Port:        8080,
		Insecure:    false,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)
	assert.NotNil(t, backend)
	assert.Equal(t, "localhost", backend.host)
	assert.Equal(t, 8080, backend.port)
	assert.False(t, backend.insecure)

	// Verify it implements required interfaces
	var _ backends.ClientBackend = backend
	var _ backends.RealtimeBackend = backend
	var _ backends.ClientBackendWithRealtime = backend
}

func TestNewGinClientBackendInvalidType(t *testing.T) {
	config := &backends.ClientConfig{
		BackendType: backends.ClientBackendType("invalid"), // Wrong type
		Host:        "localhost",
		Port:        8080,
	}

	backend, err := NewGinClientBackend(config)
	assert.Error(t, err)
	assert.Nil(t, backend)
	assert.Contains(t, err.Error(), "invalid backend type")
}

func TestNewGinClientBackendInsecure(t *testing.T) {
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "localhost",
		Port:        8080,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)
	assert.NotNil(t, backend)
	assert.True(t, backend.insecure)
}

func TestNewGinClientBackendSkipTLSVerify(t *testing.T) {
	config := &backends.ClientConfig{
		BackendType:   backends.GinClientBackendType,
		Host:          "localhost",
		Port:          8443,
		Insecure:      false,
		SkipTLSVerify: true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)
	assert.NotNil(t, backend)
	assert.True(t, backend.skipTLSVerify)
	assert.NotNil(t, backend.restyClient)
}

func TestGinClientBackendSendRawMessage(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "ok"}`))
	}))
	defer server.Close()

	// Parse server URL to get host and port
	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        host,
		Port:        port,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	resp, err := backend.SendRawMessage(`{"test": "message"}`, true)
	assert.NoError(t, err)
	assert.Contains(t, resp, "ok")
}

func TestGinClientBackendSendRawMessageHTTPS(t *testing.T) {
	// Create HTTPS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "ok"}`))
	}))
	defer server.Close()

	// Parse server URL to get host and port
	parts := strings.Split(strings.TrimPrefix(server.URL, "https://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType:   backends.GinClientBackendType,
		Host:          host,
		Port:          port,
		Insecure:      false,
		SkipTLSVerify: true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	resp, err := backend.SendRawMessage(`{"test": "message"}`, false)
	assert.NoError(t, err)
	assert.Contains(t, resp, "ok")
}

func TestGinClientBackendSendRawMessageError(t *testing.T) {
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "invalid-host-that-does-not-exist",
		Port:        99999,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	_, err = backend.SendRawMessage(`{"test": "message"}`, true)
	assert.Error(t, err)
}

func TestGinClientBackendSendMessageInsecure(t *testing.T) {
	// Create a valid RPC reply
	rpcReply, _ := rpc.CreateRPCReplyMsg("test", `{"status": "success"}`)
	replyJSON, _ := rpcReply.ToJSON()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(replyJSON))
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        host,
		Port:        port,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	resp, err := backend.SendMessage("test", `{"data": "test"}`, "", true, context.Background())
	assert.NoError(t, err)
	assert.Contains(t, resp, "success")
}

func TestGinClientBackendSendMessageSecure(t *testing.T) {
	// Create a valid RPC reply
	rpcReply, _ := rpc.CreateRPCReplyMsg("test", `{"status": "success"}`)
	replyJSON, _ := rpcReply.ToJSON()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(replyJSON))
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "https://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType:   backends.GinClientBackendType,
		Host:          host,
		Port:          port,
		Insecure:      false,
		SkipTLSVerify: true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	// Use a valid private key for signing
	prvKey := "ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05"
	resp, err := backend.SendMessage("test", `{"data": "test"}`, prvKey, false, context.Background())
	assert.NoError(t, err)
	assert.Contains(t, resp, "success")
}

func TestGinClientBackendSendMessageWithError(t *testing.T) {
	// Create an error RPC reply
	failure := core.CreateFailure(400, "test error message")
	failureJSON, _ := failure.ToJSON()
	rpcReply, _ := rpc.CreateRPCErrorReplyMsg(rpc.ErrorPayloadType, failureJSON)
	replyJSON, _ := rpcReply.ToJSON()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(replyJSON))
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        host,
		Port:        port,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	_, err = backend.SendMessage("test", `{"data": "test"}`, "", true, context.Background())
	assert.Error(t, err)

	// Verify it's a ColoniesError
	coloniesErr, ok := err.(*core.ColoniesError)
	assert.True(t, ok)
	assert.Equal(t, 400, coloniesErr.Status)
	assert.Contains(t, coloniesErr.Message, "test error message")
}

func TestGinClientBackendSendMessageInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        host,
		Port:        port,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	_, err = backend.SendMessage("test", `{"data": "test"}`, "", true, context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected a valid Colonies RPC message")
}

func TestGinClientBackendSendMessageNetworkError(t *testing.T) {
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "invalid-host-that-does-not-exist",
		Port:        99999,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	_, err = backend.SendMessage("test", `{"data": "test"}`, "", true, context.Background())
	assert.Error(t, err)
}

func TestGinClientBackendSendMessageWithCanceledContext(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        host,
		Port:        port,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	// Create a context that gets canceled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = backend.SendMessage("test", `{"data": "test"}`, "", true, ctx)
	assert.Error(t, err)
}

func TestGinClientBackendCheckHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        host,
		Port:        port,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	err = backend.CheckHealth()
	assert.NoError(t, err)
}

func TestGinClientBackendCheckHealthHTTPS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "https://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType:   backends.GinClientBackendType,
		Host:          host,
		Port:          port,
		Insecure:      false,
		SkipTLSVerify: true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	err = backend.CheckHealth()
	assert.NoError(t, err)
}

func TestGinClientBackendCheckHealthError(t *testing.T) {
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "invalid-host-that-does-not-exist",
		Port:        99999,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	err = backend.CheckHealth()
	assert.Error(t, err)
}

func TestGinClientBackendClose(t *testing.T) {
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "localhost",
		Port:        8080,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	// Close should not return an error
	err = backend.Close()
	assert.NoError(t, err)
}

func TestGinClientBackendEstablishRealtimeConnInsecure(t *testing.T) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pubsub" {
			http.NotFound(w, r)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read the subscription message
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// Echo back confirmation
		conn.WriteMessage(websocket.TextMessage, message)
	}))
	defer server.Close()

	// Parse the server address
	addr := server.Listener.Addr().String()
	parts := strings.Split(addr, ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        host,
		Port:        port,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	conn, err := backend.EstablishRealtimeConn(`{"type": "subscribe"}`)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	// Read the echoed message
	_, data, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Contains(t, string(data), "subscribe")
}

func TestGinClientBackendEstablishRealtimeConnSecure(t *testing.T) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pubsub" {
			http.NotFound(w, r)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read and echo
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		conn.WriteMessage(websocket.TextMessage, message)
	}))
	defer server.Close()

	// Parse the server address
	addr := strings.TrimPrefix(server.URL, "https://")
	parts := strings.Split(addr, ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType:   backends.GinClientBackendType,
		Host:          host,
		Port:          port,
		Insecure:      false,
		SkipTLSVerify: true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	conn, err := backend.EstablishRealtimeConn(`{"type": "subscribe"}`)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()
}

func TestGinClientBackendEstablishRealtimeConnError(t *testing.T) {
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "invalid-host-that-does-not-exist",
		Port:        99999,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	_, err = backend.EstablishRealtimeConn(`{"type": "subscribe"}`)
	assert.Error(t, err)
}

func TestGinClientBackendFactory(t *testing.T) {
	factory := NewGinClientBackendFactory()
	assert.NotNil(t, factory)

	// Test GetBackendType
	backendType := factory.GetBackendType()
	assert.Equal(t, backends.GinClientBackendType, backendType)

	// Test CreateBackend
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "localhost",
		Port:        8080,
		Insecure:    true,
	}

	backend, err := factory.CreateBackend(config)
	assert.NoError(t, err)
	assert.NotNil(t, backend)

	// Verify backend implements required interfaces
	var _ backends.ClientBackend = backend
}

func TestGinClientBackendFactoryInvalidConfig(t *testing.T) {
	factory := NewGinClientBackendFactory()

	config := &backends.ClientConfig{
		BackendType: backends.ClientBackendType("invalid"), // Wrong type
		Host:        "localhost",
		Port:        8080,
	}

	backend, err := factory.CreateBackend(config)
	assert.Error(t, err)
	assert.Nil(t, backend)
}

func TestWebSocketRealtimeConnection(t *testing.T) {
	// Create a WebSocket test server
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo messages back
		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			conn.WriteMessage(msgType, message)
		}
	}))
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws://" + server.Listener.Addr().String()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	// Wrap in RealtimeConnection
	rtConn := NewWebSocketRealtimeConnection(conn)
	assert.NotNil(t, rtConn)

	// Test WriteMessage
	testMsg := []byte("test message")
	err = rtConn.WriteMessage(websocket.TextMessage, testMsg)
	assert.NoError(t, err)

	// Test ReadMessage
	msgType, data, err := rtConn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, websocket.TextMessage, msgType)
	assert.Equal(t, testMsg, data)

	// Test SetReadLimit
	assert.NotPanics(t, func() {
		rtConn.SetReadLimit(1024 * 1024)
	})

	// Test Close
	err = rtConn.Close()
	assert.NoError(t, err)
}

func TestWebSocketRealtimeConnectionImplementsInterface(t *testing.T) {
	// Create a mock WebSocket connection
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		conn.Close()
	}))
	defer server.Close()

	wsURL := "ws://" + server.Listener.Addr().String()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)

	rtConn := NewWebSocketRealtimeConnection(conn)

	// Verify it implements RealtimeConnection interface
	var _ backends.RealtimeConnection = rtConn

	rtConn.Close()
}

func TestGinClientBackendProtocolSelection(t *testing.T) {
	// Test insecure (HTTP)
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "localhost",
		Port:        8080,
		Insecure:    true,
	}
	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)
	assert.True(t, backend.insecure)

	// Test secure (HTTPS)
	config = &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "localhost",
		Port:        8443,
		Insecure:    false,
	}
	backend, err = NewGinClientBackend(config)
	assert.NoError(t, err)
	assert.False(t, backend.insecure)
}

func TestGetGinClientBackendFactory(t *testing.T) {
	factory := GetGinClientBackendFactory()
	assert.NotNil(t, factory)

	// Verify it returns a GinClientBackendFactory
	_, ok := factory.(*GinClientBackendFactory)
	assert.True(t, ok)

	// Verify the factory works
	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        "localhost",
		Port:        8080,
		Insecure:    true,
	}

	backend, err := factory.CreateBackend(config)
	assert.NoError(t, err)
	assert.NotNil(t, backend)
}

func TestGinClientBackendSendMessageWithInvalidFailureJSON(t *testing.T) {
	// Create an error RPC reply with invalid failure JSON
	rpcReply := &rpc.RPCReplyMsg{
		PayloadType: rpc.ErrorPayloadType,
		Error:       true,
		Payload:     "not valid json for failure",
	}
	replyJSON, _ := json.Marshal(rpcReply)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(replyJSON)
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	config := &backends.ClientConfig{
		BackendType: backends.GinClientBackendType,
		Host:        host,
		Port:        port,
		Insecure:    true,
	}

	backend, err := NewGinClientBackend(config)
	assert.NoError(t, err)

	_, err = backend.SendMessage("test", `{"data": "test"}`, "", true, context.Background())
	assert.Error(t, err)
}
