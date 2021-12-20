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

		return errors.New(failure.Message())
	}

	return nil
}

func AddColony(colony *core.Colony, rootPassword string) error {
	client := client()

	colonyJSON, err := colony.ToJSON()
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("RootPassword", rootPassword).
		SetBody(colonyJSON).
		Post("https://localhost:8080/colonies")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return err
	}

	return nil
}

func GetColonies(rootPassword string) ([]*core.Colony, error) {
	client := client()

	var colonies []*core.Colony
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("RootPassword", rootPassword).
		Get("https://localhost:8080/colonies")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return colonies, err
	}

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
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

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return nil, err
	}

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	colony, err := core.CreateColonyFromJSON(unquotedResp)
	if err != nil {
		return nil, err
	}

	return colony, nil
}

func AddComputer(computer *core.Computer, prvKey string) error {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return err
	}

	computerJSON, err := computer.ToJSON()
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetBody(computerJSON).
		Post("https://localhost:8080/colonies/" + computer.ColonyID() + "/computers")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return err
	}

	return nil
}

func ApproveComputer(computer *core.Computer, prvKey string) error {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Put("https://localhost:8080/colonies/" + computer.ColonyID() + "/computers/" + computer.ID() + "/approve")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return err
	}

	return nil
}

func RejectComputer(computer *core.Computer, prvKey string) error {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Put("https://localhost:8080/colonies/" + computer.ColonyID() + "/computers/" + computer.ID() + "/reject")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return err
	}

	return nil
}

func GetComputersByColonyID(colonyID string, prvKey string) ([]*core.Computer, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Get("https://localhost:8080/colonies/" + colonyID + "/computers")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return nil, err
	}

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	computers, err := core.ConvertJSONToComputerArray(unquotedResp)
	if err != nil {
		return nil, err
	}

	return computers, nil
}

func GetComputerByID(computerID string, colonyID string, prvKey string) (*core.Computer, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Get("https://localhost:8080/colonies/" + colonyID + "/computers/" + computerID)

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return nil, err
	}

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	computer, err := core.ConvertJSONToComputer(unquotedResp)
	if err != nil {
		return nil, err
	}

	return computer, nil
}

func AddProcess(process *core.Process, prvKey string) error {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return err
	}

	processJSON, err := process.ToJSON()
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetBody(processJSON).
		Post("https://localhost:8080/colonies/" + process.TargetColonyID() + "/processes")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return err
	}

	return nil
}

func GetWaitingProcesses(computerID string, colonyID string, count int, prvKey string) ([]*core.Process, error) {
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
		SetHeader("ComputerId", computerID).
		SetHeader("Count", strconv.Itoa(count)).
		SetHeader("State", strconv.Itoa(core.WAITING)).
		Get("https://localhost:8080/colonies/" + colonyID + "/processes")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return processes, err
	}

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return processes, err
	}
	processes, err = core.ConvertJSONToProcessArray(unquotedResp)

	return processes, nil
}

func AssignProcess(computerID string, colonyID string, prvKey string) (*core.Process, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetHeader("ComputerId", computerID).
		Get("https://localhost:8080/colonies/" + colonyID + "/processes/assign")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return nil, err
	}

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}
	process, err := core.ConvertJSONToProcess(unquotedResp)
	if err != nil {
		return nil, err
	}

	return process, nil
}
