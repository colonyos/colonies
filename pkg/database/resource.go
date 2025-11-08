package database

import "github.com/colonyos/colonies/pkg/core"

type ServiceDatabase interface {
	// ServiceDefinition methods
	AddServiceDefinition(sd *core.ServiceDefinition) error
	GetServiceDefinitionByID(id string) (*core.ServiceDefinition, error)
	GetServiceDefinitionByName(namespace, name string) (*core.ServiceDefinition, error)
	GetServiceDefinitions() ([]*core.ServiceDefinition, error)
	GetServiceDefinitionsByNamespace(namespace string) ([]*core.ServiceDefinition, error)
	GetServiceDefinitionsByGroup(group string) ([]*core.ServiceDefinition, error)
	UpdateServiceDefinition(sd *core.ServiceDefinition) error
	RemoveServiceDefinitionByID(id string) error
	RemoveServiceDefinitionByName(namespace, name string) error
	CountServiceDefinitions() (int, error)

	// Service methods
	AddService(service *core.Service) error
	GetServiceByID(id string) (*core.Service, error)
	GetServiceByName(namespace, name string) (*core.Service, error)
	GetServices() ([]*core.Service, error)
	GetServicesByNamespace(namespace string) ([]*core.Service, error)
	GetServicesByKind(kind string) ([]*core.Service, error)
	GetServicesByNamespaceAndKind(namespace, kind string) ([]*core.Service, error)
	UpdateService(service *core.Service) error
	UpdateServiceStatus(id string, status map[string]interface{}) error
	RemoveServiceByID(id string) error
	RemoveServiceByName(namespace, name string) error
	RemoveServicesByNamespace(namespace string) error
	CountServices() (int, error)
	CountServicesByNamespace(namespace string) (int, error)

	// ServiceHistory methods
	AddServiceHistory(history *core.ServiceHistory) error
	GetServiceHistory(serviceID string, limit int) ([]*core.ServiceHistory, error)
	GetServiceHistoryByGeneration(serviceID string, generation int64) (*core.ServiceHistory, error)
	RemoveServiceHistory(serviceID string) error
}
