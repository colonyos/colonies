package core

import (
	. "colonies/pkg/utils"
	"testing"
)

func TestCreateAttribute(t *testing.T) {
	attribute := CreateAttribute(GenerateRandomID(), OUT, "test_key", "test_value")
	if len(attribute.ID()) != 64 {
		Fatal(t, "invalid attribute id length")
	}

	if attribute.AttributeType() != OUT {
		Fatal(t, "invalid attribute type")
	}

	if attribute.Key() != "test_key" {
		Fatal(t, "invalid attribute key")
	}

	if attribute.Value() != "test_value" {
		Fatal(t, "invalid attribute value")
	}
}
