package generator

import (
	"fmt"
	"testing"

	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/stretchr/testify/assert"
)

func TestGenerator(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	scheduler := CreateGeneratorScheduler(db)
	fmt.Println(scheduler)
}
