package client

import (
	"context"
	"encoding/json"

	"github.com/colonyos/colonies/pkg/channel"
	"github.com/colonyos/colonies/pkg/rpc"
)

// ChannelAppend appends a message to a channel
func (client *ColoniesClient) ChannelAppend(processID string, channelName string, sequence int64, inReplyTo int64, payload []byte, prvKey string) error {
	msg := rpc.CreateChannelAppendMsg(processID, channelName, sequence, inReplyTo, payload)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.ChannelAppendPayloadType, jsonString, prvKey, false, context.TODO())
	return err
}

// ChannelRead reads messages from a channel after a given index
func (client *ColoniesClient) ChannelRead(processID string, channelName string, afterIndex int64, limit int, prvKey string) ([]*channel.MsgEntry, error) {
	msg := rpc.CreateChannelReadMsg(processID, channelName, afterIndex, limit)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.ChannelReadPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	var entries []*channel.MsgEntry
	err = json.Unmarshal([]byte(respBodyString), &entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}
