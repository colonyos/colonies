package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) AddFile(file *core.File, prvKey string) (*core.File, error) {
	msg := rpc.CreateAddFileMsg(file)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFile(respBodyString)
}

func (client *ColoniesClient) GetFileByID(colonyName string, fileID string, prvKey string) ([]*core.File, error) {
	msg := rpc.CreateGetFileMsg(colonyName, fileID, "", "", false)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFileArray(respBodyString)
}

func (client *ColoniesClient) GetLatestFileByName(colonyName string, label string, name string, prvKey string) ([]*core.File, error) {
	msg := rpc.CreateGetFileMsg(colonyName, "", label, name, true)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFileArray(respBodyString)
}

func (client *ColoniesClient) GetFileByName(colonyName string, label string, name string, prvKey string) ([]*core.File, error) {
	msg := rpc.CreateGetFileMsg(colonyName, "", label, name, false)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFileArray(respBodyString)
}

func (client *ColoniesClient) GetFileData(colonyName string, label string, prvKey string) ([]*core.FileData, error) {
	msg := rpc.CreateGetFilesMsg(colonyName, label)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFilesPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFileDataArray(respBodyString)
}

func (client *ColoniesClient) GetFileLabels(colonyName string, prvKey string) ([]*core.Label, error) {
	msg := rpc.CreateGetAllFileLabelsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFileLabelsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	labels, err := core.ConvertJSONToLabelArray(respBodyString)
	if err != nil {
		return nil, err
	}

	return labels, err
}

func (client *ColoniesClient) GetFileLabelsByName(colonyName string, name string, exact bool, prvKey string) ([]*core.Label, error) {
	msg := rpc.CreateGetFileLabelsMsg(colonyName, name, exact)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFileLabelsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	labels, err := core.ConvertJSONToLabelArray(respBodyString)
	if err != nil {
		return nil, err
	}

	return labels, err
}

func (client *ColoniesClient) RemoveFileByID(colonyName string, fileID string, prvKey string) error {
	msg := rpc.CreateRemoveFileMsg(colonyName, fileID, "", "")
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RemoveFileByName(colonyName string, label string, name string, prvKey string) error {
	msg := rpc.CreateRemoveFileMsg(colonyName, "", label, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveFilePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}