package database

import "colonies/pkg/core"

type Database interface {
	// Colony functions ...
	AddColony(colony *core.Colony) error
	GetColonies() ([]*core.Colony, error)
	GetColonyByID(id string) (*core.Colony, error)
	DeleteColonyByID(colonyID string) error

	// Worker functions ...
	AddWorker(worker *core.Worker) error
	GetWorkers() ([]*core.Worker, error)
	GetWorkerByID(workerID string) (*core.Worker, error)
	GetWorkersByColonyID(workerID string) ([]*core.Worker, error)
	ApproveWorker(worker *core.Worker) error
	RejectWorker(worker *core.Worker) error
	DeleteWorkerByID(workerID string) error
	DeleteWorkersByColonyID(colonyID string) error

	// Task functions ...
	AddTask(task *core.Task)
	GetTasks() ([]*core.Task, error)
	GetTaskByID(taskID string) (*core.Task, error)
	SearchTasks(colonyID string, workerID string) ([]*core.Task, error)
	DeleteTaskByID(taskID string) error
	DeleteAllTasks() error
	ResetTask(task *core.Task)
	ResetAllTasks(task *core.Task)
	AssignWorker(workerID string, task *core.Task)
	UnassignWorker(task *core.Task)
	MarkSuccessful(task *core.Task) error
	MarkFailed(task *core.Task) error
	NumberOfTasks() (int, error)
	NumberOfRunningTasks()
	NumberOfSuccessfulTasks()
	NumberOfFailedTasks()

	// Attribute functions
	AddAttribute(attribute *core.Attribute) error
	GetAttributeByID(attributeID string) (*core.Attribute, error)
	GetAttribute(taskID string, key string, attributeType int) (*core.Attribute, error)
	GetAttributes(taskID string, attributeType int) ([]*core.Attribute, error)
	UpdateAttribute(attribute *core.Attribute) error
	DeleteAttributeByID(attributeID string)
	DeleteAttributesByTaskID(taskID string, attributeType int)
	DeleteAllAttributesByTaskID(taskID string) error
	DeleteAllAttributes() error
}
