package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) AddUser(user *core.User, prvKey string) (*core.User, error) {
	msg := rpc.CreateAddUserMsg(user)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddUserPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	addedUser, err := core.ConvertJSONToUser(respBodyString)
	if err != nil {
		return nil, err
	}

	return addedUser, nil
}

func (client *ColoniesClient) GetUser(colonyName string, username string, prvKey string) (*core.User, error) {
	msg := rpc.CreateGetUserMsg(colonyName, username)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetUserPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	userFromServer, err := core.ConvertJSONToUser(respBodyString)
	if err != nil {
		return nil, err
	}

	return userFromServer, nil
}

func (client *ColoniesClient) GetUserByID(colonyName string, userID string, prvKey string) (*core.User, error) {
	msg := rpc.CreateGetUserByIDMsg(colonyName, userID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetUserByIDPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	userFromServer, err := core.ConvertJSONToUser(respBodyString)
	if err != nil {
		return nil, err
	}

	return userFromServer, nil
}

func (client *ColoniesClient) GetUsers(colonyName string, prvKey string) ([]*core.User, error) {
	msg := rpc.CreateGetUsersMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetUsersPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	usersFromServer, err := core.ConvertJSONToUserArray(respBodyString)
	if err != nil {
		return nil, err
	}

	return usersFromServer, nil
}

func (client *ColoniesClient) RemoveUser(colonyName string, username string, prvKey string) error {
	msg := rpc.CreateRemoveUserMsg(colonyName, username)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveUserPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) ChangeUserID(colonyName, userID string, prvKey string) error {
	msg := rpc.CreateChangeUserIDMsg(colonyName, userID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.ChangeUserIDPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}