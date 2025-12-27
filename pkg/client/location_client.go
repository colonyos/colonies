package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) AddLocation(location *core.Location, prvKey string) (*core.Location, error) {
	msg := rpc.CreateAddLocationMsg(location)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddLocationPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	addedLocation, err := core.ConvertJSONToLocation(respBodyString)
	if err != nil {
		return nil, err
	}

	return addedLocation, nil
}

func (client *ColoniesClient) GetLocation(colonyName string, locationName string, prvKey string) (*core.Location, error) {
	msg := rpc.CreateGetLocationMsg(colonyName, locationName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLocationPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	locationFromServer, err := core.ConvertJSONToLocation(respBodyString)
	if err != nil {
		return nil, err
	}

	return locationFromServer, nil
}

func (client *ColoniesClient) GetLocations(colonyName string, prvKey string) ([]*core.Location, error) {
	msg := rpc.CreateGetLocationsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetLocationsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	locationsFromServer, err := core.ConvertJSONToLocationArray(respBodyString)
	if err != nil {
		return nil, err
	}

	return locationsFromServer, nil
}

func (client *ColoniesClient) RemoveLocation(colonyName string, locationName string, prvKey string) error {
	msg := rpc.CreateRemoveLocationMsg(colonyName, locationName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveLocationPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}
