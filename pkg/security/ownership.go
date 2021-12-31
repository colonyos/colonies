package security

type Ownership interface {
	CheckIfColonyExists(colonyID string) error
	CheckIfRuntimeIsApproved(runtimeID string, colonyID string) error
}
