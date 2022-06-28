package server

import (
	"time"

	"github.com/colonyos/colonies/pkg/database"
)

func tryBecomeLeader(leaderChan chan bool, timeout int, db database.Database) error {
	for {
		time.Sleep(1 * time.Second)
		err := db.Lock(timeout)
		if err == nil {
			leaderChan <- true
		}
	}
}
