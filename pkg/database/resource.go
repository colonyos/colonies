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

	// Resource methods
	AddResource(resource *core.Resource) error
	GetResourceByID(id string) (*core.Resource, error)
	GetResourceByName(namespace, name string) (*core.Resource, error)
	GetResources() ([]*core.Resource, error)
	GetResourcesByNamespace(namespace string) ([]*core.Resource, error)
	GetResourcesByKind(kind string) ([]*core.Resource, error)
	GetResourcesByNamespaceAndKind(namespace, kind string) ([]*core.Resource, error)
	UpdateResource(resource *core.Resource) error
	UpdateResourceStatus(id string, status map[string]interface{}) error
	RemoveResourceByID(id string) error
	RemoveResourceByName(namespace, name string) error
	RemoveResourcesByNamespace(namespace string) error
	CountResources() (int, error)
	CountResourcesByNamespace(namespace string) (int, error)
}
