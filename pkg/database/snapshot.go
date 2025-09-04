package database

import "github.com/colonyos/colonies/pkg/core"

type SnapshotDatabase interface {
	CreateSnapshot(colonyName string, label string, name string) (*core.Snapshot, error)
	GetSnapshotByID(colonNamey string, snapshotID string) (*core.Snapshot, error)
	GetSnapshotsByColonyName(colonyName string) ([]*core.Snapshot, error)
	RemoveSnapshotByID(colonyName string, snapshotID string) error
	GetSnapshotByName(colonyName string, name string) (*core.Snapshot, error)
	RemoveSnapshotByName(colonyName string, name string) error
	RemoveSnapshotsByColonyName(colonyName string) error
}