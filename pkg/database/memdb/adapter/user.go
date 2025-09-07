package adapter

import (
	"context"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// UserDatabase interface implementation

func (a *ColonyOSAdapter) AddUser(user *core.User) error {
	doc := &memdb.VelocityDocument{
		ID:     user.ID,
		Fields: a.userToFields(user),
	}
	
	return a.db.Insert(context.Background(), UsersCollection, doc)
}

func (a *ColonyOSAdapter) GetUsers() ([]*core.User, error) {
	result, err := a.db.List(context.Background(), UsersCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	users := make([]*core.User, 0, len(result))
	for _, doc := range result {
		user, err := a.fieldsToUser(doc.Fields)
		if err == nil {
			users = append(users, user)
		}
	}
	
	return users, nil
}

func (a *ColonyOSAdapter) GetUsersByColonyName(colonyName string) ([]*core.User, error) {
	users, err := a.GetUsers()
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.User
	for _, user := range users {
		if user.ColonyName == colonyName {
			filtered = append(filtered, user)
		}
	}
	
	return filtered, nil
}

func (a *ColonyOSAdapter) GetUserByID(colonyName string, userID string) (*core.User, error) {
	doc, err := a.db.Get(context.Background(), UsersCollection, userID)
	if err != nil {
		return nil, err
	}
	
	user, err := a.fieldsToUser(doc.Fields)
	if err != nil {
		return nil, err
	}
	
	// Check if user belongs to the specified colony
	if user.ColonyName != colonyName {
		return nil, fmt.Errorf("user not found in colony")
	}
	
	return user, nil
}

func (a *ColonyOSAdapter) GetUserByName(colonyName string, name string) (*core.User, error) {
	users, err := a.GetUsersByColonyName(colonyName)
	if err != nil {
		return nil, err
	}
	
	for _, user := range users {
		if user.Name == name {
			return user, nil
		}
	}
	
	return nil, fmt.Errorf("user not found")
}

func (a *ColonyOSAdapter) RemoveUserByID(colonyName string, userID string) error {
	// Verify user exists in the specified colony
	_, err := a.GetUserByID(colonyName, userID)
	if err != nil {
		return err
	}
	
	return a.db.Delete(context.Background(), UsersCollection, userID)
}

func (a *ColonyOSAdapter) RemoveUserByName(colonyName string, name string) error {
	user, err := a.GetUserByName(colonyName, name)
	if err != nil {
		return err
	}
	
	return a.db.Delete(context.Background(), UsersCollection, user.ID)
}

func (a *ColonyOSAdapter) RemoveUsersByColonyName(colonyName string) error {
	users, err := a.GetUsersByColonyName(colonyName)
	if err != nil {
		return err
	}
	
	for _, user := range users {
		if err := a.db.Delete(context.Background(), UsersCollection, user.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) CountUsers() (int, error) {
	result, err := a.db.List(context.Background(), UsersCollection, 10000, 0)
	if err != nil {
		return 0, err
	}
	return len(result), nil
}

// Conversion helper methods

func (a *ColonyOSAdapter) userToFields(user *core.User) map[string]interface{} {
	return map[string]interface{}{
		"id":           user.ID,
		"name":         user.Name,
		"colony_name":  user.ColonyName,
		"email":        user.Email,
		"phone":        user.Phone,
	}
}

func (a *ColonyOSAdapter) fieldsToUser(fields map[string]interface{}) (*core.User, error) {
	user := &core.User{}
	
	if id, ok := fields["id"].(string); ok {
		user.ID = id
	}
	if name, ok := fields["name"].(string); ok {
		user.Name = name
	}
	if colonyName, ok := fields["colony_name"].(string); ok {
		user.ColonyName = colonyName
	}
	if email, ok := fields["email"].(string); ok {
		user.Email = email
	}
	if phone, ok := fields["phone"].(string); ok {
		user.Phone = phone
	}
	
	return user, nil
}