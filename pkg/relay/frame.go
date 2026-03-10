package relay

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	FrameTypeRequest  byte = 0x01
	FrameTypeResponse byte = 0x02

	// WebSocket frame types for multiplexed WS connections through the tunnel
	FrameTypeWSUpgrade   byte = 0x03 // Relay -> tunnel client: open WS to local server
	FrameTypeWSUpgradeOK byte = 0x04 // Tunnel client -> relay: WS opened successfully
	FrameTypeWSData      byte = 0x05 // Bidirectional: WS message data
	FrameTypeWSClose     byte = 0x06 // Either direction: WS connection closed

	// WebSocket message types encoded in the Method byte of WSData frames
	WSMsgTypeText   byte = 0x01
	WSMsgTypeBinary byte = 0x02

	MethodGET     byte = 0x01
	MethodPOST    byte = 0x02
	MethodPUT     byte = 0x03
	MethodDELETE  byte = 0x04
	MethodPATCH   byte = 0x05
	MethodHEAD    byte = 0x06
	MethodOPTIONS byte = 0x07

	// Fixed header size: 16 (UUID) + 1 (type) + 1 (method) + 2 (status) + 4 (path len) + 4 (headers len) + 4 (body len) = 32
	HeaderSize = 32

	// MaxFrameSize is the maximum allowed frame size (64 MB) to prevent memory exhaustion.
	MaxFrameSize = 64 * 1024 * 1024
)

type Frame struct {
	ID         string              // 16-byte UUID as hex string (32 chars)
	Type       byte                // FrameTypeRequest or FrameTypeResponse
	Method     byte                // HTTP method enum (for requests)
	StatusCode uint16              // HTTP status code (for responses)
	Path       string              // Request path
	Headers    map[string][]string // HTTP headers (supports multi-value, e.g. Set-Cookie)
	Body       []byte              // Raw body bytes
}

func MethodFromString(method string) byte {
	switch method {
	case "GET":
		return MethodGET
	case "POST":
		return MethodPOST
	case "PUT":
		return MethodPUT
	case "DELETE":
		return MethodDELETE
	case "PATCH":
		return MethodPATCH
	case "HEAD":
		return MethodHEAD
	case "OPTIONS":
		return MethodOPTIONS
	default:
		return MethodGET
	}
}

func MethodToString(method byte) string {
	switch method {
	case MethodGET:
		return "GET"
	case MethodPOST:
		return "POST"
	case MethodPUT:
		return "PUT"
	case MethodDELETE:
		return "DELETE"
	case MethodPATCH:
		return "PATCH"
	case MethodHEAD:
		return "HEAD"
	case MethodOPTIONS:
		return "OPTIONS"
	default:
		return "GET"
	}
}

// Marshal encodes a Frame into binary format.
func (f *Frame) Marshal() ([]byte, error) {
	idBytes, err := uuidToBytes(f.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid frame ID: %w", err)
	}

	pathBytes := []byte(f.Path)

	headersBytes, err := json.Marshal(f.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal headers: %w", err)
	}

	totalSize := HeaderSize + len(pathBytes) + len(headersBytes) + len(f.Body)
	buf := make([]byte, totalSize)

	copy(buf[0:16], idBytes)
	buf[16] = f.Type
	buf[17] = f.Method
	binary.BigEndian.PutUint16(buf[18:20], f.StatusCode)
	binary.BigEndian.PutUint32(buf[20:24], uint32(len(pathBytes)))
	binary.BigEndian.PutUint32(buf[24:28], uint32(len(headersBytes)))
	binary.BigEndian.PutUint32(buf[28:32], uint32(len(f.Body)))

	offset := HeaderSize
	copy(buf[offset:], pathBytes)
	offset += len(pathBytes)
	copy(buf[offset:], headersBytes)
	offset += len(headersBytes)
	copy(buf[offset:], f.Body)

	return buf, nil
}

// Unmarshal decodes binary data into a Frame.
func Unmarshal(data []byte) (*Frame, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("frame too short: %d bytes, minimum %d", len(data), HeaderSize)
	}
	if len(data) > MaxFrameSize {
		return nil, fmt.Errorf("frame too large: %d bytes, maximum %d", len(data), MaxFrameSize)
	}

	id := bytesToUUID(data[0:16])
	frameType := data[16]
	method := data[17]
	statusCode := binary.BigEndian.Uint16(data[18:20])
	pathLen := binary.BigEndian.Uint32(data[20:24])
	headersLen := binary.BigEndian.Uint32(data[24:28])
	bodyLen := binary.BigEndian.Uint32(data[28:32])

	expectedLen := HeaderSize + int(pathLen) + int(headersLen) + int(bodyLen)
	if len(data) < expectedLen {
		return nil, fmt.Errorf("frame data too short: got %d, expected %d", len(data), expectedLen)
	}

	offset := uint32(HeaderSize)
	path := string(data[offset : offset+pathLen])
	offset += pathLen

	var headers map[string][]string
	if headersLen > 0 {
		if err := json.Unmarshal(data[offset:offset+headersLen], &headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}
	}
	offset += headersLen

	body := make([]byte, bodyLen)
	copy(body, data[offset:offset+bodyLen])

	return &Frame{
		ID:         id,
		Type:       frameType,
		Method:     method,
		StatusCode: statusCode,
		Path:       path,
		Headers:    headers,
		Body:       body,
	}, nil
}

// uuidToBytes converts a hex UUID string (with or without dashes) to 16 bytes.
func uuidToBytes(id string) ([]byte, error) {
	clean := stripDashes(id)

	if len(clean) != 32 {
		return nil, fmt.Errorf("UUID must be 32 hex chars, got %d", len(clean))
	}

	return hex.DecodeString(clean)
}

// bytesToUUID converts 16 bytes to a hex UUID string (32 chars, no dashes).
func bytesToUUID(b []byte) string {
	return hex.EncodeToString(b)
}

// stripDashes removes dashes from a string (used for UUID normalization).
func stripDashes(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '-' {
			result = append(result, s[i])
		}
	}
	return string(result)
}
