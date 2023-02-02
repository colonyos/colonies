package validator

type ownership interface {
	checkIfColonyExists(colonyID string) error
	checkIfExecutorIsValid(executorID string, colonyID string, approved bool) error
}
