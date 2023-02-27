package client

import (
	"crypto/tls"
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

func (client *ColoniesClient) sendMessage(method string, jsonString string, prvKey string, insecure bool) (string, error) {
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

		return "", errors.New(failure.Message)
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

func (client *ColoniesClient) AddColony(colony *core.Colony, prvKey string) (*core.Colony, error) {
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddColonyPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	addedColony, err := core.ConvertJSONToColony(respBodyString)
	if err != nil {
		return nil, err
	}

	return addedColony, nil
}

func (client *ColoniesClient) DeleteColony(colonyID string, prvKey string) error {
	msg := rpc.CreateDeleteColonyMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.DeleteColonyPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.GetColoniesPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToColonyArray(respBodyString)
}

func (client *ColoniesClient) GetColonyByID(colonyID string, prvKey string) (*core.Colony, error) {
	msg := rpc.CreateGetColonyMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetColonyPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.AddExecutorPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToExecutor(respBodyString)
}

func (client *ColoniesClient) GetExecutors(colonyID string, prvKey string) ([]*core.Executor, error) {
	msg := rpc.CreateGetExecutorsMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetExecutorsPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToExecutorArray(respBodyString)
}

func (client *ColoniesClient) GetExecutor(executorID string, prvKey string) (*core.Executor, error) {
	msg := rpc.CreateGetExecutorMsg(executorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetExecutorPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToExecutor(respBodyString)
}

func (client *ColoniesClient) ApproveExecutor(executorID string, prvKey string) error {
	msg := rpc.CreateApproveExecutorMsg(executorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.ApproveExecutorPayloadType, jsonString, prvKey, false)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RejectExecutor(executorID string, prvKey string) error {
	msg := rpc.CreateRejectExecutorMsg(executorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RejectExecutorPayloadType, jsonString, prvKey, false)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) DeleteExecutor(executorID string, prvKey string) error {
	msg := rpc.CreateDeleteExecutorMsg(executorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.DeleteExecutorPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.SubmitFunctionSpecPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) Assign(colonyID string, timeout int, prvKey string) (*core.Process, error) {
	msg := rpc.CreateAssignProcessMsg(colonyID)
	msg.Latest = false
	msg.Timeout = timeout
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AssignProcessPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) AssignLatestProcess(colonyID string, timeout int, prvKey string) (*core.Process, error) {
	msg := rpc.CreateAssignProcessMsg(colonyID)
	msg.Latest = true
	msg.Timeout = timeout
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AssignProcessPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) GetProcessHistForColony(state int, colonyID string, seconds int, prvKey string) ([]*core.Process, error) {
	msg := rpc.CreateGetProcessHistMsg(colonyID, "", seconds, state)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetProcessHistPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessArray(respBodyString)
}

func (client *ColoniesClient) GetProcessHistForExecutor(state int, colonyID string, executorID string, seconds int, prvKey string) ([]*core.Process, error) {
	msg := rpc.CreateGetProcessHistMsg(colonyID, executorID, seconds, state)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetProcessHistPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessArray(respBodyString)
}

func (client *ColoniesClient) getProcesses(state int, colonyID string, count int, prvKey string) ([]*core.Process, error) {
	msg := rpc.CreateGetProcessesMsg(colonyID, count, state)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetProcessesPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessArray(respBodyString)
}

func (client *ColoniesClient) GetWaitingProcesses(colonyID string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.WAITING, colonyID, count, prvKey)
}

func (client *ColoniesClient) GetRunningProcesses(colonyID string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.RUNNING, colonyID, count, prvKey)
}

func (client *ColoniesClient) GetSuccessfulProcesses(colonyID string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.SUCCESS, colonyID, count, prvKey)
}

func (client *ColoniesClient) GetFailedProcesses(colonyID string, count int, prvKey string) ([]*core.Process, error) {
	return client.getProcesses(core.FAILED, colonyID, count, prvKey)
}

func (client *ColoniesClient) ColonyStatistics(colonyID string, prvKey string) (*core.Statistics, error) {
	msg := rpc.CreateGetColonyStatisticsMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetColonyStatisticsPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.GetStatisiticsPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.GetProcessPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) DeleteProcess(processID string, prvKey string) error {
	msg := rpc.CreateDeleteProcessMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.DeleteProcessPayloadType, jsonString, prvKey, false)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) DeleteAllProcesses(colonyID string, prvKey string) error {
	msg := rpc.CreateDeleteAllProcessesMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.DeleteAllProcessesPayloadType, jsonString, prvKey, false)
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

	_, err = client.sendMessage(rpc.CloseSuccessfulPayloadType, jsonString, prvKey, false)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) CloseWithOutput(processID string, output []string, prvKey string) error {
	msg := rpc.CreateCloseSuccessfulMsg(processID)
	msg.Output = output
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.CloseSuccessfulPayloadType, jsonString, prvKey, false)
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

	_, err = client.sendMessage(rpc.CloseFailedPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.AddAttributePayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.GetAttributePayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.SubmitWorkflowSpecPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessGraph(respBodyString)
}

func (client *ColoniesClient) AddChild(processGraphID string, processID string, funcSpec *core.FunctionSpec, prvKey string) (*core.Process, error) {
	msg := rpc.CreateAddChildMsg(processGraphID, processID, funcSpec)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddChildPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.GetProcessGraphPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessGraph(respBodyString)
}

func (client *ColoniesClient) getProcessGraphs(state int, colonyID string, count int, prvKey string) ([]*core.ProcessGraph, error) {
	msg := rpc.CreateGetProcessGraphsMsg(colonyID, count, state)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetProcessGraphsPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessGraphArray(respBodyString)
}

func (client *ColoniesClient) GetWaitingProcessGraphs(colonyID string, count int, prvKey string) ([]*core.ProcessGraph, error) {
	return client.getProcessGraphs(core.WAITING, colonyID, count, prvKey)
}

func (client *ColoniesClient) GetRunningProcessGraphs(colonyID string, count int, prvKey string) ([]*core.ProcessGraph, error) {
	return client.getProcessGraphs(core.RUNNING, colonyID, count, prvKey)
}

func (client *ColoniesClient) GetSuccessfulProcessGraphs(colonyID string, count int, prvKey string) ([]*core.ProcessGraph, error) {
	return client.getProcessGraphs(core.SUCCESS, colonyID, count, prvKey)
}

func (client *ColoniesClient) GetFailedProcessGraphs(colonyID string, count int, prvKey string) ([]*core.ProcessGraph, error) {
	return client.getProcessGraphs(core.FAILED, colonyID, count, prvKey)
}

func (client *ColoniesClient) DeleteProcessGraph(processGraphID string, prvKey string) error {
	msg := rpc.CreateDeleteProcessGraphMsg(processGraphID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.DeleteProcessGraphPayloadType, jsonString, prvKey, false)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) DeleteAllProcessGraphs(colonyID string, prvKey string) error {
	msg := rpc.CreateDeleteAllProcessGraphsMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.DeleteAllProcessGraphsPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.AddGeneratorPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.GetGeneratorPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToGenerator(respBodyString)
}

func (client *ColoniesClient) ResolveGenerator(generatorName string, prvKey string) (*core.Generator, error) {
	msg := rpc.CreateResolveGeneratorMsg(generatorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.ResolveGeneratorPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToGenerator(respBodyString)
}

func (client *ColoniesClient) GetGenerators(colonyID string, count int, prvKey string) ([]*core.Generator, error) {
	msg := rpc.CreateGetGeneratorsMsg(colonyID, count)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetGeneratorsPayloadType, jsonString, prvKey, false)
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

	_, err = client.sendMessage(rpc.PackGeneratorPayloadType, jsonString, prvKey, false)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) DeleteGenerator(generatorID string, prvKey string) error {
	msg := rpc.CreateDeleteGeneratorMsg(generatorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.DeleteGeneratorPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.AddCronPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.GetCronPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToCron(respBodyString)
}

func (client *ColoniesClient) GetCrons(colonyID string, count int, prvKey string) ([]*core.Cron, error) {
	msg := rpc.CreateGetCronsMsg(colonyID, count)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetCronsPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.RunCronPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToCron(respBodyString)
}

func (client *ColoniesClient) DeleteCron(cronID string, prvKey string) error {
	msg := rpc.CreateDeleteCronMsg(cronID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.DeleteCronPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.AddFunctionPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFunction(respBodyString)
}

func (client *ColoniesClient) GetFunctionsByExecutorID(executorID string, prvKey string) ([]*core.Function, error) {
	msg := rpc.CreateGetFunctionsByExecutorIDMsg(executorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFunctionsPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFunctionArray(respBodyString)
}

func (client *ColoniesClient) GetFunctionsByColonyID(colonyID string, prvKey string) ([]*core.Function, error) {
	msg := rpc.CreateGetFunctionsByColonyIDMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFunctionsPayloadType, jsonString, prvKey, false)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFunctionArray(respBodyString)
}

func (client *ColoniesClient) DeleteFunction(functionID string, prvKey string) error {
	msg := rpc.CreateDeleteFunctionMsg(functionID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.DeleteFunctionPayloadType, jsonString, prvKey, false)
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

	respBodyString, err := client.sendMessage(rpc.VersionPayloadType, jsonString, "", true)
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

	respBodyString, err := client.sendMessage(rpc.GetClusterPayloadType, jsonString, prvKey, false)
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

	_, err = client.sendMessage(rpc.ResetDatabasePayloadType, jsonString, prvKey, false)
	if err != nil {
		return err
	}

	return nil
}
