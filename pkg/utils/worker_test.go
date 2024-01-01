package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	pool := NewWorkerPool(50).Start()

	calls := 100
	aggErrChan := make(chan error, calls)

	for i := 0; i < calls; i++ {
		errChan := pool.Call(func() error {
			return nil
		})

		go func() {
			err := <-errChan
			aggErrChan <- err
		}()
	}

	expectedErrs := calls
	counter := 0
O:
	for {
		select {
		case err := <-aggErrChan:
			if err != nil {
				t.Error(err)
			}
			counter++
			expectedErrs--
			if expectedErrs == 0 {
				break O
			}
		}
	}

	assert.Equal(t, counter, calls)
	pool.Stop()

}
