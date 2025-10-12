package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) Statistics(prvKey string) (*core.Statistics, error) {
	msg := rpc.CreateGetStatisticsMsg()
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetStatisiticsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToStatistics(respBodyString)
}

func (client *ColoniesClient) Version() (string, string, error) {
	msg := rpc.CreateVersionMsg("", "")
	jsonString, err := msg.ToJSON()
	if err != nil {
		return "", "", err
	}

	respBodyString, err := client.sendMessage(rpc.VersionPayloadType, jsonString, "", true, context.TODO())
	if err != nil {
		return "", "", err
	}

	version, err := rpc.CreateVersionMsgFromJSON(respBodyString)
	if err != nil {
		return "", "", err
	}

	return version.BuildVersion, version.BuildTime, nil
}

func (client *ColoniesClient) GetServerInfo() (*core.ServerInfo, error) {
	msg := rpc.CreateGetServerInfoMsg()
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetServerInfoPayloadType, jsonString, "", true, context.TODO())
	if err != nil {
		return nil, err
	}

	serverInfo, err := core.CreateServerInfoFromJSON(respBodyString)
	if err != nil {
		return nil, err
	}

	return serverInfo, nil
}

func (client *ColoniesClient) CheckHealth() error {
	return client.backend.CheckHealth()
}

func (client *ColoniesClient) GetClusterInfo(prvKey string) (*cluster.Config, error) {
	msg := rpc.CreateGetClusterMsg()
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetClusterPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return cluster.ConvertJSONToConfig(respBodyString)
}

func (client *ColoniesClient) ChangeServerID(serverID string, prvKey string) error {
	msg := rpc.CreateChangeServerIDMsg(serverID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.ChangeServerIDPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}