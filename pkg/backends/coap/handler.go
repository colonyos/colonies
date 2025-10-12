package coap

// RPCHandler is an interface for handling RPC messages
// This is implemented by the Server struct which provides access to all business logic
type RPCHandler interface {
	HandleRPC(jsonPayload string) (string, error)
}
