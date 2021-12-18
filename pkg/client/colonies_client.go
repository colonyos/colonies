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
		failure, err := core.CreateFailureFromJSON(jsonString)
		if err != nil {
			return err
		}

		return errors.New(failure.Message())
	}

	return nil
}

func AddColony(colony *core.Colony, apiKey string) error {
	client := client()

	colonyJSON, err := colony.ToJSON()
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Api-Key", apiKey).
		SetBody(colonyJSON).
		Post("https://localhost:8080/colonies")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return err
	}

	return nil
}

func GetColonies(apiKey string) ([]*core.Colony, error) {
	client := client()

	var colonies []*core.Colony
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Api-Key", apiKey).
		Get("https://localhost:8080/colonies")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return colonies, err
	}

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return colonies, err
	}

	colonies, err = core.CreateColonyArrayFromJSON(unquotedResp)
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

func AddWorker(worker *core.Worker, prvKey string) error {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return err
	}

	workerJSON, err := worker.ToJSON()
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		SetBody(workerJSON).
		Post("https://localhost:8080/colonies/" + worker.ColonyID() + "/workers")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return err
	}

	return nil
}

func GetWorkersByColonyID(colonyID string, prvKey string) ([]*core.Worker, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Get("https://localhost:8080/colonies/" + colonyID + "/workers")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return nil, err
	}

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	workers, err := core.CreateWorkerArrayFromJSON(unquotedResp)
	if err != nil {
		return nil, err
	}

	return workers, nil
}

func GetWorkerByID(workerID string, colonyID string, prvKey string) (*core.Worker, error) {
	client := client()
	digest, sig, id, err := security.GenerateCredentials(prvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Id", id).
		SetHeader("Digest", digest).
		SetHeader("Signature", sig).
		Get("https://localhost:8080/colonies/" + colonyID + "/workers/" + workerID)

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return nil, err
	}

	unquotedResp, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return nil, err
	}

	worker, err := core.CreateWorkerFromJSON(unquotedResp)
	if err != nil {
		return nil, err
	}

	return worker, nil
}
