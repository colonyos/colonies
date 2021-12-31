package security

type Validator interface {
	RequireRoot(rootPassword string, expectedRootPassword string) error
	RequireColonyOwner(recoveredID string, colonyID string) error
	RequireRuntimeMembership(runtimeID string, colonyID string) error
}
