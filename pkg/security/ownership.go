package security

type Ownership interface {
	CheckIfColonyExists(colonyID string) error
	CheckIfRuntimeBelongsToColony(runtimeID string, colonyID string) error
}
