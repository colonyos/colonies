package security

type Ownership interface {
	CheckIfColonyExists(colonyID string) error
	CheckIfWorkerBelongsToColony(workerID string, colonyID string) error
}
