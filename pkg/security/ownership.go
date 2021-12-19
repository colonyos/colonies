package security

type Ownership interface {
	CheckIfColonyExists(colonyID string) error
	CheckIfComputerBelongsToColony(computerID string, colonyID string) error
}
