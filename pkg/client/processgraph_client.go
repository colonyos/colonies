package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) SubmitWorkflowSpec(workflowSpec *core.WorkflowSpec, prvKey string) (*core.ProcessGraph, error) {
	msg := rpc.CreateSubmitWorkflowSpecMsg(workflowSpec)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.SubmitWorkflowSpecPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessGraph(respBodyString)
}

func (client *ColoniesClient) AddChild(processGraphID string, parentProcessID string, childProcessID string, funcSpec *core.FunctionSpec, insert bool, prvKey string) (*core.Process, error) {
	msg := rpc.CreateAddChildMsg(processGraphID, parentProcessID, childProcessID, funcSpec, insert)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddChildPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) GetProcessGraph(processGraphID string, prvKey string) (*core.ProcessGraph, error) {
	msg := rpc.CreateGetProcessGraphMsg(processGraphID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetProcessGraphPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessGraph(respBodyString)
}

func (client *ColoniesClient) getProcessGraphs(state int, colonyName string, count int, prvKey string) ([]*core.ProcessGraph, error) {
	msg := rpc.CreateGetProcessGraphsMsg(colonyName, count, state)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetProcessGraphsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessGraphArray(respBodyString)
}

func (client *ColoniesClient) GetWaitingProcessGraphs(colonyName string, count int, prvKey string) ([]*core.ProcessGraph, error) {
	return client.getProcessGraphs(core.WAITING, colonyName, count, prvKey)
}

func (client *ColoniesClient) GetRunningProcessGraphs(colonyName string, count int, prvKey string) ([]*core.ProcessGraph, error) {
	return client.getProcessGraphs(core.RUNNING, colonyName, count, prvKey)
}

func (client *ColoniesClient) GetSuccessfulProcessGraphs(colonyName string, count int, prvKey string) ([]*core.ProcessGraph, error) {
	return client.getProcessGraphs(core.SUCCESS, colonyName, count, prvKey)
}

func (client *ColoniesClient) GetFailedProcessGraphs(colonyName string, count int, prvKey string) ([]*core.ProcessGraph, error) {
	return client.getProcessGraphs(core.FAILED, colonyName, count, prvKey)
}

func (client *ColoniesClient) RemoveProcessGraph(processGraphID string, prvKey string) error {
	msg := rpc.CreateRemoveProcessGraphMsg(processGraphID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveProcessGraphPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RemoveAllProcessGraphs(colonyName string, prvKey string) error {
	msg := rpc.CreateRemoveAllProcessGraphsMsg(colonyName)
	msg.State = core.NOTSET
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveAllProcessGraphsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RemoveAllProcessGraphsWithState(colonyName string, state int, prvKey string) error {
	msg := rpc.CreateRemoveAllProcessGraphsMsg(colonyName)
	msg.State = state
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveAllProcessGraphsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}