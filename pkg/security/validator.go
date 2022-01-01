package security

type Validator interface {
	RequireServerOwner(recoveredID string, serverID string) error
	RequireColonyOwner(recoveredID string, colonyID string) error
	RequireRuntimeMembership(recoveredID string, colonyID string) error
}
