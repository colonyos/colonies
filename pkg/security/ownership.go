package security

type Ownership interface {
	CheckIfColonyExists(colonyID string) error
	CheckIfColonyHasWorker(colonyID string, workerID string) error
}
