package validator

type ownership interface {
	resolveColony(colonyName string) (string, error)
	checkIfExecutorIsValid(executorID string, colonyID string, approved bool) error
	checkIfUserIsValid(userID string, colonyID string) error
}
