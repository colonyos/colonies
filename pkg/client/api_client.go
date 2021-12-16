package client

import (
	"colonies/pkg/core"
	"colonies/pkg/crypto"
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

	unquotedResponse, err := strconv.Unquote(string(resp.Body()))
	if err != nil {
		return colonies, err
	}

	colonies, err = core.CreateColonyArrayFromJSON(unquotedResponse)
	if err != nil {
		return colonies, err
	}

	return colonies, nil
}

func AddWorker(worker *core.Worker, colonyPrvKey string) error {
	client := client()

	workerJSON, err := worker.ToJSON()
	if err != nil {
		return err
	}

	sig, err := security.GenerateSignature(workerJSON, colonyPrvKey)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Signature", sig).
		SetBody(workerJSON).
		Post("https://localhost:8080/colonies/" + worker.ColonyID() + "/workers")

	err = checkStatusCode(resp.StatusCode(), string(resp.Body()))
	if err != nil {
		return err
	}

	return nil
}
