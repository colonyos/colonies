package validator

type ownership interface {
	checkIfColonyExists(colonyID string) error
	checkIfRuntimeIsValid(runtimeID string, colonyID string) error
}
