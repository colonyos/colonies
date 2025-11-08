package database

import "github.com/colonyos/colonies/pkg/core"

type ResourceDatabase interface {
	// ResourceDefinition methods
	AddResourceDefinition(rd *core.ResourceDefinition) error
	GetResourceDefinitionByID(id string) (*core.ResourceDefinition, error)
	GetResourceDefinitionByName(namespace, name string) (*core.ResourceDefinition, error)
	GetResourceDefinitions() ([]*core.ResourceDefinition, error)
	GetResourceDefinitionsByNamespace(namespace string) ([]*core.ResourceDefinition, error)
	GetResourceDefinitionsByGroup(group string) ([]*core.ResourceDefinition, error)
	UpdateResourceDefinition(rd *core.ResourceDefinition) error
	RemoveResourceDefinitionByID(id string) error
	RemoveResourceDefinitionByName(namespace, name string) error
	CountResourceDefinitions() (int, error)

	// Service methods
	AddResource(service *core.Service) error
	GetResourceByID(id string) (*core.Service, error)
	GetResourceByName(namespace, name string) (*core.Service, error)
	GetResources() ([]*core.Service, error)
	GetResourcesByNamespace(namespace string) ([]*core.Service, error)
	GetResourcesByKind(kind string) ([]*core.Service, error)
	GetResourcesByNamespaceAndKind(namespace, kind string) ([]*core.Service, error)
	UpdateResource(service *core.Service) error
	UpdateResourceStatus(id string, status map[string]interface{}) error
	RemoveResourceByID(id string) error
	RemoveResourceByName(namespace, name string) error
	RemoveResourcesByNamespace(namespace string) error
	CountResources() (int, error)
	CountResourcesByNamespace(namespace string) (int, error)

	// ResourceHistory methods
	AddResourceHistory(history *core.ResourceHistory) error
	GetResourceHistory(resourceID string, limit int) ([]*core.ResourceHistory, error)
	GetResourceHistoryByGeneration(resourceID string, generation int64) (*core.ResourceHistory, error)
	RemoveResourceHistory(resourceID string) error
}
