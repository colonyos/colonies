package gin

import (
	"context"
	"crypto/tls"
	"errors"
	"net/url"
	"strconv"

	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
)

// GinClientBackend implements HTTP/REST client backend using Gin/Resty
type GinClientBackend struct {
	restyClient   *resty.Client
	host          string
	port          int
	insecure      bool
	skipTLSVerify bool
}

// NewGinClientBackend creates a new Gin client backend
func NewGinClientBackend(config *backends.ClientConfig) (*GinClientBackend, error) {
	if config.BackendType != backends.GinClientBackendType {
		return nil, errors.New("invalid backend type for gin client")
	}

	backend := &GinClientBackend{
		host:          config.Host,
		port:          config.Port,
		insecure:      config.Insecure,
		skipTLSVerify: config.SkipTLSVerify,
		restyClient:   resty.New(),
	}

	if config.SkipTLSVerify {
		backend.restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return backend, nil
}

// SendRawMessage sends a raw JSON message via HTTP
func (g *GinClientBackend) SendRawMessage(jsonString string, insecure bool) (string, error) {
	protocol := "https"
	if g.insecure {
		protocol = "http"
	}
	resp, err := g.restyClient.R().
		SetBody(jsonString).
		Post(protocol + "://" + g.host + ":" + strconv.Itoa(g.port) + "/api")
	if err != nil {
		return "", err
	}

	return string(resp.Body()), nil
}

// SendMessage sends an RPC message with authentication via HTTP
func (g *GinClientBackend) SendMessage(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
	var rpcMsg *rpc.RPCMsg
	var err error
	if insecure {
		rpcMsg, err = rpc.CreateInsecureRPCMsg(method, jsonString)
		if err != nil {
			return "", err
		}
	} else {
		rpcMsg, err = rpc.CreateRPCMsg(method, jsonString, prvKey)
		if err != nil {
			return "", err
		}
	}
	jsonString, err = rpcMsg.ToJSON()
	if err != nil {
		return "", err
	}

	protocol := "https"
	if g.insecure {
		protocol = "http"
	}
	resp, err := g.restyClient.R().
		SetContext(ctx).
		SetBody(jsonString).
		Post(protocol + "://" + g.host + ":" + strconv.Itoa(g.port) + "/api")
	if err != nil {
		return "", err
	}

	respBodyString := string(resp.Body())

	rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(respBodyString)
	if err != nil {
		return "", errors.New("Expected a valid Colonies RPC message, but got this: " + respBodyString)
	}

	if rpcReplyMsg.Error {
		failure, err := core.ConvertJSONToFailure(rpcReplyMsg.DecodePayload())
		if err != nil {
			return "", err
		}

		return "", &core.ColoniesError{Status: failure.Status, Message: failure.Message}
	}

	return rpcReplyMsg.DecodePayload(), nil
}

// EstablishRealtimeConn establishes a realtime connection using WebSocket
func (g *GinClientBackend) EstablishRealtimeConn(jsonString string) (backends.RealtimeConnection, error) {
	dialer := *websocket.DefaultDialer
	var u url.URL

	if g.insecure {
		u = url.URL{Scheme: "ws", Host: g.host + ":" + strconv.Itoa(g.port), Path: "/pubsub"}
	} else {
		u = url.URL{Scheme: "wss", Host: g.host + ":" + strconv.Itoa(g.port), Path: "/pubsub"}
		if g.skipTLSVerify {
			dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}

	wsConn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	err = wsConn.WriteMessage(websocket.TextMessage, []byte(jsonString))
	if err != nil {
		return nil, err
	}

	return NewWebSocketRealtimeConnection(wsConn), nil
}

// CheckHealth checks the health of the server via HTTP
func (g *GinClientBackend) CheckHealth() error {
	protocol := "https"
	if g.insecure {
		protocol = "http"
	}
	_, err := g.restyClient.R().
		Get(protocol + "://" + g.host + ":" + strconv.Itoa(g.port) + "/health")

	return err
}

// Close closes the backend and cleans up resources
func (g *GinClientBackend) Close() error {
	// Resty client doesn't need explicit cleanup
	return nil
}

// GinClientBackendFactory creates gin client backends
type GinClientBackendFactory struct{}

// NewGinClientBackendFactory creates a new gin client backend factory
func NewGinClientBackendFactory() *GinClientBackendFactory {
	return &GinClientBackendFactory{}
}

// CreateBackend creates a new gin client backend
func (f *GinClientBackendFactory) CreateBackend(config *backends.ClientConfig) (backends.ClientBackend, error) {
	return NewGinClientBackend(config)
}

// GetBackendType returns the backend type this factory creates
func (f *GinClientBackendFactory) GetBackendType() backends.ClientBackendType {
	return backends.GinClientBackendType
}

// Compile-time checks that GinClientBackend implements the required interfaces
var _ backends.ClientBackend = (*GinClientBackend)(nil)
var _ backends.RealtimeBackend = (*GinClientBackend)(nil)
var _ backends.ClientBackendWithRealtime = (*GinClientBackend)(nil)