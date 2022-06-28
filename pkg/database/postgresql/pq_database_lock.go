package postgresql

import (
	"errors"
	"time"
)

func (db *PQDatabase) Lock(timeout int) error {
	errChan := make(chan error, 1)
	go func() {
		sqlStatement := `SELECT pg_advisory_lock(1)`
		_, err := db.postgresql.Exec(sqlStatement)
		if err != nil {
			errChan <- err
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return errors.New("lock request timed out")
	}
}

func (db *PQDatabase) Unlock() error {
	sqlStatement := `SELECT pg_advisory_unlock(1)`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}
