package backends

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockClientBackend implements ClientBackend for testing
type MockClientBackend struct {
	sendRawMessageFunc func(jsonString string, insecure bool) (string, error)
	sendMessageFunc    func(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error)
	checkHealthFunc    func() error
	closeFunc          func() error
}

func (m *MockClientBackend) SendRawMessage(jsonString string, insecure bool) (string, error) {
	if m.sendRawMessageFunc != nil {
		return m.sendRawMessageFunc(jsonString, insecure)
	}
	return "", nil
}

func (m *MockClientBackend) SendMessage(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(method, jsonString, prvKey, insecure, ctx)
	}
	return "", nil
}

func (m *MockClientBackend) CheckHealth() error {
	if m.checkHealthFunc != nil {
		return m.checkHealthFunc()
	}
	return nil
}

func (m *MockClientBackend) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// MockClientBackendFactory implements ClientBackendFactory for testing
type MockClientBackendFactory struct {
	backend     ClientBackend
	createError error
}

func (f *MockClientBackendFactory) CreateBackend(config *ClientConfig) (ClientBackend, error) {
	if f.createError != nil {
		return nil, f.createError
	}
	return f.backend, nil
}

func (f *MockClientBackendFactory) GetBackendType() ClientBackendType {
	return GinClientBackendType
}

// ============== interfaces.go tests ==============

func TestCreateDefaultClientConfig(t *testing.T) {
	config := CreateDefaultClientConfig("localhost", 8080, true, false)

	assert.NotNil(t, config)
	assert.Equal(t, GinClientBackendType, config.BackendType)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 8080, config.Port)
	assert.True(t, config.Insecure)
	assert.False(t, config.SkipTLSVerify)
}

func TestCreateDefaultClientConfigWithTLSSkip(t *testing.T) {
	config := CreateDefaultClientConfig("example.com", 443, false, true)

	assert.Equal(t, "example.com", config.Host)
	assert.Equal(t, 443, config.Port)
	assert.False(t, config.Insecure)
	assert.True(t, config.SkipTLSVerify)
}

// ============== multi_backend.go tests ==============

func TestNewMultiBackendClientNoConfigs(t *testing.T) {
	factories := make(map[ClientBackendType]ClientBackendFactory)

	_, err := NewMultiBackendClient([]*ClientConfig{}, factories)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one backend configuration required")
}

func TestNewMultiBackendClientNoFactory(t *testing.T) {
	configs := []*ClientConfig{
		{BackendType: GinClientBackendType, Host: "localhost", Port: 8080},
	}
	factories := make(map[ClientBackendType]ClientBackendFactory)

	// No factory registered for the backend type
	_, err := NewMultiBackendClient(configs, factories)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no backends could be initialized")
}

func TestNewMultiBackendClientFactoryError(t *testing.T) {
	configs := []*ClientConfig{
		{BackendType: GinClientBackendType, Host: "localhost", Port: 8080},
	}
	factories := map[ClientBackendType]ClientBackendFactory{
		GinClientBackendType: &MockClientBackendFactory{
			createError: errors.New("factory error"),
		},
	}

	_, err := NewMultiBackendClient(configs, factories)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no backends could be initialized")
}

func TestNewMultiBackendClientSuccess(t *testing.T) {
	mockBackend := &MockClientBackend{}
	configs := []*ClientConfig{
		{BackendType: GinClientBackendType, Host: "localhost", Port: 8080},
	}
	factories := map[ClientBackendType]ClientBackendFactory{
		GinClientBackendType: &MockClientBackendFactory{backend: mockBackend},
	}

	client, err := NewMultiBackendClient(configs, factories)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestMultiBackendClientSendRawMessageSuccess(t *testing.T) {
	mockBackend := &MockClientBackend{
		sendRawMessageFunc: func(jsonString string, insecure bool) (string, error) {
			return `{"result": "ok"}`, nil
		},
	}
	configs := []*ClientConfig{
		{BackendType: GinClientBackendType, Host: "localhost", Port: 8080},
	}
	factories := map[ClientBackendType]ClientBackendFactory{
		GinClientBackendType: &MockClientBackendFactory{backend: mockBackend},
	}

	client, _ := NewMultiBackendClient(configs, factories)
	result, err := client.SendRawMessage(`{"test": true}`, true)

	assert.NoError(t, err)
	assert.Equal(t, `{"result": "ok"}`, result)
}

func TestMultiBackendClientSendRawMessageFallback(t *testing.T) {
	failingBackend := &MockClientBackend{
		sendRawMessageFunc: func(jsonString string, insecure bool) (string, error) {
			return "", errors.New("primary failed")
		},
	}
	successBackend := &MockClientBackend{
		sendRawMessageFunc: func(jsonString string, insecure bool) (string, error) {
			return `{"fallback": true}`, nil
		},
	}

	// Create client with two backends manually
	client := &MultiBackendClient{
		backends: []ClientBackend{failingBackend, successBackend},
		configs: []*ClientConfig{
			{BackendType: GinClientBackendType},
			{BackendType: GinClientBackendType},
		},
	}

	result, err := client.SendRawMessage(`{}`, true)
	assert.NoError(t, err)
	assert.Equal(t, `{"fallback": true}`, result)
}

func TestMultiBackendClientSendRawMessageAllFail(t *testing.T) {
	failingBackend := &MockClientBackend{
		sendRawMessageFunc: func(jsonString string, insecure bool) (string, error) {
			return "", errors.New("backend failed")
		},
	}

	client := &MultiBackendClient{
		backends: []ClientBackend{failingBackend},
		configs:  []*ClientConfig{{BackendType: GinClientBackendType}},
	}

	_, err := client.SendRawMessage(`{}`, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all backends failed")
}

func TestMultiBackendClientSendMessageSuccess(t *testing.T) {
	mockBackend := &MockClientBackend{
		sendMessageFunc: func(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
			return `{"method": "` + method + `"}`, nil
		},
	}

	client := &MultiBackendClient{
		backends: []ClientBackend{mockBackend},
		configs:  []*ClientConfig{{BackendType: GinClientBackendType}},
	}

	result, err := client.SendMessage("POST", `{}`, "key", true, context.Background())
	assert.NoError(t, err)
	assert.Contains(t, result, "POST")
}

func TestMultiBackendClientSendMessageFallback(t *testing.T) {
	failingBackend := &MockClientBackend{
		sendMessageFunc: func(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
			return "", errors.New("primary failed")
		},
	}
	successBackend := &MockClientBackend{
		sendMessageFunc: func(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
			return `{"success": true}`, nil
		},
	}

	client := &MultiBackendClient{
		backends: []ClientBackend{failingBackend, successBackend},
		configs: []*ClientConfig{
			{BackendType: GinClientBackendType},
			{BackendType: GinClientBackendType},
		},
	}

	result, err := client.SendMessage("GET", `{}`, "key", false, context.Background())
	assert.NoError(t, err)
	assert.Equal(t, `{"success": true}`, result)
}

func TestMultiBackendClientSendMessageAllFail(t *testing.T) {
	failingBackend := &MockClientBackend{
		sendMessageFunc: func(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
			return "", errors.New("send failed")
		},
	}

	client := &MultiBackendClient{
		backends: []ClientBackend{failingBackend},
		configs:  []*ClientConfig{{BackendType: GinClientBackendType}},
	}

	_, err := client.SendMessage("POST", `{}`, "key", true, context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all backends failed")
}

func TestMultiBackendClientCheckHealthAllHealthy(t *testing.T) {
	healthyBackend := &MockClientBackend{
		checkHealthFunc: func() error { return nil },
	}

	client := &MultiBackendClient{
		backends: []ClientBackend{healthyBackend, healthyBackend},
		configs: []*ClientConfig{
			{BackendType: GinClientBackendType},
			{BackendType: GinClientBackendType},
		},
	}

	err := client.CheckHealth()
	assert.NoError(t, err)
}

func TestMultiBackendClientCheckHealthPartiallyHealthy(t *testing.T) {
	healthyBackend := &MockClientBackend{
		checkHealthFunc: func() error { return nil },
	}
	unhealthyBackend := &MockClientBackend{
		checkHealthFunc: func() error { return errors.New("unhealthy") },
	}

	client := &MultiBackendClient{
		backends: []ClientBackend{healthyBackend, unhealthyBackend},
		configs: []*ClientConfig{
			{BackendType: GinClientBackendType},
			{BackendType: GinClientBackendType},
		},
	}

	// Should not return error if at least one is healthy
	err := client.CheckHealth()
	assert.NoError(t, err)
}

func TestMultiBackendClientCheckHealthAllUnhealthy(t *testing.T) {
	unhealthyBackend := &MockClientBackend{
		checkHealthFunc: func() error { return errors.New("unhealthy") },
	}

	client := &MultiBackendClient{
		backends: []ClientBackend{unhealthyBackend},
		configs:  []*ClientConfig{{BackendType: GinClientBackendType}},
	}

	err := client.CheckHealth()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all backends unhealthy")
}

func TestMultiBackendClientCloseSuccess(t *testing.T) {
	mockBackend := &MockClientBackend{
		closeFunc: func() error { return nil },
	}

	client := &MultiBackendClient{
		backends: []ClientBackend{mockBackend},
		configs:  []*ClientConfig{{BackendType: GinClientBackendType}},
	}

	err := client.Close()
	assert.NoError(t, err)
}

func TestMultiBackendClientCloseWithErrors(t *testing.T) {
	failingBackend := &MockClientBackend{
		closeFunc: func() error { return errors.New("close failed") },
	}

	client := &MultiBackendClient{
		backends: []ClientBackend{failingBackend},
		configs:  []*ClientConfig{{BackendType: GinClientBackendType}},
	}

	err := client.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "errors closing backends")
}

// ============== ParseClientBackendsFromEnv tests ==============

func TestParseClientBackendsFromEnvEmpty(t *testing.T) {
	backends := ParseClientBackendsFromEnv("")
	assert.Len(t, backends, 1)
	assert.Equal(t, GinClientBackendType, backends[0])
}

func TestParseClientBackendsFromEnvHttp(t *testing.T) {
	backends := ParseClientBackendsFromEnv("http")
	assert.Len(t, backends, 1)
	assert.Equal(t, GinClientBackendType, backends[0])
}

func TestParseClientBackendsFromEnvGin(t *testing.T) {
	backends := ParseClientBackendsFromEnv("gin")
	assert.Len(t, backends, 1)
	assert.Equal(t, GinClientBackendType, backends[0])
}

func TestParseClientBackendsFromEnvMultiple(t *testing.T) {
	backends := ParseClientBackendsFromEnv("http, gin")
	assert.Len(t, backends, 2)
	assert.Equal(t, GinClientBackendType, backends[0])
	assert.Equal(t, GinClientBackendType, backends[1])
}

func TestParseClientBackendsFromEnvUnknown(t *testing.T) {
	backends := ParseClientBackendsFromEnv("unknown")
	// Should default to gin when no valid backends
	assert.Len(t, backends, 1)
	assert.Equal(t, GinClientBackendType, backends[0])
}

func TestParseClientBackendsFromEnvMixedValidInvalid(t *testing.T) {
	backends := ParseClientBackendsFromEnv("http, unknown, gin")
	// Should only include valid backends
	assert.Len(t, backends, 2)
}

func TestParseClientBackendsFromEnvCaseInsensitive(t *testing.T) {
	backends := ParseClientBackendsFromEnv("HTTP")
	assert.Len(t, backends, 1)
	assert.Equal(t, GinClientBackendType, backends[0])
}

func TestParseClientBackendsFromEnvWithSpaces(t *testing.T) {
	backends := ParseClientBackendsFromEnv("  http  ,  gin  ")
	assert.Len(t, backends, 2)
}

// ============== realtime.go constants tests ==============

func TestMessageTypeConstants(t *testing.T) {
	// Verify constants match expected WebSocket values
	assert.Equal(t, 1, TextMessage)
	assert.Equal(t, 2, BinaryMessage)
	assert.Equal(t, 8, CloseMessage)
	assert.Equal(t, 9, PingMessage)
	assert.Equal(t, 10, PongMessage)
}
