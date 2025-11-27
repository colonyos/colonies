package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) AddExecutor(executor *core.Executor, prvKey string) (*core.Executor, error) {
	msg := rpc.CreateAddExecutorMsg(executor)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddExecutorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToExecutor(respBodyString)
}

func (client *ColoniesClient) ReportAllocation(colonyName string, executorName string, alloc core.Allocations, prvKey string) error {
	msg := rpc.CreateReportAllocationsMsg(colonyName, executorName, alloc)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.ReportAllocationsPayloadType, jsonString, prvKey, false, context.TODO())

	return err
}

func (client *ColoniesClient) GetExecutors(colonyName string, prvKey string) ([]*core.Executor, error) {
	msg := rpc.CreateGetExecutorsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetExecutorsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToExecutorArray(respBodyString)
}

func (client *ColoniesClient) GetExecutor(colonyName string, executorName string, prvKey string) (*core.Executor, error) {
	msg := rpc.CreateGetExecutorMsg(colonyName, executorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetExecutorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToExecutor(respBodyString)
}

func (client *ColoniesClient) GetExecutorByID(colonyName string, executorID string, prvKey string) (*core.Executor, error) {
	msg := rpc.CreateGetExecutorByIDMsg(colonyName, executorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetExecutorByIDPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToExecutor(respBodyString)
}

func (client *ColoniesClient) ApproveExecutor(colonyName string, executorName string, prvKey string) error {
	msg := rpc.CreateApproveExecutorMsg(colonyName, executorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.ApproveExecutorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RejectExecutor(colonyName string, executorID string, prvKey string) error {
	msg := rpc.CreateRejectExecutorMsg(colonyName, executorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RejectExecutorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RemoveExecutor(colonyName string, executorName string, prvKey string) error {
	msg := rpc.CreateRemoveExecutorMsg(colonyName, executorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveExecutorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) GetProcessHistForExecutor(state int, colonyName string, executorID string, seconds int, prvKey string) ([]*core.Process, error) {
	msg := rpc.CreateGetProcessHistMsg(colonyName, executorID, seconds, state)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetProcessHistPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessArray(respBodyString)
}

func (client *ColoniesClient) GetLogsByExecutor(colonyName, executorName string, count int, prvKey string) ([]*core.Log, error) {
	msg := rpc.CreateGetLogsMsg(colonyName, "", count, 0)
	msg.ExecutorName = executorName
	jsonString, err := msg.ToJSON()
	if err != nil {
		return []*core.Log{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLogsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return []*core.Log{}, err
	}

	return core.ConvertJSONToLogArray(respBodyString)
}

func (client *ColoniesClient) GetLogsByExecutorSince(colonyName, executorName string, count int, since int64, prvKey string) ([]*core.Log, error) {
	msg := rpc.CreateGetLogsMsg(colonyName, "", count, since)
	msg.ExecutorName = executorName
	jsonString, err := msg.ToJSON()
	if err != nil {
		return []*core.Log{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLogsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return []*core.Log{}, err
	}

	return core.ConvertJSONToLogArray(respBodyString)
}

// GetLogsByExecutorLatest returns the latest logs for an executor (most recent count logs)
func (client *ColoniesClient) GetLogsByExecutorLatest(colonyName, executorName string, count int, prvKey string) ([]*core.Log, error) {
	msg := rpc.CreateGetLogsMsg(colonyName, "", count, 0)
	msg.ExecutorName = executorName
	msg.Latest = true
	jsonString, err := msg.ToJSON()
	if err != nil {
		return []*core.Log{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLogsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return []*core.Log{}, err
	}

	return core.ConvertJSONToLogArray(respBodyString)
}

func (client *ColoniesClient) ChangeExecutorID(colonyName, executorID string, prvKey string) error {
	msg := rpc.CreateChangeExecutorIDMsg(colonyName, executorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.ChangeExecutorIDPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}