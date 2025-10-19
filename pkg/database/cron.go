package database

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

type CronDatabase interface {
	AddCron(cron *core.Cron) error
	UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, lastProcessGraphID string) error
	GetCronByID(cronID string) (*core.Cron, error)
	GetCronByName(colonyName string, cronName string) (*core.Cron, error)
	FindCronsByColonyName(colonyName string, count int) ([]*core.Cron, error)
	FindAllCrons() ([]*core.Cron, error)
	RemoveCronByID(cronID string) error
	RemoveAllCronsByColonyName(colonyName string) error
}