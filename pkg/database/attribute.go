package database

import "github.com/colonyos/colonies/pkg/core"

type AttributeDatabase interface {
	AddAttribute(attribute core.Attribute) error
	AddAttributes(attribute []core.Attribute) error
	GetAttributeByID(attributeID string) (core.Attribute, error)
	GetAttributesByColonyName(colonyName string) ([]core.Attribute, error)
	GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error)
	GetAttributes(targetID string) ([]core.Attribute, error)
	GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error)
	UpdateAttribute(attribute core.Attribute) error
	RemoveAttributeByID(attributeID string) error
	RemoveAllAttributesByColonyName(colonyName string) error
	RemoveAllAttributesByColonyNameWithState(colonyName string, state int) error
	RemoveAllAttributesByProcessGraphID(processGraphID string) error
	RemoveAllAttributesInProcessGraphsByColonyName(colonyName string) error
	RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName string, state int) error
	RemoveAttributesByTargetID(targetID string, attributeType int) error
	RemoveAllAttributesByTargetID(targetID string) error
	RemoveAllAttributes() error
}