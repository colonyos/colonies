package client

import (
	"colonies/pkg/core"
	"colonies/pkg/rpc"
	"colonies/pkg/security/crypto"
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
)

type ColoniesClient struct {
	restyClient *resty.Client
	host        string
	port        int
}

func CreateColoniesClient(host string, port int, insecure bool) *ColoniesClient {
	client := &ColoniesClient{}
	client.restyClient = resty.New()

	client.host = host
	client.port = port

	if insecure {
		client.restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client
}

func (client *ColoniesClient) checkStatusCode(statusCode int, jsonString string) error {
	if statusCode != 200 {
		failure, err := core.ConvertJSONToFailure(jsonString)
		if err != nil {
			return err
		}

		return errors.New(failure.Message)
	}

	return nil
}

func (client *ColoniesClient) sendMessage(jsonString string, prvKey string) (string, error) {
	signature, err := crypto.CreateCrypto().GenerateSignature(jsonString, prvKey)
	if err != nil {
		return "", err
	}

	resp, err := client.restyClient.R().
		SetHeader("Signature", signature).
		SetBody(jsonString).
		Post("https://" + client.host + ":" + strconv.Itoa(client.port) + "/api")

	if err != nil {
		return "", err
	}

	respBodyString := string(resp.Body())
	err = client.checkStatusCode(resp.StatusCode(), respBodyString)
	if err != nil {
		return "", err
	}

	return respBodyString, nil
}

func (client *ColoniesClient) SubscribeProcesses(runtimeType string,
	state int,
	timeout int,
	prvKey string) (chan *core.Process, error) {
	u := url.URL{Scheme: "wss", Host: client.host + ":" + strconv.Itoa(client.port), Path: "/pubsub"}

	processChan := make(chan *core.Process)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return processChan, err
	}

	msg := rpc.CreateSubscribeProcessesMsg(runtimeType, state, timeout)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return processChan, err
	}

	signature, err := crypto.CreateCrypto().GenerateSignature(jsonString, prvKey)
	if err != nil {
		return processChan, err
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte(signature+"@"+jsonString))
	if err != nil {
		return processChan, err
	}

	go func(conn *websocket.Conn) {
		for {
			_, jsonBytes, err := conn.ReadMessage()
			if err != nil {
				// TODO
				fmt.Println(err)
				continue
			}
			process, err := core.ConvertJSONToProcess(string(jsonBytes))
			processChan <- process
			// TODO: defer conn.Close()
		}
	}(conn)

	return processChan, nil
}

func (client *ColoniesClient) AddColony(colony *core.Colony, prvKey string) (*core.Colony, error) {
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	addedColony, err := core.ConvertJSONToColony(respBodyString)
	if err != nil {
		return nil, err
	}

	return addedColony, nil
}

func (client *ColoniesClient) GetColonies(prvKey string) ([]*core.Colony, error) {
	msg := rpc.CreateGetColoniesMsg()
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
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

	respBodyString, err := client.sendMessage(jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToColony(respBodyString)
}

func (client *ColoniesClient) AddRuntime(runtime *core.Runtime, prvKey string) (*core.Runtime, error) {
	msg := rpc.CreateAddRuntimeMsg(runtime)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToRuntime(respBodyString)
}

func (client *ColoniesClient) GetRuntimes(colonyID string, prvKey string) ([]*core.Runtime, error) {
	msg := rpc.CreateGetRuntimesMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToRuntimeArray(respBodyString)
}

func (client *ColoniesClient) GetRuntime(runtimeID string, prvKey string) (*core.Runtime, error) {
	msg := rpc.CreateGetRuntimeMsg(runtimeID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToRuntime(respBodyString)
}

func (client *ColoniesClient) ApproveRuntime(runtimeID string, prvKey string) error {
	msg := rpc.CreateApproveRuntimeMsg(runtimeID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(jsonString, prvKey)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RejectRuntime(runtimeID string, prvKey string) error {
	msg := rpc.CreateRejectRuntimeMsg(runtimeID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(jsonString, prvKey)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) SubmitProcessSpec(processSpec *core.ProcessSpec, prvKey string) (*core.Process, error) {
	msg := rpc.CreateSubmitProcessSpecMsg(processSpec)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) AssignProcess(colonyID string, prvKey string) (*core.Process, error) {
	msg := rpc.CreateAssignProcessMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) getProcesses(state int, colonyID string, count int, prvKey string) ([]*core.Process, error) {
	msg := rpc.CreateGetProcessesMsg(colonyID, count, state)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
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

func (client *ColoniesClient) GetProcessByID(processID string, prvKey string) (*core.Process, error) {
	msg := rpc.CreateGetProcessMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func (client *ColoniesClient) MarkSuccessful(processID string, prvKey string) error {
	msg := rpc.CreateMarkSuccessfulMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(jsonString, prvKey)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) MarkFailed(processID string, prvKey string) error {
	msg := rpc.CreateMarkFailedMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(jsonString, prvKey)
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) AddAttribute(attribute *core.Attribute, prvKey string) (*core.Attribute, error) {
	msg := rpc.CreateAddAttributeMsg(attribute)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToAttribute(respBodyString)
}

func (client *ColoniesClient) GetAttribute(attributeID string, prvKey string) (*core.Attribute, error) {
	msg := rpc.CreateGetAttributeMsg(attributeID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToAttribute(respBodyString)
}
