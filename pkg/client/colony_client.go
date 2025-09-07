package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) AddColony(colony *core.Colony, prvKey string) (*core.Colony, error) {
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddColonyPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	addedColony, err := core.ConvertJSONToColony(respBodyString)
	if err != nil {
		return nil, err
	}

	return addedColony, nil
}

func (client *ColoniesClient) RemoveColony(colonyName string, prvKey string) error {
	msg := rpc.CreateRemoveColonyMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveColonyPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) GetColonies(prvKey string) ([]*core.Colony, error) {
	msg := rpc.CreateGetColoniesMsg()
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetColoniesPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToColonyArray(respBodyString)
}

func (client *ColoniesClient) GetColonyByName(colonyName string, prvKey string) (*core.Colony, error) {
	msg := rpc.CreateGetColonyMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetColonyPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToColony(respBodyString)
}

func (client *ColoniesClient) ColonyStatistics(colonyName string, prvKey string) (*core.Statistics, error) {
	msg := rpc.CreateGetColonyStatisticsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetColonyStatisticsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToStatistics(respBodyString)
}

func (client *ColoniesClient) ChangeColonyID(colonyName, colonyID string, prvKey string) error {
	msg := rpc.CreateChangeColonyIDMsg(colonyName, colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.ChangeColonyIDPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) PauseColonyAssignments(colonyName string, prvKey string) error {
	msg := rpc.CreatePauseAssignmentsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.PauseAssignmentsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) ResumeColonyAssignments(colonyName string, prvKey string) error {
	msg := rpc.CreateResumeAssignmentsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.ResumeAssignmentsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) AreColonyAssignmentsPaused(colonyName string, prvKey string) (bool, error) {
	msg := rpc.CreateGetPauseStatusMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return false, err
	}

	replyString, err := client.sendMessage(rpc.GetPauseStatusPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return false, err
	}

	reply, err := rpc.CreatePauseStatusReplyMsgFromJSON(replyString)
	if err != nil {
		return false, err
	}

	return reply.IsPaused, nil
}