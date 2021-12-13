package server

import (
	"colonies/pkg/database"
	. "colonies/pkg/utils"
	"fmt"
	"testing"
)

func TestCreateColonyController(t *testing.T) {
	db, err := database.PrepareTests()
	CheckError(t, err)

	controller := CreateColoniesController(db)

	fmt.Println(controller)
}
