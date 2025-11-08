package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

// AddResourceDefinition adds a new ResourceDefinition (requires colony owner privileges)
func (client *ColoniesClient) AddResourceDefinition(rd *core.ResourceDefinition, prvKey string) (*core.ResourceDefinition, error) {
	msg := rpc.CreateAddResourceDefinitionMsg(rd)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddResourceDefinitionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToResourceDefinition(respBodyString)
}

// GetResourceDefinition retrieves a ResourceDefinition by name
func (client *ColoniesClient) GetResourceDefinition(colonyName, name string, prvKey string) (*core.ResourceDefinition, error) {
	msg := rpc.CreateGetResourceDefinitionMsg(colonyName, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetResourceDefinitionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToResourceDefinition(respBodyString)
}

// GetResourceDefinitions retrieves all ResourceDefinitions in a colony
func (client *ColoniesClient) GetResourceDefinitions(colonyName string, prvKey string) ([]*core.ResourceDefinition, error) {
	msg := rpc.CreateGetResourceDefinitionsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetResourceDefinitionsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToResourceDefinitionArray(respBodyString)
}

// RemoveResourceDefinition removes a ResourceDefinition by namespace and name (requires colony owner privileges)
func (client *ColoniesClient) RemoveResourceDefinition(namespace, name string, prvKey string) error {
	msg := rpc.CreateRemoveResourceDefinitionMsg(namespace, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveResourceDefinitionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

// AddResource adds a new Resource instance
func (client *ColoniesClient) AddResource(resource *core.Resource, prvKey string) (*core.Resource, error) {
	msg := rpc.CreateAddResourceMsg(resource)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddResourcePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToResource(respBodyString)
}

// GetResource retrieves a Resource by namespace and name
func (client *ColoniesClient) GetResource(namespace, name string, prvKey string) (*core.Resource, error) {
	msg := rpc.CreateGetResourceMsg(namespace, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetResourcePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToResource(respBodyString)
}

// GetResources retrieves Resources by namespace and optionally by kind
func (client *ColoniesClient) GetResources(namespace, kind string, prvKey string) ([]*core.Resource, error) {
	msg := rpc.CreateGetResourcesMsg(namespace, kind)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetResourcesPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToResourceArray(respBodyString)
}

// UpdateResource updates an existing Resource
func (client *ColoniesClient) UpdateResource(resource *core.Resource, prvKey string) (*core.Resource, error) {
	msg := rpc.CreateUpdateResourceMsg(resource)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.UpdateResourcePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToResource(respBodyString)
}

// GetResourceHistory retrieves history for a resource
func (client *ColoniesClient) GetResourceHistory(resourceID string, limit int, prvKey string) ([]*core.ResourceHistory, error) {
	msg := rpc.CreateGetResourceHistoryMsg(resourceID, limit)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetResourceHistoryPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToResourceHistoryArray(respBodyString)
}

// RemoveResource removes a Resource by namespace and name
func (client *ColoniesClient) RemoveResource(namespace, name string, prvKey string) error {
	msg := rpc.CreateRemoveResourceMsg(namespace, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveResourcePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}
