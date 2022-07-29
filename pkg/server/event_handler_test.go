package server

import (
	"context"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestEventHandler(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler()
	errChan := make(chan error)
	go func() {
		errChan <- handler.wait("test_runtime_type", core.WAITING, ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.signal("test_runtime_type", core.WAITING)
	}()
	err := <-errChan
	assert.Nil(t, err) // OK
	allListeners, listeners := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
}

func TestEventHandlerTimeout(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler()
	errChan := make(chan error)
	go func() {
		errChan <- handler.wait("test_runtime_type", core.WAITING, ctx)
	}()
	err := <-errChan
	assert.NotNil(t, err) // Will timeout
	allListeners, listeners := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
}

func TestEventHandlerStop(t *testing.T) {
	handler := createEventHandler()

	handler.stop()

	stopped := false
	for {
		stopped = handler.hasStopped()
		if stopped {
			break
		}
	}
	assert.True(t, stopped)
}

func TestEventHandlerTimeout2(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler()
	errChan := make(chan error)
	go func() {
		errChan <- handler.wait("test_runtime_type", core.WAITING, ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.signal("test_runtime_type2", core.WAITING) // NOTE: we are signaling to another target
	}()
	err := <-errChan
	assert.NotNil(t, err) // Will timeout
	allListeners, listeners := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
}

func TestEventHandlerTimeout3(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler()
	errChan := make(chan error)
	go func() {
		errChan <- handler.wait("test_runtime_type", core.WAITING, ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.signal("test_runtime_type", core.RUNNING) // NOTE: we are signaling to another target
	}()
	err := <-errChan
	assert.NotNil(t, err) // Will timeout
	allListeners, listeners := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
}

func TestEventHandlerMany(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler()

	var errChans []chan error

	for i := 0; i < 10; i++ {
		errChan := make(chan error)
		errChans = append(errChans, errChan)
		go func() {
			errChan <- handler.wait("test_runtime_type", core.WAITING, ctx)
		}()
	}

	go func() {
		errChan := make(chan error)
		errChan <- handler.wait("test_runtime_type2", core.WAITING, ctx)
	}()

	time.Sleep(1000 * time.Millisecond)

	go func() {
		handler.signal("test_runtime_type", core.WAITING)
	}()

	// Wait for listers
	for _, errChan := range errChans {
		err := <-errChan
		assert.Nil(t, err) // OK
	}
	allListeners, listeners := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 1) // Note 1, as test_runtime_type2 was never signaled
	assert.Equal(t, listeners, 0)
}
