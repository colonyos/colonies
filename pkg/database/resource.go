package database

import "github.com/colonyos/colonies/pkg/core"

type BlueprintDatabase interface {
	// BlueprintDefinition methods
	AddBlueprintDefinition(sd *core.BlueprintDefinition) error
	GetBlueprintDefinitionByID(id string) (*core.BlueprintDefinition, error)
	GetBlueprintDefinitionByName(namespace, name string) (*core.BlueprintDefinition, error)
	GetBlueprintDefinitions() ([]*core.BlueprintDefinition, error)
	GetBlueprintDefinitionsByNamespace(namespace string) ([]*core.BlueprintDefinition, error)
	GetBlueprintDefinitionsByGroup(group string) ([]*core.BlueprintDefinition, error)
	UpdateBlueprintDefinition(sd *core.BlueprintDefinition) error
	RemoveBlueprintDefinitionByID(id string) error
	RemoveBlueprintDefinitionByName(namespace, name string) error
	CountBlueprintDefinitions() (int, error)

	// Blueprint methods
	AddBlueprint(blueprint *core.Blueprint) error
	GetBlueprintByID(id string) (*core.Blueprint, error)
	GetBlueprintByName(namespace, name string) (*core.Blueprint, error)
	GetBlueprints() ([]*core.Blueprint, error)
	GetBlueprintsByNamespace(namespace string) ([]*core.Blueprint, error)
	GetBlueprintsByKind(kind string) ([]*core.Blueprint, error)
	GetBlueprintsByNamespaceAndKind(namespace, kind string) ([]*core.Blueprint, error)
	UpdateBlueprint(blueprint *core.Blueprint) error
	UpdateBlueprintStatus(id string, status map[string]interface{}) error
	RemoveBlueprintByID(id string) error
	RemoveBlueprintByName(namespace, name string) error
	RemoveBlueprintsByNamespace(namespace string) error
	CountBlueprints() (int, error)
	CountBlueprintsByNamespace(namespace string) (int, error)

	// BlueprintHistory methods
	AddBlueprintHistory(history *core.BlueprintHistory) error
	GetBlueprintHistory(blueprintID string, limit int) ([]*core.BlueprintHistory, error)
	GetBlueprintHistoryByGeneration(blueprintID string, generation int64) (*core.BlueprintHistory, error)
	RemoveBlueprintHistory(blueprintID string) error
}
