package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseServerBackendsFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		backendsEnv string
		expected    BackendType
	}{
		{
			name:        "Empty string defaults to HTTP",
			backendsEnv: "",
			expected:    GinBackendType,
		},
		{
			name:        "HTTP only",
			backendsEnv: "http",
			expected:    GinBackendType,
		},
		{
			name:        "Gin only (alias for HTTP)",
			backendsEnv: "gin",
			expected:    GinBackendType,
		},
		{
			name:        "LibP2P only",
			backendsEnv: "libp2p",
			expected:    LibP2PBackendType,
		},
		{
			name:        "P2P only (alias for LibP2P)",
			backendsEnv: "p2p",
			expected:    LibP2PBackendType,
		},
		{
			name:        "HTTP and LibP2P (comma-separated)",
			backendsEnv: "http,libp2p",
			expected:    LibP2PBackendType, // LibP2P backend runs both
		},
		{
			name:        "LibP2P and HTTP (comma-separated, order doesn't matter)",
			backendsEnv: "libp2p,http",
			expected:    LibP2PBackendType, // LibP2P backend runs both
		},
		{
			name:        "HTTP with whitespace",
			backendsEnv: " http ",
			expected:    GinBackendType,
		},
		{
			name:        "LibP2P and HTTP with whitespace",
			backendsEnv: " libp2p , http ",
			expected:    LibP2PBackendType,
		},
		{
			name:        "Case insensitive - HTTP",
			backendsEnv: "HTTP",
			expected:    GinBackendType,
		},
		{
			name:        "Case insensitive - LibP2P",
			backendsEnv: "LIBP2P",
			expected:    LibP2PBackendType,
		},
		{
			name:        "Unknown backend defaults to HTTP",
			backendsEnv: "unknown",
			expected:    GinBackendType,
		},
		{
			name:        "Mixed valid and invalid",
			backendsEnv: "http,invalid",
			expected:    GinBackendType,
		},
		{
			name:        "LibP2P with invalid",
			backendsEnv: "libp2p,invalid",
			expected:    LibP2PBackendType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseServerBackendsFromEnv(tt.backendsEnv)
			assert.Equal(t, tt.expected, result, "Expected backend type %v but got %v for input: %q", tt.expected, result, tt.backendsEnv)
		})
	}
}

func TestParseServerBackendsFromEnv_HTTPOnly(t *testing.T) {
	// Test that "http" only returns GinBackendType (HTTP only, no LibP2P)
	result := ParseServerBackendsFromEnv("http")
	assert.Equal(t, GinBackendType, result, "Setting COLONIES_SERVER_BACKENDS=http should enable HTTP backend only")
}

func TestParseServerBackendsFromEnv_LibP2POnly(t *testing.T) {
	// Test that "libp2p" returns LibP2PBackendType (runs both HTTP and LibP2P)
	result := ParseServerBackendsFromEnv("libp2p")
	assert.Equal(t, LibP2PBackendType, result, "Setting COLONIES_SERVER_BACKENDS=libp2p should enable LibP2P backend (with HTTP compatibility)")
}

func TestParseServerBackendsFromEnv_Both(t *testing.T) {
	// Test that "http,libp2p" returns LibP2PBackendType (runs both)
	result := ParseServerBackendsFromEnv("http,libp2p")
	assert.Equal(t, LibP2PBackendType, result, "Setting COLONIES_SERVER_BACKENDS=http,libp2p should enable LibP2P backend (with HTTP compatibility)")
}
