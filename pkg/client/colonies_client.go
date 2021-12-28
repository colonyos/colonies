package client

import (
	"colonies/pkg/core"
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

func AddColony(colony *core.Colony, rootPassword string, host string, port int) (*core.Colony, error) {
	client := client()

	colonyJSON, err := colony.ToJSON()
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("RootPassword", rootPassword).
		SetBody(colonyJSON).
		Post("https://" + host + ":" + strconv.Itoa(port) + "/colonies")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return nil, err
	}

	addedColony, err := core.ConvertJSONToColony(unquotedResp)
	if err != nil {
		return nil, err
	}

	return addedColony, nil
}

func GetColonies(rootPassword string, host string, port int) ([]*core.Colony, error) {
	client := client()

	var colonies []*core.Colony
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("RootPassword", rootPassword).
		Get("https://" + host + ":" + strconv.Itoa(port) + "/colonies")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return colonies, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return colonies, err
	}

	colonies, err = core.ConvertJSONToColonyArray(unquotedResp)
	if err != nil {
		return colonies, err
	}

	return colonies, nil
}

func GetColonyByID(colonyID string, prvKey string) (*core.Colony, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Get("https://localhost:8080/colonies/" + colonyID)

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return nil, err
	}

	colony, err := core.ConvertJSONToColony(unquotedResp)
	if err != nil {
		return nil, err
	}

	return colony, nil
}

func AddRuntime(runtime *core.Runtime, prvKey string, host string, port int) (*core.Runtime, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	runtimeJSON, err := runtime.ToJSON()
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetBody(runtimeJSON).
		Post("https://" + host + ":" + strconv.Itoa(port) + "/colonies/" + runtime.ColonyID + "/runtimes")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return nil, err
	}

	addedRuntime, err := core.ConvertJSONToRuntime(unquotedResp)
	if err != nil {
		return nil, err
	}

	return addedRuntime, nil
}

func ApproveRuntime(runtime *core.Runtime, prvKey string, host string, port int) error {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Put("https://" + host + ":" + strconv.Itoa(port) + "/colonies/" + runtime.ColonyID + "/runtimes/" + runtime.ID + "/approve")

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

func RejectRuntime(runtime *core.Runtime, prvKey string, host string, port int) error {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Put("https://" + host + ":" + strconv.Itoa(port) + "/colonies/" + runtime.ColonyID + "/runtimes/" + runtime.ID + "/reject")

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

func GetRuntimesByColonyID(colonyID string, prvKey string, host string, port int) ([]*core.Runtime, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Get("https://" + host + ":" + strconv.Itoa(port) + "/colonies/" + colonyID + "/runtimes")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return nil, err
	}

	runtimes, err := core.ConvertJSONToRuntimeArray(unquotedResp)
	if err != nil {
		return nil, err
	}

	return runtimes, nil
}

func GetRuntimeByID(runtimeID string, colonyID string, prvKey string, host string, port int) (*core.Runtime, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Get("https://" + host + ":" + strconv.Itoa(port) + "/colonies/" + colonyID + "/runtimes/" + runtimeID)

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return nil, err
	}

	runtime, err := core.ConvertJSONToRuntime(unquotedResp)
	if err != nil {
		return nil, err
	}

	return runtime, nil
}

func PublishProcessSpec(processSpec *core.ProcessSpec, prvKey string, host string, port int) (*core.Process, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	processSpecJSON, err := processSpec.ToJSON()
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetBody(processSpecJSON).
		Post("https://" + host + ":" + strconv.Itoa(port) + "/colonies/" + processSpec.Conditions.ColonyID + "/processes")

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	err = checkStatusCode(resp.StatusCode(), unquotedResp)
	if err != nil {
		return nil, err
	}

	addedProcess, err := core.ConvertJSONToProcess(unquotedResp)
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
