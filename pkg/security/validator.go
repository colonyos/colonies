package security

type Validator interface {
	RequireServerOwner(recoveredID string, serverID string) error
	RequireColonyOwner(recoveredID string, colonyID string) error
	RequireExecutorMembership(recoveredID string, colonyID string, approved bool) error
}
