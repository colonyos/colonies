package security

type Ownership interface {
	CheckIfColonyExists(colonyID string) error
	CheckIfRuntimeIsValid(runtimeID string, colonyID string) error
}
