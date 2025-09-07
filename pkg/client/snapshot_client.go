package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) CreateSnapshot(colonyName string, label string, name string, prvKey string) (*core.Snapshot, error) {
	msg := rpc.CreateCreateSnapshotMsg(colonyName, label, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.CreateSnapshotPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	snapshot, err := core.ConvertJSONToSnapshot(respBodyString)
	if err != nil {
		return nil, err
	}

	return snapshot, err
}

func (client *ColoniesClient) GetSnapshotByID(colonyName string, snapshotID string, prvKey string) (*core.Snapshot, error) {
	msg := rpc.CreateGetSnapshotMsg(colonyName, snapshotID, "")
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetSnapshotPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	snapshot, err := core.ConvertJSONToSnapshot(respBodyString)
	if err != nil {
		return nil, err
	}

	return snapshot, err
}

func (client *ColoniesClient) GetSnapshotByName(colonyName string, name string, prvKey string) (*core.Snapshot, error) {
	msg := rpc.CreateGetSnapshotMsg(colonyName, "", name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetSnapshotPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	snapshot, err := core.ConvertJSONToSnapshot(respBodyString)
	if err != nil {
		return nil, err
	}

	return snapshot, err
}

func (client *ColoniesClient) GetSnapshotsByColonyName(colonyName string, prvKey string) ([]*core.Snapshot, error) {
	msg := rpc.CreateGetSnapshotsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetSnapshotsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	snapshots, err := core.ConvertJSONToSnapshotsArray(respBodyString)
	if err != nil {
		return nil, err
	}

	return snapshots, err
}

func (client *ColoniesClient) RemoveSnapshotByID(colonyName string, snapshotID string, prvKey string) error {
	msg := rpc.CreateRemoveSnapshotMsg(colonyName, snapshotID, "")
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveSnapshotPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return err
}

func (client *ColoniesClient) RemoveSnapshotByName(colonyName string, name string, prvKey string) error {
	msg := rpc.CreateRemoveSnapshotMsg(colonyName, "", name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveSnapshotPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return err
}

func (client *ColoniesClient) RemoveAllSnapshots(colonyName string, prvKey string) error {
	msg := rpc.CreateRemoveAllSnapshotsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveAllSnapshotsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return err
}