package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
)

type ColoniesClient struct {
	restyClient   *resty.Client
	host          string
	port          int
	insecure      bool
	skipTLSVerify bool
}

func CreateColoniesClient(host string, port int, insecure bool, skipTLSVerify bool) *ColoniesClient {
	client := &ColoniesClient{}
	client.restyClient = resty.New()

	client.host = host
	client.port = port
	client.insecure = insecure
	client.skipTLSVerify = skipTLSVerify

	if skipTLSVerify {
		client.restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client
}

func (client *ColoniesClient) SendRawMessage(jsonString string, insecure bool) (string, error) {
	protocol := "https"
	if client.insecure {
		protocol = "http"
	}
	resp, err := client.restyClient.R().
		SetBody(jsonString).
		Post(protocol + "://" + client.host + ":" + strconv.Itoa(client.port) + "/api")
	if err != nil {
		return "", err
	}

	return string(resp.Body()), nil
}

func (client *ColoniesClient) sendMessage(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
	var rpcMsg *rpc.RPCMsg
	var err error
	if insecure {
		rpcMsg, err = rpc.CreateInsecureRPCMsg(method, jsonString)
		if err != nil {
			return "", err
		}
	} else {
		rpcMsg, err = rpc.CreateRPCMsg(method, jsonString, prvKey)
		if err != nil {
			return "", err
		}
	}
	jsonString, err = rpcMsg.ToJSON()
	if err != nil {
		return "", err
	}

	protocol := "https"
	if client.insecure {
		protocol = "http"
	}
	resp, err := client.restyClient.R().
		SetContext(ctx).
		SetBody(jsonString).
		Post(protocol + "://" + client.host + ":" + strconv.Itoa(client.port) + "/api")
	if err != nil {
		return "", err
	}

	respBodyString := string(resp.Body())

	rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(respBodyString)
	if err != nil {
		return "", errors.New("Expected a valid Colonies RPC message, but got this: " + respBodyString)
	}

	if rpcReplyMsg.Error {
		failure, err := core.ConvertJSONToFailure(rpcReplyMsg.DecodePayload())
		if err != nil {
			return "", err
		}

		return "", &core.ColoniesError{Status: failure.Status, Message: failure.Message}
	}

	return rpcReplyMsg.DecodePayload(), nil
}

func (client *ColoniesClient) establishWebSocketConn(jsonString string) (*websocket.Conn, error) {
	dialer := *websocket.DefaultDialer
	var u url.URL

	if client.insecure {
		u = url.URL{Scheme: "ws", Host: client.host + ":" + strconv.Itoa(client.port), Path: "/pubsub"}
	} else {
		u = url.URL{Scheme: "wss", Host: client.host + ":" + strconv.Itoa(client.port), Path: "/pubsub"}
		if client.skipTLSVerify {
			dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}

	wsConn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	err = wsConn.WriteMessage(websocket.TextMessage, []byte(jsonString))
	if err != nil {
		return nil, err
	}

	return wsConn, nil
}

func (client *ColoniesClient) SubscribeProcesses(executorType string, state int, timeout int, prvKey string) (*ProcessSubscription, error) {
	msg := rpc.CreateSubscribeProcessesMsg(executorType, state, timeout)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	rpcMsg, err := rpc.CreateRPCMsg(rpc.SubscribeProcessesPayloadType, jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	jsonString, err = rpcMsg.ToJSON()
	if err != nil {
		return nil, err
	}

	wsConn, err := client.establishWebSocketConn(jsonString)
	if err != nil {
		return nil, err
	}

	subscription := createProcessSubscription(wsConn)
	go func(subscription *ProcessSubscription) {
		for {
			_, jsonBytes, err := subscription.wsConn.ReadMessage()
			if err != nil {
				subscription.ErrChan <- err
				continue
			}

			rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(string(jsonBytes))
			if err != nil {
				subscription.ErrChan <- err
				continue
			}

			if rpcReplyMsg.Error {
				failureMsg, err := core.ConvertJSONToFailure(rpcReplyMsg.DecodePayload())
				if err != nil {
					subscription.ErrChan <- err
				}
				subscription.ErrChan <- errors.New(failureMsg.Message)
			}

			process, err := core.ConvertJSONToProcess(rpcReplyMsg.DecodePayload())
			if err != nil {
				subscription.ErrChan <- err
				continue
			}

			subscription.ProcessChan <- process
		}
	}(subscription)

	return subscription, nil
}

func (client *ColoniesClient) SubscribeProcess(processID string, executorType string, state int, timeout int, prvKey string) (*ProcessSubscription, error) {
	msg := rpc.CreateSubscribeProcessMsg(processID, executorType, state, timeout)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	rpcMsg, err := rpc.CreateRPCMsg(rpc.SubscribeProcessPayloadType, jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	jsonString, err = rpcMsg.ToJSON()
	if err != nil {
		return nil, err
	}

	wsConn, err := client.establishWebSocketConn(jsonString)
	if err != nil {
		return nil, err
	}

	subscription := createProcessSubscription(wsConn)
	go func(subscription *ProcessSubscription) {
		for {
			_, jsonBytes, err := subscription.wsConn.ReadMessage()
			if err != nil {
				subscription.ErrChan <- err
				continue
			}
			rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(string(jsonBytes))
			if err != nil {
				subscription.ErrChan <- err
				continue
			}

			if rpcReplyMsg.Error {
				failureMsg, err := core.ConvertJSONToFailure(rpcReplyMsg.DecodePayload())
				if err != nil {
					subscription.ErrChan <- err
				}
				subscription.ErrChan <- errors.New(failureMsg.Message)
			}

			process, err := core.ConvertJSONToProcess(rpcReplyMsg.DecodePayload())
			if err != nil {
				subscription.ErrChan <- err
				continue
			}
			subscription.ProcessChan <- process
		}
	}(subscription)

	return subscription, nil
}

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

func (client *ColoniesClient) Assign(colonyName string, timeout int, prvKey string) (*core.Process, error) {
	msg := rpc.CreateAssignProcessMsg(colonyName)
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

func (client *ColoniesClient) AssignWithContext(colonyName string, timeout int, ctx context.Context, prvKey string) (*core.Process, error) {
	msg := rpc.CreateAssignProcessMsg(colonyName)
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

func (client *ColoniesClient) getProcesses(state int, colonyName string, executorType string, count int, prvKey string) ([]*core.Process, error) {
	msg := rpc.CreateGetProcessesMsg(colonyName, count, state, executorType)
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

func (client *ColoniesClient) getProcessesWithExecutorType(state int, colonyName string, count int, executorType string, prvKey string) ([]*core.Process, error) {
	msg := rpc.CreateGetProcessesMsg(colonyName, count, state, "")
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

func (client *ColoniesClient) GetWaitingProcesses(colonyName string, executorType string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.WAITING, colonyName, executorType, count, prvKey)
}

func (client *ColoniesClient) GetRunningProcesses(colonyName string, executorType string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.RUNNING, colonyName, executorType, count, prvKey)
}

func (client *ColoniesClient) GetSuccessfulProcesses(colonyName string, executorType string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.SUCCESS, colonyName, executorType, count, prvKey)
}

func (client *ColoniesClient) GetFailedProcesses(colonyName string, executorType string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.FAILED, colonyName, executorType, count, prvKey)
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

func (client *ColoniesClient) Statistics(prvKey string) (*core.Statistics, error) {
	msg := rpc.CreateGetStatisticsMsg()
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetStatisiticsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToStatistics(respBodyString)
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

func (client *ColoniesClient) AddAttribute(attribute core.Attribute, prvKey string) (core.Attribute, error) {
	msg := rpc.CreateAddAttributeMsg(attribute)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return core.Attribute{}, err
	}

	respBodyString, err := client.sendMessage(rpc.AddAttributePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return core.Attribute{}, err
	}

	return core.ConvertJSONToAttribute(respBodyString)
}

func (client *ColoniesClient) GetAttribute(attributeID string, prvKey string) (core.Attribute, error) {
	msg := rpc.CreateGetAttributeMsg(attributeID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return core.Attribute{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetAttributePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return core.Attribute{}, err
	}

	return core.ConvertJSONToAttribute(respBodyString)
}

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

func (client *ColoniesClient) AddGenerator(generator *core.Generator, prvKey string) (*core.Generator, error) {
	msg := rpc.CreateAddGeneratorMsg(generator)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddGeneratorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToGenerator(respBodyString)
}

func (client *ColoniesClient) GetGenerator(generatorID string, prvKey string) (*core.Generator, error) {
	msg := rpc.CreateGetGeneratorMsg(generatorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetGeneratorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToGenerator(respBodyString)
}

func (client *ColoniesClient) ResolveGenerator(colonyName string, generatorName string, prvKey string) (*core.Generator, error) {
	msg := rpc.CreateResolveGeneratorMsg(colonyName, generatorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.ResolveGeneratorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToGenerator(respBodyString)
}

func (client *ColoniesClient) GetGenerators(colonyName string, count int, prvKey string) ([]*core.Generator, error) {
	msg := rpc.CreateGetGeneratorsMsg(colonyName, count)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetGeneratorsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToGeneratorArray(respBodyString)
}

func (client *ColoniesClient) PackGenerator(generatorID string, arg string, prvKey string) error {
	msg := rpc.CreatePackGeneratorMsg(generatorID, arg)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.PackGeneratorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RemoveGenerator(generatorID string, prvKey string) error {
	msg := rpc.CreateRemoveGeneratorMsg(generatorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveGeneratorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) AddCron(cron *core.Cron, prvKey string) (*core.Cron, error) {
	msg := rpc.CreateAddCronMsg(cron)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddCronPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToCron(respBodyString)
}

func (client *ColoniesClient) GetCron(cronID string, prvKey string) (*core.Cron, error) {
	msg := rpc.CreateGetCronMsg(cronID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetCronPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToCron(respBodyString)
}

func (client *ColoniesClient) GetCrons(colonyName string, count int, prvKey string) ([]*core.Cron, error) {
	msg := rpc.CreateGetCronsMsg(colonyName, count)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetCronsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToCronArray(respBodyString)
}

func (client *ColoniesClient) RunCron(cronID string, prvKey string) (*core.Cron, error) {
	msg := rpc.CreateRunCronMsg(cronID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.RunCronPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToCron(respBodyString)
}

func (client *ColoniesClient) RemoveCron(cronID string, prvKey string) error {
	msg := rpc.CreateRemoveCronMsg(cronID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveCronPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) AddFunction(function *core.Function, prvKey string) (*core.Function, error) {
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddFunctionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFunction(respBodyString)
}

func (client *ColoniesClient) GetFunctionsByExecutorName(colonyName string, executorName string, prvKey string) ([]*core.Function, error) {
	msg := rpc.CreateGetFunctionsMsg(colonyName, executorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFunctionsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFunctionArray(respBodyString)
}

func (client *ColoniesClient) GetFunctionsByColonyName(colonyName string, prvKey string) ([]*core.Function, error) {
	msg := rpc.CreateGetFunctionsByColonyNameMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFunctionsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFunctionArray(respBodyString)
}

func (client *ColoniesClient) RemoveFunction(functionID string, prvKey string) error {
	msg := rpc.CreateRemoveFunctionMsg(functionID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveFunctionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) Version() (string, string, error) {
	msg := rpc.CreateVersionMsg("", "")
	jsonString, err := msg.ToJSON()
	if err != nil {
		return "", "", err
	}

	respBodyString, err := client.sendMessage(rpc.VersionPayloadType, jsonString, "", true, context.TODO())
	if err != nil {
		return "", "", err
	}

	version, err := rpc.CreateVersionMsgFromJSON(respBodyString)
	if err != nil {
		return "", "", err
	}

	return version.BuildVersion, version.BuildTime, nil
}

func (client *ColoniesClient) CheckHealth() error {
	protocol := "https"
	if client.insecure {
		protocol = "http"
	}
	_, err := client.restyClient.R().
		Get(protocol + "://" + client.host + ":" + strconv.Itoa(client.port) + "/health")

	return err
}

func (client *ColoniesClient) GetClusterInfo(prvKey string) (*cluster.Config, error) {
	msg := rpc.CreateGetClusterMsg()
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetClusterPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return cluster.ConvertJSONToConfig(respBodyString)
}

func (client *ColoniesClient) ResetDatabase(prvKey string) error {
	msg := rpc.CreateResetDatabaseMsg()
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.ResetDatabasePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) AddLog(processID string, logmsg string, prvKey string) error {
	msg := rpc.CreateAddLogMsg(processID, logmsg)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.AddLogPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) GetLogsByProcessID(processID string, count int, prvKey string) ([]core.Log, error) {
	msg := rpc.CreateGetLogsMsg(processID, count, 0)
	msg.ExecutorID = ""
	jsonString, err := msg.ToJSON()
	if err != nil {
		return []core.Log{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLogsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return []core.Log{}, err
	}

	return core.ConvertJSONToLogArray(respBodyString)
}

func (client *ColoniesClient) GetLogsByProcessIDSince(processID string, count int, since int64, prvKey string) ([]core.Log, error) {
	msg := rpc.CreateGetLogsMsg(processID, count, since)
	msg.ExecutorID = ""
	jsonString, err := msg.ToJSON()
	if err != nil {
		return []core.Log{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLogsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return []core.Log{}, err
	}

	return core.ConvertJSONToLogArray(respBodyString)
}

func (client *ColoniesClient) GetLogsByExecutorID(executorID string, count int, prvKey string) ([]core.Log, error) {
	msg := rpc.CreateGetLogsMsg("", count, 0)
	msg.ExecutorID = executorID
	jsonString, err := msg.ToJSON()
	if err != nil {
		return []core.Log{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLogsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return []core.Log{}, err
	}

	return core.ConvertJSONToLogArray(respBodyString)
}

func (client *ColoniesClient) GetLogsByExecutorIDSince(executorID string, count int, since int64, prvKey string) ([]core.Log, error) {
	msg := rpc.CreateGetLogsMsg("", count, since)
	msg.ExecutorID = executorID
	jsonString, err := msg.ToJSON()
	if err != nil {
		return []core.Log{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLogsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return []core.Log{}, err
	}

	return core.ConvertJSONToLogArray(respBodyString)
}

func (client *ColoniesClient) AddFile(file *core.File, prvKey string) (*core.File, error) {
	msg := rpc.CreateAddFileMsg(file)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFile(respBodyString)
}

func (client *ColoniesClient) GetFileByID(colonyName string, fileID string, prvKey string) ([]*core.File, error) {
	msg := rpc.CreateGetFileMsg(colonyName, fileID, "", "", false)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFileArray(respBodyString)
}

func (client *ColoniesClient) GetLatestFileByName(colonyName string, label string, name string, prvKey string) ([]*core.File, error) {
	msg := rpc.CreateGetFileMsg(colonyName, "", label, name, true)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFileArray(respBodyString)
}

func (client *ColoniesClient) GetFileByName(colonyName string, label string, name string, prvKey string) ([]*core.File, error) {
	msg := rpc.CreateGetFileMsg(colonyName, "", label, name, false)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFileArray(respBodyString)
}

func (client *ColoniesClient) GetFilenames(colonyName string, label string, prvKey string) ([]string, error) {
	msg := rpc.CreateGetFilesMsg(colonyName, label)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFilesPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	var filenames []string
	err = json.Unmarshal([]byte(respBodyString), &filenames)
	return filenames, err
}

func (client *ColoniesClient) GetFileLabels(colonyName string, prvKey string) ([]*core.Label, error) {
	msg := rpc.CreateGetAllFileLabelsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFileLabelsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	labels, err := core.ConvertJSONToLabelArray(respBodyString)
	if err != nil {
		return nil, err
	}

	return labels, err
}

func (client *ColoniesClient) GetFileLabelsByName(colonyName string, name string, prvKey string) ([]*core.Label, error) {
	msg := rpc.CreateGetFileLabelsMsg(colonyName, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFileLabelsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	labels, err := core.ConvertJSONToLabelArray(respBodyString)
	if err != nil {
		return nil, err
	}

	return labels, err
}

func (client *ColoniesClient) RemoveFileByID(colonyName string, fileID string, prvKey string) error {
	msg := rpc.CreateRemoveFileMsg(colonyName, fileID, "", "")
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RemoveFileByName(colonyName string, label string, name string, prvKey string) error {
	msg := rpc.CreateRemoveFileMsg(colonyName, "", label, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) CreateSnapshot(colonyName string, label string, name string, prvKey string) (*core.Snapshot, error) {
	msg := rpc.CreateCreateSnapshotMsg(colonyName, label, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.CreateSnapshotPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	snapshot, err := core.ConvertJSONToSnapshot(respBodyString)
	if err != nil {
		return nil, err
	}

	return snapshot, err
}

func (client *ColoniesClient) GetSnapshotByID(colonyName string, snapshotID string, prvKey string) (*core.Snapshot, error) {
	msg := rpc.CreateGetSnapshotMsg(colonyName, snapshotID, "")
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetSnapshotPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	snapshot, err := core.ConvertJSONToSnapshot(respBodyString)
	if err != nil {
		return nil, err
	}

	return snapshot, err
}

func (client *ColoniesClient) GetSnapshotByName(colonyName string, name string, prvKey string) (*core.Snapshot, error) {
	msg := rpc.CreateGetSnapshotMsg(colonyName, "", name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetSnapshotPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	snapshot, err := core.ConvertJSONToSnapshot(respBodyString)
	if err != nil {
		return nil, err
	}

	return snapshot, err
}

func (client *ColoniesClient) GetSnapshotsByColonyName(colonyName string, prvKey string) ([]*core.Snapshot, error) {
	msg := rpc.CreateGetSnapshotsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetSnapshotsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	snapshots, err := core.ConvertJSONToSnapshotsArray(respBodyString)
	if err != nil {
		return nil, err
	}

	return snapshots, err
}

func (client *ColoniesClient) RemoveSnapshotByID(colonyName string, snapshotID string, prvKey string) error {
	msg := rpc.CreateRemoveSnapshotMsg(colonyName, snapshotID, "")
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveSnapshotPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return err
}

func (client *ColoniesClient) RemoveSnapshotByName(colonyName string, name string, prvKey string) error {
	msg := rpc.CreateRemoveSnapshotMsg(colonyName, "", name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveSnapshotPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return err
}
