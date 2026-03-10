package relay

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
)

type authRequest struct {
	Timestamp string `json:"timestamp"`
	Signature string `json:"signature"`
}

type authResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type TunnelClient struct {
	relayHost    string
	serverPrvKey string
	localAddr    string
	insecure     bool
	done         chan struct{}
	once         sync.Once
}

func NewTunnelClient(relayHost, serverPrvKey, localAddr string, insecure bool) *TunnelClient {
	return &TunnelClient{
		relayHost:    relayHost,
		serverPrvKey: serverPrvKey,
		localAddr:    localAddr,
		insecure:     insecure,
		done:         make(chan struct{}),
	}
}

// Start launches the reconnect loop in a background goroutine.
func (tc *TunnelClient) Start() {
	go tc.reconnectLoop()
}

// Stop signals the tunnel client to shut down.
func (tc *TunnelClient) Stop() {
	tc.once.Do(func() {
		close(tc.done)
	})
}

func (tc *TunnelClient) reconnectLoop() {
	backoff := initialBackoff

	for {
		select {
		case <-tc.done:
			return
		default:
		}

		err := tc.connect()
		if err != nil {
			log.WithFields(log.Fields{"Error": err, "Backoff": backoff}).Warn("Relay tunnel disconnected, reconnecting")
		}

		select {
		case <-tc.done:
			return
		case <-time.After(backoff):
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (tc *TunnelClient) wsURL() string {
	scheme := "wss"
	if tc.insecure {
		scheme = "ws"
	}
	return fmt.Sprintf("%s://%s/tunnel", scheme, tc.relayHost)
}

func (tc *TunnelClient) localURL(path string) string {
	scheme := "http"
	return fmt.Sprintf("%s://%s%s", scheme, tc.localAddr, path)
}

func (tc *TunnelClient) connect() error {
	url := tc.wsURL()
	log.WithFields(log.Fields{"URL": url}).Info("Connecting to relay tunnel")

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	if tc.insecure {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	// Authenticate
	if err := tc.authenticate(conn); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	log.Info("Relay tunnel connected and authenticated")

	// Reset backoff on successful connection by handling requests
	return tc.handleRequests(conn)
}

func (tc *TunnelClient) authenticate(conn *websocket.Conn) error {
	c := crypto.CreateCrypto()
	ts := time.Now().UTC().Format(time.RFC3339)

	sig, err := c.GenerateSignature(ts, tc.serverPrvKey)
	if err != nil {
		return fmt.Errorf("failed to sign timestamp: %w", err)
	}

	authMsg := authRequest{
		Timestamp: ts,
		Signature: sig,
	}

	authJSON, err := json.Marshal(authMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, authJSON); err != nil {
		return fmt.Errorf("failed to send auth message: %w", err)
	}

	_, respData, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	var resp authResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("invalid auth response: %w", err)
	}

	if resp.Status != "connected" {
		return fmt.Errorf("auth rejected: %s", resp.Error)
	}

	return nil
}

func (tc *TunnelClient) handleRequests(conn *websocket.Conn) error {
	// Use a mutex to serialize writes to the WebSocket connection
	var writeMu sync.Mutex

	// Track active local WS connections
	var wsConnsMu sync.Mutex
	wsConns := make(map[string]*websocket.Conn)

	defer func() {
		// Clean up all local WS connections on disconnect
		wsConnsMu.Lock()
		for _, localWS := range wsConns {
			localWS.Close()
		}
		wsConnsMu.Unlock()
	}()

	for {
		select {
		case <-tc.done:
			return nil
		default:
		}

		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		if msgType != websocket.BinaryMessage {
			continue
		}

		frame, err := Unmarshal(data)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Warn("Failed to unmarshal frame")
			continue
		}

		switch frame.Type {
		case FrameTypeRequest:
			go tc.forwardRequest(conn, &writeMu, frame)

		case FrameTypeWSUpgrade:
			go tc.handleWSUpgrade(conn, &writeMu, frame, &wsConnsMu, wsConns)

		case FrameTypeWSData:
			wsConnsMu.Lock()
			localWS, ok := wsConns[frame.ID]
			wsConnsMu.Unlock()
			if ok {
				wsMsgType := websocket.TextMessage
				if frame.Method == WSMsgTypeBinary {
					wsMsgType = websocket.BinaryMessage
				}
				if err := localWS.WriteMessage(wsMsgType, frame.Body); err != nil {
					log.WithFields(log.Fields{"Error": err, "ConnID": frame.ID}).Warn("Failed to write to local WS")
				}
			}

		case FrameTypeWSClose:
			wsConnsMu.Lock()
			localWS, ok := wsConns[frame.ID]
			delete(wsConns, frame.ID)
			wsConnsMu.Unlock()
			if ok {
				localWS.Close()
			}
		}
	}
}

// handleWSUpgrade opens a local WebSocket to the ColonyOS server and forwards messages bidirectionally.
func (tc *TunnelClient) handleWSUpgrade(
	tunnelConn *websocket.Conn,
	writeMu *sync.Mutex,
	frame *Frame,
	wsConnsMu *sync.Mutex,
	wsConns map[string]*websocket.Conn,
) {
	connID := frame.ID

	// Build local WS URL
	localWSURL := fmt.Sprintf("ws://%s%s", tc.localAddr, frame.Path)

	// Forward relevant headers, excluding per-hop WS upgrade headers
	// (gorilla/websocket adds its own Sec-Websocket-*, Connection, Upgrade headers)
	reqHeaders := http.Header{}
	for key, values := range frame.Headers {
		lowerKey := strings.ToLower(key)
		if strings.HasPrefix(lowerKey, "sec-websocket") ||
			lowerKey == "connection" || lowerKey == "upgrade" {
			continue
		}
		for _, v := range values {
			reqHeaders.Add(key, v)
		}
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	localWS, _, err := dialer.Dial(localWSURL, reqHeaders)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "URL": localWSURL}).Warn("Failed to open local WS")
		// Send WSClose back to relay
		closeFrame := &Frame{ID: connID, Type: FrameTypeWSClose}
		if closeData, err := closeFrame.Marshal(); err == nil {
			writeMu.Lock()
			tunnelConn.WriteMessage(websocket.BinaryMessage, closeData)
			writeMu.Unlock()
		}
		return
	}

	// Register local WS connection
	wsConnsMu.Lock()
	wsConns[connID] = localWS
	wsConnsMu.Unlock()

	// Send WSUpgradeOK
	okFrame := &Frame{ID: connID, Type: FrameTypeWSUpgradeOK}
	okData, err := okFrame.Marshal()
	if err != nil {
		localWS.Close()
		return
	}
	writeMu.Lock()
	err = tunnelConn.WriteMessage(websocket.BinaryMessage, okData)
	writeMu.Unlock()
	if err != nil {
		localWS.Close()
		return
	}

	// Read from local WS, forward through tunnel as WSData
	go func() {
		defer func() {
			wsConnsMu.Lock()
			delete(wsConns, connID)
			wsConnsMu.Unlock()
			localWS.Close()

			// Send WSClose to relay
			closeFrame := &Frame{ID: connID, Type: FrameTypeWSClose}
			if closeData, err := closeFrame.Marshal(); err == nil {
				writeMu.Lock()
				tunnelConn.WriteMessage(websocket.BinaryMessage, closeData)
				writeMu.Unlock()
			}
		}()

		for {
			msgType, msg, err := localWS.ReadMessage()
			if err != nil {
				return
			}

			wsMsgType := WSMsgTypeText
			if msgType == websocket.BinaryMessage {
				wsMsgType = WSMsgTypeBinary
			}

			dataFrame := &Frame{
				ID:     connID,
				Type:   FrameTypeWSData,
				Method: wsMsgType,
				Body:   msg,
			}
			frameData, err := dataFrame.Marshal()
			if err != nil {
				return
			}
			writeMu.Lock()
			err = tunnelConn.WriteMessage(websocket.BinaryMessage, frameData)
			writeMu.Unlock()
			if err != nil {
				return
			}
		}
	}()
}

func (tc *TunnelClient) forwardRequest(conn *websocket.Conn, writeMu *sync.Mutex, reqFrame *Frame) {
	respFrame := tc.doHTTPRequest(reqFrame)

	data, err := respFrame.Marshal()
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "FrameID": reqFrame.ID}).Error("Failed to marshal response frame")
		return
	}

	writeMu.Lock()
	err = conn.WriteMessage(websocket.BinaryMessage, data)
	writeMu.Unlock()

	if err != nil {
		log.WithFields(log.Fields{"Error": err, "FrameID": reqFrame.ID}).Error("Failed to send response frame")
	}
}

func (tc *TunnelClient) doHTTPRequest(reqFrame *Frame) *Frame {
	method := MethodToString(reqFrame.Method)
	url := tc.localURL(reqFrame.Path)

	var bodyReader io.Reader
	if len(reqFrame.Body) > 0 {
		bodyReader = bytes.NewReader(reqFrame.Body)
	}

	httpReq, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return tc.errorFrame(reqFrame.ID, http.StatusBadGateway, fmt.Sprintf("failed to create request: %v", err))
	}

	// Copy headers from the frame to the HTTP request
	for key, values := range reqFrame.Headers {
		for _, v := range values {
			httpReq.Header.Add(key, v)
		}
	}

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return tc.errorFrame(reqFrame.ID, http.StatusBadGateway, fmt.Sprintf("request failed: %v", err))
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return tc.errorFrame(reqFrame.ID, http.StatusBadGateway, fmt.Sprintf("failed to read response body: %v", err))
	}

	// Copy response headers
	respHeaders := make(map[string][]string)
	for key, values := range httpResp.Header {
		respHeaders[key] = values
	}

	return &Frame{
		ID:         reqFrame.ID,
		Type:       FrameTypeResponse,
		StatusCode: uint16(httpResp.StatusCode),
		Headers:    respHeaders,
		Body:       respBody,
	}
}

func (tc *TunnelClient) errorFrame(id string, statusCode int, msg string) *Frame {
	return &Frame{
		ID:         id,
		Type:       FrameTypeResponse,
		StatusCode: uint16(statusCode),
		Headers:    map[string][]string{"Content-Type": {"text/plain"}},
		Body:       []byte(msg),
	}
}
