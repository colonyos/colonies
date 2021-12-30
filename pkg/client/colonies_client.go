package client

import (
	"colonies/pkg/core"
	"colonies/pkg/rpc"
	"colonies/pkg/security"
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

func sendRPCMessageNoSignature(client *resty.Client, jsonString string, host string, port int) (string, error) {
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

func sendRPCMessage(client *resty.Client, jsonString string, prvKey string, host string, port int) (string, error) {
	signature, err := security.GenerateSignature(jsonString, prvKey)
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

// OK
func AddColony(colony *core.Colony, rootPassword string, host string, port int) (*core.Colony, error) {
	client := client()

	rpc := rpc.CreateAddColonyRPC(rootPassword, colony)
	jsonString, err := rpc.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendRPCMessageNoSignature(client, jsonString, host, port)
	if err != nil {
		return nil, err
	}

	addedColony, err := core.ConvertJSONToColony(respBodyString)
	if err != nil {
		return nil, err
	}

	return addedColony, nil
}

// OK
func GetColonies(rootPassword string, host string, port int) ([]*core.Colony, error) {
	client := client()

	rpc := rpc.CreateGetColoniesRPC(rootPassword)
	jsonString, err := rpc.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendRPCMessageNoSignature(client, jsonString, host, port)
	if err != nil {
		return nil, err
	}

	colonies, err := core.ConvertJSONToColonyArray(respBodyString)
	if err != nil {
		return colonies, err
	}

	return colonies, nil
}

// OK
func GetColonyByID(colonyID string, prvKey string, host string, port int) (*core.Colony, error) {
	client := client()

	rpc := rpc.CreateGetColonyRPC(colonyID)
	jsonString, err := rpc.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendRPCMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	colony, err := core.ConvertJSONToColony(respBodyString)
	if err != nil {
		return nil, err
	}

	return colony, nil
}

// OK
func AddRuntime(runtime *core.Runtime, prvKey string, host string, port int) (*core.Runtime, error) {
	client := client()

	rpc := rpc.CreateAddRuntimeRPC(runtime)
	jsonString, err := rpc.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendRPCMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	addedRuntime, err := core.ConvertJSONToRuntime(respBodyString)
	if err != nil {
		return nil, err
	}

	return addedRuntime, nil
}

// OK
func GetRuntimes(colonyID string, prvKey string, host string, port int) ([]*core.Runtime, error) {
	client := client()

	rpc := rpc.CreateGetRuntimesRPC(colonyID)
	jsonString, err := rpc.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendRPCMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	runtimes, err := core.ConvertJSONToRuntimeArray(respBodyString)
	if err != nil {
		return nil, err
	}

	return runtimes, nil
}

// Ok
func GetRuntime(runtimeID string, prvKey string, host string, port int) (*core.Runtime, error) {
	client := client()

	rpc := rpc.CreateGetRuntimeRPC(runtimeID)
	jsonString, err := rpc.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendRPCMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	runtime, err := core.ConvertJSONToRuntime(respBodyString)
	if err != nil {
		return nil, err
	}

	return runtime, nil
}

// Ok
func ApproveRuntime(runtimeID string, prvKey string, host string, port int) error {
	client := client()

	rpc := rpc.CreateApproveRuntimeRPC(runtimeID)
	jsonString, err := rpc.ToJSON()
	if err != nil {
		return err
	}

	_, err = sendRPCMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return err
	}

	return nil
}

// Ok
func RejectRuntime(runtimeID string, prvKey string, host string, port int) error {
	client := client()

	rpc := rpc.CreateRejectRuntimeRPC(runtimeID)
	jsonString, err := rpc.ToJSON()
	if err != nil {
		return err
	}

	_, err = sendRPCMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return err
	}

	return nil
}

// OK
func SubmitProcessSpec(processSpec *core.ProcessSpec, prvKey string, host string, port int) (*core.Process, error) {
	client := client()

	rpc := rpc.CreateSubmitProcessSpecRPC(processSpec)
	jsonString, err := rpc.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := sendRPCMessage(client, jsonString, prvKey, host, port)
	if err != nil {
		return nil, err
	}

	addedProcess, err := core.ConvertJSONToProcess(respBodyString)
	if err != nil {
		return nil, err
	}

	return addedProcess, nil
}

func AddAttribute(attribute *core.Attribute, colonyID string, prvKey string) (*core.Attribute, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	jsonString, err := attribute.ToJSON()
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetBody(jsonString).
		Post("https://localhost:8080/colonies/" + colonyID + "/processes/" + attribute.TargetID + "/attributes")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return nil, err
	}

	addedAttribute, err := core.ConvertJSONToAttribute(unquotedResp)
	if err != nil {
		return nil, err
	}

	return addedAttribute, nil
}

func GetAttribute(attributeID string, processID string, colonyID string, prvKey string) (*core.Attribute, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Get("https://localhost:8080/colonies/" + colonyID + "/processes/" + processID + "/attributes/" + attributeID)

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return nil, err
	}

	attribute, err := core.ConvertJSONToAttribute(unquotedResp)
	if err != nil {
		return nil, err
	}

	return attribute, nil
}

func GetProcessByID(processID string, colonyID string, prvKey string) (*core.Process, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Get("https://localhost:8080/colonies/" + colonyID + "/processes/" + processID)

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return nil, err
	}

	process, err := core.ConvertJSONToProcess(unquotedResp)
	if err != nil {
		return nil, err
	}

	return process, nil
}

func GetWaitingProcesses(runtimeID string, colonyID string, count int, prvKey string) ([]*core.Process, error) {
	var processes []*core.Process
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return processes, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetHeader("RuntimeId", runtimeID).
		SetHeader("Count", strconv.Itoa(count)).
		SetHeader("State", strconv.Itoa(core.WAITING)).
		Get("https://localhost:8080/colonies/" + colonyID + "/processes")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return processes, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return processes, err
	}

	processes, err = core.ConvertJSONToProcessArray(unquotedResp)

	return processes, nil
}

func GetRunningProcesses(colonyID string, count int, prvKey string) ([]*core.Process, error) { // TODO: unittest
	var processes []*core.Process
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return processes, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetHeader("Count", strconv.Itoa(count)).
		SetHeader("State", strconv.Itoa(core.RUNNING)).
		Get("https://localhost:8080/colonies/" + colonyID + "/processes")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return processes, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return processes, err
	}

	processes, err = core.ConvertJSONToProcessArray(unquotedResp)

	return processes, nil
}

func GetSuccessfulProcesses(colonyID string, count int, prvKey string) ([]*core.Process, error) { // TODO: unittest
	var processes []*core.Process
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return processes, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetHeader("Count", strconv.Itoa(count)).
		SetHeader("State", strconv.Itoa(core.SUCCESS)).
		Get("https://localhost:8080/colonies/" + colonyID + "/processes")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return processes, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return processes, err
	}

	processes, err = core.ConvertJSONToProcessArray(unquotedResp)

	return processes, nil
}

func GetFailedProcesses(colonyID string, count int, prvKey string) ([]*core.Process, error) { // TODO: unittest
	var processes []*core.Process
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return processes, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetHeader("Count", strconv.Itoa(count)).
		SetHeader("State", strconv.Itoa(core.FAILED)).
		Get("https://localhost:8080/colonies/" + colonyID + "/processes")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return processes, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return processes, err
	}

	processes, err = core.ConvertJSONToProcessArray(unquotedResp)

	return processes, nil
}

func AssignProcess(runtimeID string, colonyID string, prvKey string) (*core.Process, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetHeader("RuntimeId", runtimeID).
		Get("https://localhost:8080/colonies/" + colonyID + "/processes/assign")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return nil, err
	}

	process, err := core.ConvertJSONToProcess(unquotedResp)
	if err != nil {
		return nil, err
	}

	return process, nil
}

func MarkSuccessful(process *core.Process, prvKey string) error {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Put("https://localhost:8080/colonies/" + process.ProcessSpec.Conditions.ColonyID + "/processes/" + process.ID + "/finish")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return err
	}

	return nil
}

func MarkFailed(process *core.Process, prvKey string) error {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Put("https://localhost:8080/colonies/" + process.ProcessSpec.Conditions.ColonyID + "/processes/" + process.ID + "/failed")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return err
	}

	return nil
}
