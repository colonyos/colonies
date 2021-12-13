package server

import (
	"colonies/pkg/database"
	. "colonies/pkg/utils"
	"testing"
	"time"
)

func TestAddColony(t *testing.T) {
	db, err := database.PrepareTests()
	CheckError(t, err)

	controller := CreateColoniesController(db)
	apiServer := CreateAPIServer(controller, 8080, "../../cert/key.pem", "../../cert/cert.pem")

	go func() {
		apiServer.ServeForever()
	}()

	time.Sleep(1 * time.Second)

	AddColony()

	done := make(chan bool)
	<-done
}
