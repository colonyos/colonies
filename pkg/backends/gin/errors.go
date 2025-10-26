package gin

import "errors"

var (
	ErrConnectionClosed = errors.New("websocket connection is closed")
	ErrInvalidConnType  = errors.New("invalid connection type for websocket implementation")
)