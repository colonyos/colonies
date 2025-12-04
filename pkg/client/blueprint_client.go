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

// GetBlueprintDefinitionByKind retrieves a BlueprintDefinition by its Kind name
func (client *ColoniesClient) GetBlueprintDefinitionByKind(colonyName, kind string, prvKey string) (*core.BlueprintDefinition, error) {
	definitions, err := client.GetBlueprintDefinitions(colonyName, prvKey)
	if err != nil {
		return nil, err
	}

	for _, sd := range definitions {
		if sd.Spec.Names.Kind == kind {
			return sd, nil
		}
	}

	return nil, nil
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

// UpdateBlueprintWithForce updates an existing Blueprint with optional force generation bump
// If forceGeneration is true, the generation will be incremented even if the spec hasn't changed
func (client *ColoniesClient) UpdateBlueprintWithForce(blueprint *core.Blueprint, forceGeneration bool, prvKey string) (*core.Blueprint, error) {
	msg := rpc.CreateUpdateBlueprintMsgWithForce(blueprint, forceGeneration)
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

// UpdateBlueprintStatus updates only the status field of a blueprint
// This is used by reconcilers to report status without triggering a full update or generation bump
func (client *ColoniesClient) UpdateBlueprintStatus(colonyName, blueprintName string, status map[string]interface{}, prvKey string) error {
	msg := rpc.CreateUpdateBlueprintStatusMsg(colonyName, blueprintName, status)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.UpdateBlueprintStatusPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

// ReconcileBlueprint triggers immediate reconciliation of a blueprint
// The server looks up the executor type from the blueprint's handler configuration
// If force is true, the generation will be bumped to trigger redeployment
func (client *ColoniesClient) ReconcileBlueprint(namespace, name string, force bool, prvKey string) (*core.Process, error) {
	msg := rpc.CreateReconcileBlueprintMsg(namespace, name, force)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.ReconcileBlueprintPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToProcess(respBodyString)
}
