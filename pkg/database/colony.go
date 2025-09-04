package database

import "github.com/colonyos/colonies/pkg/core"

type ColonyDatabase interface {
	AddColony(colony *core.Colony) error
	GetColonies() ([]*core.Colony, error)
	GetColonyByID(id string) (*core.Colony, error)
	GetColonyByName(name string) (*core.Colony, error)
	RenameColony(colonyName string, newColonyName string) error
	RemoveColonyByName(colonyName string) error
	CountColonies() (int, error)
}