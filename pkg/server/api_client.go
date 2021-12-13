package server

import (
	"crypto/tls"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func client() *resty.Client {
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	return client
}

func AddColony() {
	client := client()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"colonyid":"dd", "password":"testpass"}`).
		//SetResult(&AuthSuccess{}). // or SetResult(AuthSuccess{}).
		Post("https://localhost:8080/colonies")
	fmt.Println(err)
	fmt.Println(resp)
}

func GenerateID()
