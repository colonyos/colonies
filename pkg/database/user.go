package database

import "github.com/colonyos/colonies/pkg/core"

type UserDatabase interface {
	AddUser(user *core.User) error
	GetUsersByColonyName(colonyName string) ([]*core.User, error)
	GetUserByID(colonyName string, userID string) (*core.User, error)
	GetUserByName(colonyName string, name string) (*core.User, error)
	RemoveUserByID(colonyName string, userID string) error
	RemoveUserByName(colonyName string, name string) error
	RemoveUsersByColonyName(colonyName string) error
}