package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/gorilla/websocket"
)

// testRelay is a minimal relay server for testing the tunnel client.
// It supports concurrent sendRequest calls using a write mutex and read-loop dispatcher.
type testRelay struct {
	server     *httptest.Server
	serverID   string
	prvKey     string
	upgrader   websocket.Upgrader
	mu         sync.Mutex
	tunnelConn *websocket.Conn
	writeMu    sync.Mutex
	pending    map[string]chan *Frame
	pendingMu  sync.Mutex
}

func newTestRelay(t *testing.T, prvKey string) *testRelay {
	t.Helper()
	c := crypto.CreateCrypto()
	serverID, err := c.GenerateID(prvKey)
	if err != nil {
		t.Fatalf("failed to generate server ID: %v", err)
	}

	tr := &testRelay{
		serverID: serverID,
		prvKey:   prvKey,
		pending:  make(map[string]chan *Frame),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/tunnel", tr.handleTunnel)
	tr.server = httptest.NewServer(mux)

	return tr
}

func (tr *testRelay) addr() string {
	return tr.server.Listener.Addr().String()
}

func (tr *testRelay) close() {
	tr.mu.Lock()
	if tr.tunnelConn != nil {
		tr.tunnelConn.Close()
	}
	tr.mu.Unlock()
	tr.server.Close()
}

func (tr *testRelay) handleTunnel(w http.ResponseWriter, r *http.Request) {
	conn, err := tr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Read auth message
	_, data, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	var authReq struct {
		Timestamp string `json:"timestamp"`
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal(data, &authReq); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"status":"error","error":"invalid auth"}`))
		conn.Close()
		return
	}

	// Verify signature using colonies crypto
	c := crypto.CreateCrypto()
	recoveredID, err := c.RecoverID(authReq.Timestamp, authReq.Signature)
	if err != nil || recoveredID != tr.serverID {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"status":"error","error":"auth failed"}`))
		conn.Close()
		return
	}

	conn.WriteMessage(websocket.TextMessage, []byte(`{"status":"connected"}`))

	tr.mu.Lock()
	tr.tunnelConn = conn
	tr.mu.Unlock()

	// Start read loop to dispatch responses to pending requests
	go tr.readLoop(conn)
}

func (tr *testRelay) readLoop(conn *websocket.Conn) {
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if msgType != websocket.BinaryMessage {
			continue
		}
		frame, err := Unmarshal(data)
		if err != nil {
			continue
		}
		if frame.Type == FrameTypeResponse {
			tr.pendingMu.Lock()
			ch, ok := tr.pending[frame.ID]
			tr.pendingMu.Unlock()
			if ok {
				ch <- frame
			}
		}
	}
}

// sendRequest sends a request frame through the tunnel and waits for a response.
// Safe to call concurrently from multiple goroutines.
func (tr *testRelay) sendRequest(frame *Frame, timeout time.Duration) (*Frame, error) {
	tr.mu.Lock()
	conn := tr.tunnelConn
	tr.mu.Unlock()

	if conn == nil {
		return nil, fmt.Errorf("no tunnel connection")
	}

	// Register response channel before sending
	respCh := make(chan *Frame, 1)
	tr.pendingMu.Lock()
	tr.pending[frame.ID] = respCh
	tr.pendingMu.Unlock()
	defer func() {
		tr.pendingMu.Lock()
		delete(tr.pending, frame.ID)
		tr.pendingMu.Unlock()
	}()

	data, err := frame.Marshal()
	if err != nil {
		return nil, err
	}

	tr.writeMu.Lock()
	err = conn.WriteMessage(websocket.BinaryMessage, data)
	tr.writeMu.Unlock()
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-respCh:
		return resp, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

func generateTestKey(t *testing.T) string {
	t.Helper()
	c := crypto.CreateCrypto()
	key, err := c.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	return key
}

func waitForTunnelConnection(tr *testRelay, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		tr.mu.Lock()
		connected := tr.tunnelConn != nil
		tr.mu.Unlock()
		if connected {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func TestTunnelClientAuthAndForward(t *testing.T) {
	prvKey := generateTestKey(t)

	// Start a local HTTP server to act as the colonies server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "hello")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("colonies response"))
	})
	localServer := httptest.NewServer(handler)
	defer localServer.Close()

	// Start test relay
	tr := newTestRelay(t, prvKey)
	defer tr.close()

	// Start tunnel client
	tc := NewTunnelClient(tr.addr(), prvKey, localServer.Listener.Addr().String(), true)
	tc.Start()
	defer tc.Stop()

	// Wait for tunnel to connect
	if !waitForTunnelConnection(tr, 5*time.Second) {
		t.Fatal("tunnel did not connect within timeout")
	}

	// Send a request through the relay
	reqFrame := &Frame{
		ID:      "aabbccdd11223344aabbccdd11223344",
		Type:    FrameTypeRequest,
		Method:  MethodGET,
		Path:    "/test",
		Headers: map[string][]string{"Accept": {"application/json"}},
	}

	resp, err := tr.sendRequest(reqFrame, 5*time.Second)
	if err != nil {
		t.Fatalf("sendRequest failed: %v", err)
	}

	if resp.ID != reqFrame.ID {
		t.Errorf("response ID mismatch: got %s, want %s", resp.ID, reqFrame.ID)
	}
	if resp.Type != FrameTypeResponse {
		t.Errorf("response type: got %d, want %d", resp.Type, FrameTypeResponse)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if string(resp.Body) != "colonies response" {
		t.Errorf("body: got %q, want %q", string(resp.Body), "colonies response")
	}
	if vals, ok := resp.Headers["X-Test"]; !ok || len(vals) == 0 || vals[0] != "hello" {
		t.Errorf("expected X-Test header to be 'hello', got %v", resp.Headers["X-Test"])
	}
}

func TestTunnelClientMultiValueHeaders(t *testing.T) {
	prvKey := generateTestKey(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back request Accept values and set multi-value response header
		w.Header().Add("Set-Cookie", "a=1")
		w.Header().Add("Set-Cookie", "b=2")
		w.WriteHeader(http.StatusOK)
	})
	localServer := httptest.NewServer(handler)
	defer localServer.Close()

	tr := newTestRelay(t, prvKey)
	defer tr.close()

	tc := NewTunnelClient(tr.addr(), prvKey, localServer.Listener.Addr().String(), true)
	tc.Start()
	defer tc.Stop()

	if !waitForTunnelConnection(tr, 5*time.Second) {
		t.Fatal("tunnel did not connect")
	}

	reqFrame := &Frame{
		ID:     "00112233445566778899aabbccddeeff",
		Type:   FrameTypeRequest,
		Method: MethodGET,
		Path:   "/cookies",
	}

	resp, err := tr.sendRequest(reqFrame, 5*time.Second)
	if err != nil {
		t.Fatalf("sendRequest failed: %v", err)
	}

	cookies := resp.Headers["Set-Cookie"]
	if len(cookies) != 2 {
		t.Fatalf("expected 2 Set-Cookie values, got %d: %v", len(cookies), cookies)
	}
	if cookies[0] != "a=1" || cookies[1] != "b=2" {
		t.Errorf("unexpected Set-Cookie values: %v", cookies)
	}
}

func TestTunnelClientPOSTWithBody(t *testing.T) {
	prvKey := generateTestKey(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write(body) // Echo body back
	})
	localServer := httptest.NewServer(handler)
	defer localServer.Close()

	tr := newTestRelay(t, prvKey)
	defer tr.close()

	tc := NewTunnelClient(tr.addr(), prvKey, localServer.Listener.Addr().String(), true)
	tc.Start()
	defer tc.Stop()

	if !waitForTunnelConnection(tr, 5*time.Second) {
		t.Fatal("tunnel did not connect")
	}

	payload := []byte(`{"key":"value"}`)
	reqFrame := &Frame{
		ID:      "11111111111111111111111111111111",
		Type:    FrameTypeRequest,
		Method:  MethodPOST,
		Path:    "/api/data",
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    payload,
	}

	resp, err := tr.sendRequest(reqFrame, 5*time.Second)
	if err != nil {
		t.Fatalf("sendRequest failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status code: got %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	if !bytes.Equal(resp.Body, payload) {
		t.Errorf("body: got %q, want %q", string(resp.Body), string(payload))
	}
}

func TestTunnelClientBinaryBody(t *testing.T) {
	prvKey := generateTestKey(t)

	// Create binary payload with all byte values 0-255
	binaryPayload := make([]byte, 256)
	for i := 0; i < 256; i++ {
		binaryPayload[i] = byte(i)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})
	localServer := httptest.NewServer(handler)
	defer localServer.Close()

	tr := newTestRelay(t, prvKey)
	defer tr.close()

	tc := NewTunnelClient(tr.addr(), prvKey, localServer.Listener.Addr().String(), true)
	tc.Start()
	defer tc.Stop()

	if !waitForTunnelConnection(tr, 5*time.Second) {
		t.Fatal("tunnel did not connect")
	}

	reqFrame := &Frame{
		ID:      "22222222222222222222222222222222",
		Type:    FrameTypeRequest,
		Method:  MethodPOST,
		Path:    "/binary",
		Headers: map[string][]string{"Content-Type": {"application/octet-stream"}},
		Body:    binaryPayload,
	}

	resp, err := tr.sendRequest(reqFrame, 5*time.Second)
	if err != nil {
		t.Fatalf("sendRequest failed: %v", err)
	}

	if !bytes.Equal(resp.Body, binaryPayload) {
		t.Errorf("binary body round-trip failed: got %d bytes, want %d bytes", len(resp.Body), len(binaryPayload))
	}
}

func TestTunnelClientConcurrentRequests(t *testing.T) {
	prvKey := generateTestKey(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Small delay to ensure requests overlap
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.Path))
	})
	localServer := httptest.NewServer(handler)
	defer localServer.Close()

	tr := newTestRelay(t, prvKey)
	defer tr.close()

	tc := NewTunnelClient(tr.addr(), prvKey, localServer.Listener.Addr().String(), true)
	tc.Start()
	defer tc.Stop()

	if !waitForTunnelConnection(tr, 5*time.Second) {
		t.Fatal("tunnel did not connect")
	}

	const numRequests = 10
	var wg sync.WaitGroup
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			id := fmt.Sprintf("%032x", idx)
			path := fmt.Sprintf("/req/%d", idx)

			reqFrame := &Frame{
				ID:     id,
				Type:   FrameTypeRequest,
				Method: MethodGET,
				Path:   path,
			}

			resp, err := tr.sendRequest(reqFrame, 10*time.Second)
			if err != nil {
				errors <- fmt.Errorf("request %d failed: %w", idx, err)
				return
			}

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("request %d: status %d", idx, resp.StatusCode)
				return
			}

			if string(resp.Body) != path {
				errors <- fmt.Errorf("request %d: body %q, want %q", idx, string(resp.Body), path)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

func TestTunnelClientReconnect(t *testing.T) {
	prvKey := generateTestKey(t)

	var requestCount int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	localServer := httptest.NewServer(handler)
	defer localServer.Close()

	// Start relay on a fixed port so we can restart it
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	startRelay := func() *testRelay {
		tr := newTestRelay(t, prvKey)
		// Replace the listener with our fixed address
		tr.server.Close()

		l, err := net.Listen("tcp", addr)
		if err != nil {
			t.Fatalf("failed to listen on %s: %v", addr, err)
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/tunnel", tr.handleTunnel)
		tr.server = &httptest.Server{
			Listener: l,
			Config:   &http.Server{Handler: mux},
		}
		tr.server.Start()
		return tr
	}

	// Start first relay
	relay1 := startRelay()

	tc := NewTunnelClient(addr, prvKey, localServer.Listener.Addr().String(), true)
	tc.Start()
	defer tc.Stop()

	if !waitForTunnelConnection(relay1, 5*time.Second) {
		t.Fatal("tunnel did not connect to first relay")
	}

	// Send a request through first connection
	reqFrame := &Frame{
		ID:     "aaaabbbbccccddddaaaabbbbccccdddd",
		Type:   FrameTypeRequest,
		Method: MethodGET,
		Path:   "/first",
	}
	resp, err := relay1.sendRequest(reqFrame, 5*time.Second)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first request: status %d", resp.StatusCode)
	}

	// Kill the relay - this will cause the tunnel client to reconnect
	relay1.close()

	// Wait a moment, then start second relay on same address
	time.Sleep(500 * time.Millisecond)
	relay2 := startRelay()
	defer relay2.close()

	// Wait for reconnection
	if !waitForTunnelConnection(relay2, 10*time.Second) {
		t.Fatal("tunnel did not reconnect to second relay")
	}

	// Send a request through the reconnected tunnel
	reqFrame2 := &Frame{
		ID:     "11112222333344441111222233334444",
		Type:   FrameTypeRequest,
		Method: MethodGET,
		Path:   "/second",
	}
	resp2, err := relay2.sendRequest(reqFrame2, 5*time.Second)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("second request: status %d", resp2.StatusCode)
	}

	// Verify both requests were forwarded to the local server
	count := atomic.LoadInt32(&requestCount)
	if count != 2 {
		t.Errorf("expected 2 requests forwarded, got %d", count)
	}
}

func TestTunnelClientGracefulShutdown(t *testing.T) {
	prvKey := generateTestKey(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	localServer := httptest.NewServer(handler)
	defer localServer.Close()

	tr := newTestRelay(t, prvKey)
	defer tr.close()

	tc := NewTunnelClient(tr.addr(), prvKey, localServer.Listener.Addr().String(), true)
	tc.Start()

	if !waitForTunnelConnection(tr, 5*time.Second) {
		t.Fatal("tunnel did not connect")
	}

	// Stop should return quickly without blocking
	done := make(chan struct{})
	go func() {
		tc.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Stop() did not return within timeout")
	}
}

func TestTunnelClientAuthFailure(t *testing.T) {
	prvKey := generateTestKey(t)
	wrongKey := generateTestKey(t)

	// Relay expects prvKey but tunnel client uses wrongKey
	tr := newTestRelay(t, prvKey)
	defer tr.close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	localServer := httptest.NewServer(handler)
	defer localServer.Close()

	tc := NewTunnelClient(tr.addr(), wrongKey, localServer.Listener.Addr().String(), true)
	tc.Start()
	defer tc.Stop()

	// The tunnel should not successfully connect
	time.Sleep(2 * time.Second)
	tr.mu.Lock()
	connected := tr.tunnelConn != nil
	tr.mu.Unlock()

	if connected {
		t.Error("tunnel should not have connected with wrong key")
	}
}
