package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) AddLog(processID string, logmsg string, prvKey string) error {
	msg := rpc.CreateAddLogMsg(processID, logmsg)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.AddLogPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) GetLogsByProcess(colonyName string, processID string, count int, prvKey string) ([]*core.Log, error) {
	msg := rpc.CreateGetLogsMsg(colonyName, processID, count, 0)
	msg.ExecutorName = ""
	jsonString, err := msg.ToJSON()
	if err != nil {
		return []*core.Log{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLogsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return []*core.Log{}, err
	}

	return core.ConvertJSONToLogArray(respBodyString)
}

func (client *ColoniesClient) GetLogsByProcessSince(colonyName string, processID string, count int, since int64, prvKey string) ([]*core.Log, error) {
	msg := rpc.CreateGetLogsMsg(colonyName, processID, count, since)
	msg.ExecutorName = ""
	jsonString, err := msg.ToJSON()
	if err != nil {
		return []*core.Log{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLogsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return []*core.Log{}, err
	}

	return core.ConvertJSONToLogArray(respBodyString)
}

func (client *ColoniesClient) SearchLogs(colonyName, text string, days int, count int, prvKey string) ([]*core.Log, error) {
	msg := rpc.CreateSearchLogsMsg(colonyName, text, days, count)
	msg.ColonyName = colonyName
	msg.Text = text
	msg.Days = days
	msg.Count = count
	jsonString, err := msg.ToJSON()
	if err != nil {
		return []*core.Log{}, err
	}

	respBodyString, err := client.sendMessage(rpc.SearchLogsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return []*core.Log{}, err
	}

	return core.ConvertJSONToLogArray(respBodyString)
}