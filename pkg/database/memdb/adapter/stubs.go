package adapter

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// This file contains stub implementations for the remaining database interfaces
// These are simplified implementations for interfaces that are less commonly used

// CronDatabase interface implementation (minimal stubs)
func (a *ColonyOSAdapter) AddCron(cron *core.Cron) error                                                                      { return nil }
func (a *ColonyOSAdapter) GetCrons(colonyName string) ([]*core.Cron, error)                                                  { return nil, nil }
func (a *ColonyOSAdapter) GetCronByID(cronID string) (*core.Cron, error)                                                     { return nil, nil }
func (a *ColonyOSAdapter) GetCronByName(colonyName, cronName string) (*core.Cron, error)                                     { return nil, nil }
func (a *ColonyOSAdapter) UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, lastProcessGraphID string) error { return nil }
func (a *ColonyOSAdapter) FindCronsByColonyName(colonyName string, count int) ([]*core.Cron, error)                         { return nil, nil }
func (a *ColonyOSAdapter) FindAllCrons() ([]*core.Cron, error)                                                               { return nil, nil }
func (a *ColonyOSAdapter) RemoveCronByID(cronID string) error                                                                 { return nil }
func (a *ColonyOSAdapter) RemoveCronByName(colonyName, cronName string) error                                                 { return nil }
func (a *ColonyOSAdapter) RemoveAllCronsByColonyName(colonyName string) error                                                 { return nil }
func (a *ColonyOSAdapter) CountCronsByColonyName(colonyName string) (int, error)                                              { return 0, nil }

// SnapshotDatabase interface implementation (minimal stubs)
func (a *ColonyOSAdapter) AddSnapshot(snapshot *core.Snapshot) error                                                              { return nil }
func (a *ColonyOSAdapter) CreateSnapshot(colonyName string, label string, name string) (*core.Snapshot, error)                  { return nil, nil }
func (a *ColonyOSAdapter) GetSnapshots(colonyName string) ([]*core.Snapshot, error)                                              { return nil, nil }
func (a *ColonyOSAdapter) GetSnapshotByID(colonyName string, snapshotID string) (*core.Snapshot, error)                         { return nil, nil }
func (a *ColonyOSAdapter) GetSnapshotsByColonyName(colonyName string) ([]*core.Snapshot, error)                                  { return nil, nil }
func (a *ColonyOSAdapter) GetSnapshotByName(colonyName string, name string) (*core.Snapshot, error)                              { return nil, nil }
func (a *ColonyOSAdapter) RemoveSnapshotByID(colonyName string, snapshotID string) error                                          { return nil }
func (a *ColonyOSAdapter) RemoveSnapshotByName(colonyName string, name string) error                                              { return nil }
func (a *ColonyOSAdapter) RemoveSnapshotsByColonyName(colonyName string) error                                                    { return nil }
func (a *ColonyOSAdapter) RemoveAllSnapshotsByColonyName(colonyName string) error                                                 { return nil }
func (a *ColonyOSAdapter) CountSnapshotsByColonyName(colonyName string) (int, error)                                              { return 0, nil }

// SecurityDatabase interface implementation (minimal stubs)
func (a *ColonyOSAdapter) SetServerID(oldServerID, newServerID string) error          { return nil }
func (a *ColonyOSAdapter) GetServerID() (string, error)                               { return "", nil }
func (a *ColonyOSAdapter) ChangeColonyID(colonyName string, oldColonyID, newColonyID string) error { return nil }
func (a *ColonyOSAdapter) ChangeUserID(colonyName string, oldUserID, newUserID string) error       { return nil }
func (a *ColonyOSAdapter) ChangeExecutorID(colonyName string, oldExecutorID, newExecutorID string) error { return nil }