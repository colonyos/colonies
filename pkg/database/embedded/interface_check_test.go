package embedded_test

import (
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/database/embedded"
)

// Compile-time interface compliance check
var _ database.Database = (*embedded.EmbeddedDatabase)(nil)
