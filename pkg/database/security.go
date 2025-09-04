package database

type SecurityDatabase interface {
	SetServerID(oldServerID, newServerID string) error
	GetServerID() (string, error)
	ChangeColonyID(colonyName string, oldColonyID, newColonyID string) error
	ChangeUserID(colonyName string, oldUserID, newUserID string) error
	ChangeExecutorID(colonyName string, oldExecutorID, newExecutorID string) error
}