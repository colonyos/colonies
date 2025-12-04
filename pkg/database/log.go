package database

import "github.com/colonyos/colonies/pkg/core"

type LogDatabase interface {
	AddLog(processID string, colonyName string, executorName string, timestamp int64, msg string) error
	GetLogsByProcessID(processID string, limit int) ([]*core.Log, error)
	GetLogsByProcessIDSince(processID string, limit int, since int64) ([]*core.Log, error)
	GetLogsByProcessIDLatest(processID string, limit int) ([]*core.Log, error)
	GetLogsByExecutor(executorName string, limit int) ([]*core.Log, error)
	GetLogsByExecutorSince(executorName string, limit int, since int64) ([]*core.Log, error)
	GetLogsByExecutorLatest(executorName string, limit int) ([]*core.Log, error)
	RemoveLogsByColonyName(colonyName string) error
	CountLogs(colonyName string) (int, error)
	SearchLogs(colonyName string, text string, days int, count int) ([]*core.Log, error)
}