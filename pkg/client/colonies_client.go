package client

import (
	"colonies/pkg/core"
	"colonies/pkg/crypto"
	"colonies/pkg/security"
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
)

func client() *resty.Client {
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	return client
}

func GeneratePrivateKey() (string, error) {
	identify, err := crypto.CreateIdendity()
	if err != nil {
		return "", nil
	}

	return identify.PrivateKeyAsHex(), nil
}

func GenerateID(privateKey string) (string, error) {
	identify, err := crypto.CreateIdendityFromString(privateKey)
	if err != nil {
		return "", nil
	}

	return identify.ID(), nil
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

func GetColony(colonyID string, colonyPrvKey string) (*core.Colony, error) {
	client := client()

	dummyData := security.GenerateRandomString()
	sig, err := security.GenerateSignature(dummyData, colonyPrvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Signature", sig).
		Get("https://localhost:8080/colonies/" + colonyID + "?dummydata=" + dummyData)

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		fmt.Println("ERROR")
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

func AddWorker(worker *core.Worker, colonyPrvKey string) error {
	client := client()

	workerJSON, err := worker.ToJSON()
	if err != nil {
		return err
	}

	digest := security.GenerateRandomString()
	sig, err := security.GenerateSignature(digest, colonyPrvKey)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
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

func GetWorker(workerID string, colonyID string, colonyPrvKey string) (*core.Worker, error) {
	client := client()

	digest := security.GenerateRandomString()
	sig, err := security.GenerateSignature(digest, colonyPrvKey)
	if err != nil {
		return nil, err
	}

	resp, err := client.R().
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
