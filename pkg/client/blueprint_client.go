package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

// AddBlueprintDefinition adds a new BlueprintDefinition (requires colony owner privileges)
func (client *ColoniesClient) AddBlueprintDefinition(sd *core.BlueprintDefinition, prvKey string) (*core.BlueprintDefinition, error) {
	msg := rpc.CreateAddBlueprintDefinitionMsg(sd)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddBlueprintDefinitionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToBlueprintDefinition(respBodyString)
}

// GetBlueprintDefinition retrieves a BlueprintDefinition by name
func (client *ColoniesClient) GetBlueprintDefinition(colonyName, name string, prvKey string) (*core.BlueprintDefinition, error) {
	msg := rpc.CreateGetBlueprintDefinitionMsg(colonyName, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetBlueprintDefinitionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToBlueprintDefinition(respBodyString)
}

// GetBlueprintDefinitions retrieves all BlueprintDefinitions in a colony
func (client *ColoniesClient) GetBlueprintDefinitions(colonyName string, prvKey string) ([]*core.BlueprintDefinition, error) {
	msg := rpc.CreateGetBlueprintDefinitionsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetBlueprintDefinitionsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToBlueprintDefinitionArray(respBodyString)
}

// RemoveBlueprintDefinition removes a BlueprintDefinition by namespace and name (requires colony owner privileges)
func (client *ColoniesClient) RemoveBlueprintDefinition(namespace, name string, prvKey string) error {
	msg := rpc.CreateRemoveBlueprintDefinitionMsg(namespace, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveBlueprintDefinitionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

// AddBlueprint adds a new Blueprint instance
func (client *ColoniesClient) AddBlueprint(blueprint *core.Blueprint, prvKey string) (*core.Blueprint, error) {
	msg := rpc.CreateAddBlueprintMsg(blueprint)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddBlueprintPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToBlueprint(respBodyString)
}

// GetBlueprint retrieves a Blueprint by namespace and name
func (client *ColoniesClient) GetBlueprint(namespace, name string, prvKey string) (*core.Blueprint, error) {
	msg := rpc.CreateGetBlueprintMsg(namespace, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetBlueprintPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToBlueprint(respBodyString)
}

// GetBlueprints retrieves Blueprints by namespace and optionally by kind
func (client *ColoniesClient) GetBlueprints(namespace, kind string, prvKey string) ([]*core.Blueprint, error) {
	msg := rpc.CreateGetBlueprintsMsg(namespace, kind)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetBlueprintsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToBlueprintArray(respBodyString)
}

// UpdateBlueprint updates an existing Blueprint
func (client *ColoniesClient) UpdateBlueprint(blueprint *core.Blueprint, prvKey string) (*core.Blueprint, error) {
	msg := rpc.CreateUpdateBlueprintMsg(blueprint)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.UpdateBlueprintPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToBlueprint(respBodyString)
}

// GetBlueprintHistory retrieves history for a blueprint
func (client *ColoniesClient) GetBlueprintHistory(blueprintID string, limit int, prvKey string) ([]*core.BlueprintHistory, error) {
	msg := rpc.CreateGetBlueprintHistoryMsg(blueprintID, limit)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetBlueprintHistoryPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToBlueprintHistoryArray(respBodyString)
}

// RemoveBlueprint removes a Blueprint by namespace and name
func (client *ColoniesClient) RemoveBlueprint(namespace, name string, prvKey string) error {
	msg := rpc.CreateRemoveBlueprintMsg(namespace, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveBlueprintPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}
