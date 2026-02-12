package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) Submit(funcSpec *core.FunctionSpec, prvKey string) (*core.Process, error) {
	msg := rpc.CreateSubmitFunctionSpecMsg(funcSpec)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.SubmitFunctionSpecPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) Assign(colonyName string, timeout int, availableCPU string, availableMem string, prvKey string) (*core.Process, error) {
	msg := rpc.CreateAssignProcessMsg(colonyName, availableCPU, availableMem)
	msg.Timeout = timeout
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AssignProcessPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) AssignWithContext(colonyName string,
	timeout int,
	ctx context.Context,
	availableCPU string,
	availableMem string,
	prvKey string) (*core.Process, error) {
	msg := rpc.CreateAssignProcessMsg(colonyName, availableCPU, availableMem)
	msg.Timeout = timeout
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AssignProcessPayloadType, jsonString, prvKey, false, ctx)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) GetProcessHistForColony(state int, colonyName string, seconds int, prvKey string) ([]*core.Process, error) {
	msg := rpc.CreateGetProcessHistMsg(colonyName, "", seconds, state)
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

func (client *ColoniesClient) getProcesses(state int, colonyName string, executorType string, label string, initiator string, count int, prvKey string) ([]*core.Process, error) {
	msg := rpc.CreateGetProcessesMsg(colonyName, count, state, executorType, label, initiator)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetProcessesPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessArray(respBodyString)
}

func (client *ColoniesClient) GetWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.WAITING, colonyName, executorType, label, initiator, count, prvKey)
}

func (client *ColoniesClient) GetRunningProcesses(colonyName string, executorType string, label string, initiator string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.RUNNING, colonyName, executorType, label, initiator, count, prvKey)
}

func (client *ColoniesClient) GetSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.SUCCESS, colonyName, executorType, label, initiator, count, prvKey)
}

func (client *ColoniesClient) GetFailedProcesses(colonyName string, executorType string, label string, initiator string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.FAILED, colonyName, executorType, label, initiator, count, prvKey)
}

func (client *ColoniesClient) GetCancelledProcesses(colonyName string, executorType string, label string, initiator string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.CANCELLED, colonyName, executorType, label, initiator, count, prvKey)
}

func (client *ColoniesClient) CancelProcess(processID string, prvKey string) error {
	msg := rpc.CreateCancelProcessMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.CancelProcessPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) GetProcess(processID string, prvKey string) (*core.Process, error) {
	msg := rpc.CreateGetProcessMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetProcessPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) RemoveProcess(processID string, prvKey string) error {
	msg := rpc.CreateRemoveProcessMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveProcessPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RemoveAllProcesses(colonyName string, prvKey string) error {
	msg := rpc.CreateRemoveAllProcessesMsg(colonyName)
	msg.State = core.NOTSET
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveAllProcessesPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RemoveAllProcessesWithState(colonyName string, state int, prvKey string) error {
	msg := rpc.CreateRemoveAllProcessesMsg(colonyName)
	msg.State = state
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveAllProcessesPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) Close(processID string, prvKey string) error {
	msg := rpc.CreateCloseSuccessfulMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.CloseSuccessfulPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) CloseWithContext(processID string, ctx context.Context, prvKey string) error {
	msg := rpc.CreateCloseSuccessfulMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.CloseSuccessfulPayloadType, jsonString, prvKey, false, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) CloseWithOutput(processID string, output []interface{}, prvKey string) error {
	msg := rpc.CreateCloseSuccessfulMsg(processID)
	msg.Output = output
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.CloseSuccessfulPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) Fail(processID string, errs []string, prvKey string) error {
	msg := rpc.CreateCloseFailedMsg(processID, errs)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.CloseFailedPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) SetOutput(processID string, output []interface{}, prvKey string) error {
	msg := rpc.CreateSetOutputMsg(processID, output)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.SetOutputPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

