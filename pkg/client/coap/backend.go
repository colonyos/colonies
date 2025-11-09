package coap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/udp"
)

// CoAPClientBackend implements client backend using CoAP
type CoAPClientBackend struct {
	host          string
	port          int
	isInsecure    bool
	skipTLSVerify bool
}

// NewCoAPClientBackend creates a new CoAP client backend
func NewCoAPClientBackend(config *backends.ClientConfig) (*CoAPClientBackend, error) {
	if config.BackendType != backends.CoAPClientBackendType {
		return nil, errors.New("invalid backend type for CoAP client")
	}

	backend := &CoAPClientBackend{
		host:          config.Host,
		port:          config.Port,
		isInsecure:    config.Insecure,
		skipTLSVerify: config.SkipTLSVerify,
	}

	return backend, nil
}

// SendRawMessage sends a raw JSON message via CoAP
func (c *CoAPClientBackend) SendRawMessage(jsonString string, insecure bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to CoAP server
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := udp.Dial(addr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to CoAP server: %w", err)
	}
	defer conn.Close()

	// Create POST request to /api endpoint
	resp, err := conn.Post(ctx, "/api", message.AppJSON, bytes.NewReader([]byte(jsonString)))
	if err != nil {
		return "", fmt.Errorf("CoAP request failed: %w", err)
	}

	// Check response code
	if resp.Code() != codes.Content {
		return "", fmt.Errorf("CoAP server returned error code: %s", resp.Code())
	}

	// Read response body
	body, err := resp.ReadBody()
	if err != nil {
		return "", fmt.Errorf("failed to read CoAP response: %w", err)
	}

	return string(body), nil
}

// SendMessage sends an RPC message with authentication via CoAP
func (c *CoAPClientBackend) SendMessage(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
	// Create RPC message
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

	// Send via CoAP
	if ctx == nil {
		ctx = context.Background()
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Connect to CoAP server
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := udp.Dial(addr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to CoAP server: %w", err)
	}
	defer conn.Close()

	// Create POST request to /api endpoint
	resp, err := conn.Post(ctxWithTimeout, "/api", message.AppJSON, bytes.NewReader([]byte(jsonString)))
	if err != nil {
		return "", fmt.Errorf("CoAP request failed: %w", err)
	}

	// Check response code
	if resp.Code() != codes.Content {
		return "", fmt.Errorf("CoAP server returned error code: %s", resp.Code())
	}

	// Read response body
	body, err := resp.ReadBody()
	if err != nil {
		return "", fmt.Errorf("failed to read CoAP response: %w", err)
	}

	// Parse response
	rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(string(body))
	if err != nil {
		return "", errors.New("expected a valid Colonies RPC message, but got: " + string(body))
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

// CheckHealth checks the health of the CoAP server
func (c *CoAPClientBackend) CheckHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to CoAP server
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := udp.Dial(addr)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer conn.Close()

	// Send GET request to /health endpoint
	resp, err := conn.Get(ctx, "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Check response code
	if resp.Code() != codes.Content {
		return fmt.Errorf("server reported unhealthy status: %s", resp.Code())
	}

	return nil
}

// Close closes the CoAP connection and cleans up blueprints
func (c *CoAPClientBackend) Close() error {
	// CoAP UDP client doesn't maintain persistent connections
	return nil
}

// CoAPClientBackendFactory creates CoAP client backends
type CoAPClientBackendFactory struct{}

// NewCoAPClientBackendFactory creates a new CoAP client backend factory
func NewCoAPClientBackendFactory() *CoAPClientBackendFactory {
	return &CoAPClientBackendFactory{}
}

// CreateBackend creates a new CoAP client backend
func (f *CoAPClientBackendFactory) CreateBackend(config *backends.ClientConfig) (backends.ClientBackend, error) {
	return NewCoAPClientBackend(config)
}

// GetBackendType returns the backend type this factory creates
func (f *CoAPClientBackendFactory) GetBackendType() backends.ClientBackendType {
	return backends.CoAPClientBackendType
}

// Compile-time check that CoAPClientBackend implements the required interfaces
var _ backends.ClientBackend = (*CoAPClientBackend)(nil)
