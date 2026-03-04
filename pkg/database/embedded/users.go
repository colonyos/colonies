package embedded

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

func copyUser(u *core.User) *core.User {
	cp := *u
	return &cp
}

func (db *EmbeddedDatabase) AddUser(user *core.User) error {
	if user == nil {
		return errors.New("User is nil")
	}

	existing, err := db.GetUserByName(user.ColonyName, user.Name)
	if err != nil {
		return err
	}

	if existing != nil {
		return errors.New("User with name <" + user.Name + "> already exists in Colony with name <" + user.ColonyName + ">")
	}

	cp := copyUser(user)
	key := cp.ColonyName + ":" + cp.Name
	if err := db.users.Put(key, cp); err != nil {
		return err
	}

	db.usersIdx.byColony.Add(key, cp.ColonyName)
	db.usersIdx.byID.Add(key, cp.ID)

	return nil
}

func (db *EmbeddedDatabase) GetUsersByColonyName(colonyName string) ([]*core.User, error) {
	keys := db.usersIdx.byColony.Lookup(colonyName)
	var users []*core.User
	for _, key := range keys {
		if u, ok := db.users.Get(key); ok {
			users = append(users, copyUser(u))
		}
	}
	return users, nil
}

func (db *EmbeddedDatabase) GetUserByID(colonyName string, userID string) (*core.User, error) {
	keys := db.usersIdx.byID.Lookup(userID)
	for _, key := range keys {
		if u, ok := db.users.Get(key); ok {
			if u.ColonyName == colonyName {
				return copyUser(u), nil
			}
		}
	}
	return nil, nil
}

func (db *EmbeddedDatabase) GetUserByName(colonyName string, name string) (*core.User, error) {
	key := colonyName + ":" + name
	u, ok := db.users.Get(key)
	if !ok {
		return nil, nil
	}
	return copyUser(u), nil
}

func (db *EmbeddedDatabase) RemoveUserByID(colonyName string, userID string) error {
	keys := db.usersIdx.byID.Lookup(userID)
	for _, key := range keys {
		if u, ok := db.users.Get(key); ok {
			if u.ColonyName == colonyName {
				db.usersIdx.byColony.Remove(key, u.ColonyName)
				db.usersIdx.byID.Remove(key, u.ID)
				return db.users.Delete(key)
			}
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveUserByName(colonyName string, name string) error {
	key := colonyName + ":" + name
	u, ok := db.users.Get(key)
	if !ok {
		return nil
	}

	db.usersIdx.byColony.Remove(key, u.ColonyName)
	db.usersIdx.byID.Remove(key, u.ID)
	return db.users.Delete(key)
}

func (db *EmbeddedDatabase) RemoveUsersByColonyName(colonyName string) error {
	keys := db.usersIdx.byColony.Lookup(colonyName)
	for _, key := range keys {
		if u, ok := db.users.Get(key); ok {
			db.usersIdx.byID.Remove(key, u.ID)
		}
		if err := db.users.Delete(key); err != nil {
			return err
		}
	}
	// Clear the colony index entry
	for _, key := range keys {
		db.usersIdx.byColony.Remove(key, colonyName)
	}
	return nil
}
