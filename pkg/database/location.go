package database

import "github.com/colonyos/colonies/pkg/core"

type LocationDatabase interface {
	AddLocation(location *core.Location) error
	GetLocationsByColonyName(colonyName string) ([]*core.Location, error)
	GetLocationByID(locationID string) (*core.Location, error)
	GetLocationByName(colonyName string, name string) (*core.Location, error)
	RemoveLocationByID(locationID string) error
	RemoveLocationByName(colonyName string, name string) error
	RemoveLocationsByColonyName(colonyName string) error
}
