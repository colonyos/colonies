package backends

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// MultiBackendClient implements ClientBackend by trying multiple backends in order
type MultiBackendClient struct {
	backends []ClientBackend
	configs  []*ClientConfig
}

// NewMultiBackendClient creates a client that tries multiple backends with fallback
func NewMultiBackendClient(configs []*ClientConfig, factories map[ClientBackendType]ClientBackendFactory) (*MultiBackendClient, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("at least one backend configuration required")
	}

	client := &MultiBackendClient{
		backends: make([]ClientBackend, 0, len(configs)),
		configs:  configs,
	}

	// Create backends for each config
	for _, config := range configs {
		factory, exists := factories[config.BackendType]
		if !exists {
			logrus.WithField("backend_type", config.BackendType).Warn("No factory registered for backend type, skipping")
			continue
		}

		backend, err := factory.CreateBackend(config)
		if err != nil {
			logrus.WithError(err).WithField("backend_type", config.BackendType).Warn("Failed to create backend, skipping")
			continue
		}

		client.backends = append(client.backends, backend)
		logrus.WithField("backend_type", config.BackendType).Info("Backend initialized successfully")
	}

	if len(client.backends) == 0 {
		return nil, fmt.Errorf("no backends could be initialized")
	}

	logrus.WithField("backend_count", len(client.backends)).Info("Multi-backend client initialized")
	return client, nil
}

// SendRawMessage tries each backend in order until one succeeds
func (m *MultiBackendClient) SendRawMessage(jsonString string, insecure bool) (string, error) {
	var lastErr error

	for i, backend := range m.backends {
		result, err := backend.SendRawMessage(jsonString, insecure)
		if err == nil {
			if i > 0 {
				logrus.WithFields(logrus.Fields{
					"backend_type": m.configs[i].BackendType,
					"attempt":      i + 1,
				}).Debug("Request succeeded on fallback backend")
			}
			return result, nil
		}

		logrus.WithError(err).WithFields(logrus.Fields{
			"backend_type": m.configs[i].BackendType,
			"attempt":      i + 1,
		}).Debug("Backend request failed, trying next")

		lastErr = err
	}

	return "", fmt.Errorf("all backends failed, last error: %w", lastErr)
}

// SendMessage tries each backend in order until one succeeds
func (m *MultiBackendClient) SendMessage(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
	var lastErr error

	for i, backend := range m.backends {
		result, err := backend.SendMessage(method, jsonString, prvKey, insecure, ctx)
		if err == nil {
			if i > 0 {
				logrus.WithFields(logrus.Fields{
					"backend_type": m.configs[i].BackendType,
					"attempt":      i + 1,
					"method":       method,
				}).Debug("Request succeeded on fallback backend")
			}
			return result, nil
		}

		logrus.WithError(err).WithFields(logrus.Fields{
			"backend_type": m.configs[i].BackendType,
			"attempt":      i + 1,
			"method":       method,
		}).Debug("Backend request failed, trying next")

		lastErr = err
	}

	return "", fmt.Errorf("all backends failed, last error: %w", lastErr)
}

// CheckHealth checks health of all backends and returns error if all are unhealthy
func (m *MultiBackendClient) CheckHealth() error {
	var errors []string
	healthyCount := 0

	for i, backend := range m.backends {
		err := backend.CheckHealth()
		if err == nil {
			healthyCount++
		} else {
			errors = append(errors, fmt.Sprintf("%s: %v", m.configs[i].BackendType, err))
		}
	}

	if healthyCount == 0 {
		return fmt.Errorf("all backends unhealthy: %s", strings.Join(errors, "; "))
	}

	if healthyCount < len(m.backends) {
		logrus.WithFields(logrus.Fields{
			"healthy":   healthyCount,
			"total":     len(m.backends),
			"degraded":  errors,
		}).Warn("Some backends are unhealthy")
	}

	return nil
}

// Close closes all backends
func (m *MultiBackendClient) Close() error {
	var errors []string

	for i, backend := range m.backends {
		if err := backend.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", m.configs[i].BackendType, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing backends: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ParseClientBackendsFromEnv parses comma-separated backend types from environment variable
// e.g., "libp2p,http" or "http" or "libp2p"
func ParseClientBackendsFromEnv(backendsEnv string) []ClientBackendType {
	if backendsEnv == "" {
		return []ClientBackendType{GinClientBackendType} // Default to HTTP
	}

	parts := strings.Split(backendsEnv, ",")
	backends := make([]ClientBackendType, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		switch part {
		case "http", "gin":
			backends = append(backends, GinClientBackendType)
		case "grpc":
			backends = append(backends, GRPCClientBackendType)
		case "libp2p", "p2p":
			backends = append(backends, LibP2PClientBackendType)
		case "coap":
			backends = append(backends, CoAPClientBackendType)
		default:
			logrus.WithField("backend", part).Warn("Unknown backend type, ignoring")
		}
	}

	if len(backends) == 0 {
		logrus.Warn("No valid backends specified, defaulting to HTTP")
		return []ClientBackendType{GinClientBackendType}
	}

	return backends
}
