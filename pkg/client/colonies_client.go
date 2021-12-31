package client

import (
	"colonies/pkg/core"
	"colonies/pkg/rpc"
	"colonies/pkg/security/crypto"
	"crypto/tls"
	"errors"
	"strconv"

	"github.com/go-resty/resty/v2"
)

func client() *resty.Client {
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	return client
}

func checkStatusCode(statusCode int, jsonString string) error {
	if statusCode != 200 {
		failure, err := core.ConvertJSONToFailure(jsonString)
		if err != nil {
			return err
		}

		return errors.New(failure.Message)
	}

	return nil
}

func sendMessageNoSignature(client *resty.Client, jsonString string, host string, port int) (string, error) {
	resp, err := client.R().
		SetBody(jsonString).
		Post("https://" + host + ":" + strconv.Itoa(port) + "/endpoint")

	respBodyString := string(resp.Body())
	err = checkStatusCode(resp.StatusCode(), respBodyString)
	if err != nil {
		return "", err
	}

	return respBodyString, nil
}

func sendMessage(client *resty.Client, jsonString string, prvKey string, host string, port int) (string, error) {
	signature, err := crypto.CreateCrypto().GenerateSignature(jsonString, prvKey)
	resp, err := client.R().
		SetHeader("Signature", signature).
		SetBody(jsonString).
		Post("https://" + host + ":" + strconv.Itoa(port) + "/endpoint")

	respBodyString := string(resp.Body())
	err = checkStatusCode(resp.StatusCode(), respBodyString)
	if err != nil {
		return "", err
	}

	return respBodyString, nil
}

func AddColony(colony *core.Colony, rootPassword string, host string, port int) (*core.Colony, error) {
	client := client()

	msg := rpc.CreateAddColonyMsg(rootPassword, colony)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessageNoSignature(client, jsonString, host, port)
	if err != nil {
		return nil, err
	}

	addedColony, err := core.ConvertJSONToColony(respBodyString)
	if err != nil {
		return nil, err
	}

	return addedColony, nil
}

func GetColonies(rootPassword string, host string, port int) ([]*core.Colony, error) {
	client := client()

	msg := rpc.CreateGetColoniesMsg(rootPassword)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessageNoSignature(client, jsonString, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToColonyArray(respBodyString)
}

func GetColonyByID(colonyID string, prvKey string, host string, port int) (*core.Colony, error) {
	client := client()

	msg := rpc.CreateGetColonyMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToColony(respBodyString)
}

func AddRuntime(runtime *core.Runtime, prvKey string, host string, port int) (*core.Runtime, error) {
	client := client()

	msg := rpc.CreateAddRuntimeMsg(runtime)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToRuntime(respBodyString)
}

func GetRuntimes(colonyID string, prvKey string, host string, port int) ([]*core.Runtime, error) {
	client := client()

	msg := rpc.CreateGetRuntimesMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToRuntimeArray(respBodyString)
}

func GetRuntime(runtimeID string, prvKey string, host string, port int) (*core.Runtime, error) {
	client := client()

	msg := rpc.CreateGetRuntimeMsg(runtimeID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToRuntime(respBodyString)
}

func ApproveRuntime(runtimeID string, prvKey string, host string, port int) error {
	client := client()

	msg := rpc.CreateApproveRuntimeMsg(runtimeID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return err
	}

	return nil
}

func RejectRuntime(runtimeID string, prvKey string, host string, port int) error {
	client := client()

	msg := rpc.CreateRejectRuntimeMsg(runtimeID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return err
	}

	return nil
}

func SubmitProcessSpec(processSpec *core.ProcessSpec, prvKey string, host string, port int) (*core.Process, error) {
	client := client()

	msg := rpc.CreateSubmitProcessSpecMsg(processSpec)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func AssignProcess(colonyID string, prvKey string, host string, port int) (*core.Process, error) {
	client := client()

	msg := rpc.CreateAssignProcessMsg(colonyID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func getProcesses(state int, colonyID string, count int, prvKey string, host string, port int) ([]*core.Process, error) {
	client := client()

	msg := rpc.CreateGetProcessesMsg(colonyID, count, state)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcessArray(respBodyString)
}

func GetWaitingProcesses(colonyID string, count int, prvKey string, host string, port int) ([]*core.Process, error) {
	return getProcesses(core.WAITING, colonyID, count, prvKey, host, port)
}

func GetRunningProcesses(colonyID string, count int, prvKey string, host string, port int) ([]*core.Process, error) {
	return getProcesses(core.RUNNING, colonyID, count, prvKey, host, port)
}

func GetSuccessfulProcesses(colonyID string, count int, prvKey string, host string, port int) ([]*core.Process, error) {
	return getProcesses(core.SUCCESS, colonyID, count, prvKey, host, port)
}

func GetFailedProcesses(colonyID string, count int, prvKey string, host string, port int) ([]*core.Process, error) {
	return getProcesses(core.FAILED, colonyID, count, prvKey, host, port)
}

func GetProcessByID(processID string, prvKey string, host string, port int) (*core.Process, error) {
	client := client()

	msg := rpc.CreateGetProcessMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}

func MarkSuccessful(processID string, prvKey string, host string, port int) error {
	client := client()

	msg := rpc.CreateMarkSuccessfulMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return err
	}

	return nil
}

func MarkFailed(processID string, prvKey string, host string, port int) error {
	client := client()

	msg := rpc.CreateMarkFailedMsg(processID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return err
	}

	return nil
}

func AddAttribute(attribute *core.Attribute, prvKey string, host string, port int) (*core.Attribute, error) {
	client := client()

	msg := rpc.CreateAddAttributeMsg(attribute)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToAttribute(respBodyString)
}

func GetAttribute(attributeID string, prvKey string, host string, port int) (*core.Attribute, error) {
	client := client()

	msg := rpc.CreateGetAttributeMsg(attributeID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToAttribute(respBodyString)
}
