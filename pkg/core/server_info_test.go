package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateServerInfo(t *testing.T) {
	info := CreateServerInfo("1.0.0", "2024-01-01T12:00:00Z")

	assert.Equal(t, "1.0.0", info.BuildVersion)
	assert.Equal(t, "2024-01-01T12:00:00Z", info.BuildTime)
	assert.NotNil(t, info.Backends)
	assert.Len(t, info.Backends, 0)
}

func TestServerInfoAddBackend(t *testing.T) {
	info := CreateServerInfo("1.0.0", "2024-01-01T12:00:00Z")

	info.AddBackend("gin", 8080, "localhost", false, true)
	info.AddBackend("grpc", 9090, "localhost", true, false)

	assert.Len(t, info.Backends, 2)

	assert.Equal(t, "gin", info.Backends[0].Type)
	assert.Equal(t, 8080, info.Backends[0].Port)
	assert.Equal(t, "localhost", info.Backends[0].Host)
	assert.False(t, info.Backends[0].TLS)
	assert.True(t, info.Backends[0].Insecure)

	assert.Equal(t, "grpc", info.Backends[1].Type)
	assert.Equal(t, 9090, info.Backends[1].Port)
	assert.True(t, info.Backends[1].TLS)
	assert.False(t, info.Backends[1].Insecure)
}

func TestServerInfoToJSON(t *testing.T) {
	info := CreateServerInfo("1.0.0", "2024-01-01T12:00:00Z")
	info.AddBackend("gin", 8080, "localhost", false, true)

	jsonStr, err := info.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)
	assert.Contains(t, jsonStr, "1.0.0")
	assert.Contains(t, jsonStr, "gin")
}

func TestServerInfoToJSONIndent(t *testing.T) {
	info := CreateServerInfo("1.0.0", "2024-01-01T12:00:00Z")
	info.AddBackend("gin", 8080, "localhost", false, true)

	jsonStr, err := info.ToJSONIndent()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)
	assert.Contains(t, jsonStr, "1.0.0")
	assert.Contains(t, jsonStr, "\n") // Should have newlines for indentation
}

func TestCreateServerInfoFromJSON(t *testing.T) {
	original := CreateServerInfo("2.0.0", "2024-06-15T10:30:00Z")
	original.AddBackend("gin", 8080, "api.example.com", true, false)
	original.AddBackend("grpc", 9090, "grpc.example.com", true, false)

	jsonStr, err := original.ToJSON()
	assert.NoError(t, err)

	parsed, err := CreateServerInfoFromJSON(jsonStr)
	assert.NoError(t, err)

	assert.Equal(t, original.BuildVersion, parsed.BuildVersion)
	assert.Equal(t, original.BuildTime, parsed.BuildTime)
	assert.Len(t, parsed.Backends, 2)
	assert.Equal(t, original.Backends[0].Type, parsed.Backends[0].Type)
	assert.Equal(t, original.Backends[0].Port, parsed.Backends[0].Port)
}

func TestCreateServerInfoFromJSONError(t *testing.T) {
	_, err := CreateServerInfoFromJSON("invalid json")
	assert.Error(t, err)
}
