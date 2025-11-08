package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

// AddServiceDefinition adds a new ServiceDefinition (requires colony owner privileges)
func (client *ColoniesClient) AddServiceDefinition(sd *core.ServiceDefinition, prvKey string) (*core.ServiceDefinition, error) {
	msg := rpc.CreateAddServiceDefinitionMsg(sd)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddServiceDefinitionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToServiceDefinition(respBodyString)
}

// GetServiceDefinition retrieves a ServiceDefinition by name
func (client *ColoniesClient) GetServiceDefinition(colonyName, name string, prvKey string) (*core.ServiceDefinition, error) {
	msg := rpc.CreateGetServiceDefinitionMsg(colonyName, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetServiceDefinitionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToServiceDefinition(respBodyString)
}

// GetServiceDefinitions retrieves all ServiceDefinitions in a colony
func (client *ColoniesClient) GetServiceDefinitions(colonyName string, prvKey string) ([]*core.ServiceDefinition, error) {
	msg := rpc.CreateGetServiceDefinitionsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetServiceDefinitionsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToServiceDefinitionArray(respBodyString)
}

// RemoveServiceDefinition removes a ServiceDefinition by namespace and name (requires colony owner privileges)
func (client *ColoniesClient) RemoveServiceDefinition(namespace, name string, prvKey string) error {
	msg := rpc.CreateRemoveServiceDefinitionMsg(namespace, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveServiceDefinitionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

// AddService adds a new Service instance
func (client *ColoniesClient) AddService(service *core.Service, prvKey string) (*core.Service, error) {
	msg := rpc.CreateAddServiceMsg(service)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddServicePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToService(respBodyString)
}

// GetService retrieves a Service by namespace and name
func (client *ColoniesClient) GetService(namespace, name string, prvKey string) (*core.Service, error) {
	msg := rpc.CreateGetServiceMsg(namespace, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetServicePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToService(respBodyString)
}

// GetServices retrieves Services by namespace and optionally by kind
func (client *ColoniesClient) GetServices(namespace, kind string, prvKey string) ([]*core.Service, error) {
	msg := rpc.CreateGetServicesMsg(namespace, kind)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetServicesPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToServiceArray(respBodyString)
}

// UpdateService updates an existing Service
func (client *ColoniesClient) UpdateService(service *core.Service, prvKey string) (*core.Service, error) {
	msg := rpc.CreateUpdateServiceMsg(service)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.UpdateServicePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToService(respBodyString)
}

// GetServiceHistory retrieves history for a service
func (client *ColoniesClient) GetServiceHistory(serviceID string, limit int, prvKey string) ([]*core.ServiceHistory, error) {
	msg := rpc.CreateGetServiceHistoryMsg(serviceID, limit)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetServiceHistoryPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToServiceHistoryArray(respBodyString)
}

// RemoveService removes a Service by namespace and name
func (client *ColoniesClient) RemoveService(namespace, name string, prvKey string) error {
	msg := rpc.CreateRemoveServiceMsg(namespace, name)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveServicePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}
