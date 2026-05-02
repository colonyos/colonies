package realtime_test

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/stretchr/testify/assert"
)

// drainFileEvent reads one event from the subscription with a generous
// timeout. Returns (nil, false) on timeout.
func drainFileEvent(t *testing.T, sub interface {
	GetEventChan() chan *core.FileEvent
	GetErrChan() chan error
}, d time.Duration) (*core.FileEvent, error) {
	t.Helper()
	select {
	case ev := <-sub.GetEventChan():
		return ev, nil
	case err := <-sub.GetErrChan():
		return nil, err
	case <-time.After(d):
		return nil, nil
	}
}

// fileSubAdapter exposes the subscription's channels through getter
// methods so we can pass it to a generic helper without caring whether
// it's a *client.FileSubscription or another type.
type fileSubAdapter struct {
	ev chan *core.FileEvent
	er chan error
}

func (a *fileSubAdapter) GetEventChan() chan *core.FileEvent { return a.ev }
func (a *fileSubAdapter) GetErrChan() chan error             { return a.er }

func newTestFile(colonyName, label, name string) *core.File {
	return &core.File{
		ColonyName:  colonyName,
		Label:       label,
		Name:        name,
		Size:        128,
		Checksum:    "deadbeef",
		ChecksumAlg: "sha256",
	}
}

// End-to-end happy path: subscribe via websocket, server publishes from
// HandleAddFile, client receives the FileEvent intact.
func TestSubscribeFiles_DeliversAddedEvent(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv1(t)
	srv.EnableFileEventBus()

	sub, err := client.SubscribeFiles(env.Colony1Name, "/inbox", nil, 30, env.Executor1PrvKey)
	assert.Nil(t, err)
	defer sub.Close()

	// Brief wait so the websocket subscription is registered with the bus
	// before we publish. The bus has no replay; events published before
	// subscribe are gone.
	time.Sleep(200 * time.Millisecond)

	file := newTestFile(env.Colony1Name, "/inbox", "x.md")
	_, err = client.AddFile(file, env.Executor1PrvKey)
	assert.Nil(t, err)

	adapter := &fileSubAdapter{ev: sub.EventChan, er: sub.ErrChan}
	ev, derr := drainFileEvent(t, adapter, 3*time.Second)
	assert.Nil(t, derr)
	assert.NotNil(t, ev)
	assert.Equal(t, core.FileAdded, ev.Kind)
	assert.Equal(t, env.Colony1Name, ev.ColonyName)
	assert.Equal(t, "/inbox", ev.Label)
	assert.Equal(t, "x.md", ev.Name)
	assert.NotEmpty(t, ev.FileID)

	srv.Shutdown()
	<-done
}

// A second AddFile of the same (label, name) produces FileUpdated. The
// matcher in the bus is still happy with this kind because we asked for
// nil kinds (= all kinds).
func TestSubscribeFiles_DeliversUpdatedEvent(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv1(t)
	srv.EnableFileEventBus()

	sub, err := client.SubscribeFiles(env.Colony1Name, "", nil, 30, env.Executor1PrvKey)
	assert.Nil(t, err)
	defer sub.Close()
	time.Sleep(200 * time.Millisecond)

	first := newTestFile(env.Colony1Name, "/inbox", "doc.md")
	_, err = client.AddFile(first, env.Executor1PrvKey)
	assert.Nil(t, err)
	adapter := &fileSubAdapter{ev: sub.EventChan, er: sub.ErrChan}
	ev, derr := drainFileEvent(t, adapter, 3*time.Second)
	assert.Nil(t, derr)
	assert.Equal(t, core.FileAdded, ev.Kind)

	// Second revision of the same name.
	second := newTestFile(env.Colony1Name, "/inbox", "doc.md")
	second.Size = 256
	_, err = client.AddFile(second, env.Executor1PrvKey)
	assert.Nil(t, err)
	ev2, derr := drainFileEvent(t, adapter, 3*time.Second)
	assert.Nil(t, derr)
	assert.NotNil(t, ev2)
	assert.Equal(t, core.FileUpdated, ev2.Kind)
	assert.Equal(t, "doc.md", ev2.Name)

	srv.Shutdown()
	<-done
}

// Kind filter: subscribe to FileAdded only, write file and remove it,
// confirm the FileRemoved event is NOT delivered.
func TestSubscribeFiles_KindFilter(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv1(t)
	srv.EnableFileEventBus()

	sub, err := client.SubscribeFiles(env.Colony1Name, "", []int{int(core.FileAdded)}, 30, env.Executor1PrvKey)
	assert.Nil(t, err)
	defer sub.Close()
	time.Sleep(200 * time.Millisecond)

	file := newTestFile(env.Colony1Name, "/inbox", "transient.md")
	added, err := client.AddFile(file, env.Executor1PrvKey)
	assert.Nil(t, err)

	adapter := &fileSubAdapter{ev: sub.EventChan, er: sub.ErrChan}
	ev, derr := drainFileEvent(t, adapter, 3*time.Second)
	assert.Nil(t, derr)
	assert.Equal(t, core.FileAdded, ev.Kind)

	err = client.RemoveFileByID(env.Colony1Name, added.ID, env.Executor1PrvKey)
	assert.Nil(t, err)

	// Removed event must NOT arrive within a reasonable window.
	ev2, _ := drainFileEvent(t, adapter, 500*time.Millisecond)
	assert.Nil(t, ev2, "kind filter should have suppressed FileRemoved event")

	srv.Shutdown()
	<-done
}

// Label-prefix scoping: subscribing to /inbox does not see writes under
// /outbox in the same colony.
func TestSubscribeFiles_LabelPrefixScoping(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv1(t)
	srv.EnableFileEventBus()

	sub, err := client.SubscribeFiles(env.Colony1Name, "/inbox", nil, 30, env.Executor1PrvKey)
	assert.Nil(t, err)
	defer sub.Close()
	time.Sleep(200 * time.Millisecond)

	// Off-prefix write — must NOT be delivered.
	off := newTestFile(env.Colony1Name, "/outbox", "n.md")
	_, err = client.AddFile(off, env.Executor1PrvKey)
	assert.Nil(t, err)

	adapter := &fileSubAdapter{ev: sub.EventChan, er: sub.ErrChan}
	ev, _ := drainFileEvent(t, adapter, 500*time.Millisecond)
	assert.Nil(t, ev, "off-prefix event must not be delivered")

	// On-prefix descendant write — must arrive.
	on := newTestFile(env.Colony1Name, "/inbox/2026", "msg.md")
	_, err = client.AddFile(on, env.Executor1PrvKey)
	assert.Nil(t, err)

	ev2, derr := drainFileEvent(t, adapter, 3*time.Second)
	assert.Nil(t, derr)
	assert.NotNil(t, ev2)
	assert.Equal(t, "/inbox/2026", ev2.Label)

	srv.Shutdown()
	<-done
}

// Cross-colony isolation at the websocket layer: a subscriber in colony 2
// does not see events from colony 1.
func TestSubscribeFiles_CrossColonyIsolation(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv1(t)
	srv.EnableFileEventBus()

	// Subscribe on colony 2 with executor 2's key (member of colony 2).
	sub, err := client.SubscribeFiles(env.Colony2Name, "", nil, 30, env.Executor2PrvKey)
	assert.Nil(t, err)
	defer sub.Close()
	time.Sleep(200 * time.Millisecond)

	// Write a file into colony 1.
	file := newTestFile(env.Colony1Name, "/l", "isolated.md")
	_, err = client.AddFile(file, env.Executor1PrvKey)
	assert.Nil(t, err)

	adapter := &fileSubAdapter{ev: sub.EventChan, er: sub.ErrChan}
	ev, _ := drainFileEvent(t, adapter, 500*time.Millisecond)
	assert.Nil(t, ev, "colony-2 subscriber must not see colony-1 events")

	srv.Shutdown()
	<-done
}

// Auth boundary: subscribing to a colony without membership must fail.
// We try to subscribe to colony 1 with executor 2's prvKey (member of
// colony 2 only) — the server should reject.
func TestSubscribeFiles_RejectsForeignColony(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv1(t)
	srv.EnableFileEventBus()

	sub, err := client.SubscribeFiles(env.Colony1Name, "", nil, 30, env.Executor2PrvKey)
	if err != nil {
		// Some auth paths fail at handshake — that's acceptable, the
		// rejection is what we want.
		assert.NotNil(t, err)
		srv.Shutdown()
		<-done
		return
	}
	defer sub.Close()

	// Or it may fail post-subscribe via the error channel.
	select {
	case rejected := <-sub.ErrChan:
		assert.NotNil(t, rejected)
	case ev := <-sub.EventChan:
		t.Fatalf("unexpected event for unauthorised subscriber: %+v", ev)
	case <-time.After(2 * time.Second):
		t.Fatal("expected rejection on errChan; got nothing")
	}

	srv.Shutdown()
	<-done
}

// Multiple subscribers in the same colony with different filters all
// receive the events that match their filter, and only those.
func TestSubscribeFiles_MultipleSubscribers(t *testing.T) {
	env, client, srv, _, done := server.SetupTestEnv1(t)
	srv.EnableFileEventBus()

	wide, err := client.SubscribeFiles(env.Colony1Name, "", nil, 30, env.Executor1PrvKey)
	assert.Nil(t, err)
	defer wide.Close()

	narrow, err := client.SubscribeFiles(env.Colony1Name, "/inbox", nil, 30, env.Executor1PrvKey)
	assert.Nil(t, err)
	defer narrow.Close()

	time.Sleep(200 * time.Millisecond)

	file := newTestFile(env.Colony1Name, "/inbox", "shared.md")
	_, err = client.AddFile(file, env.Executor1PrvKey)
	assert.Nil(t, err)

	wideAdapter := &fileSubAdapter{ev: wide.EventChan, er: wide.ErrChan}
	narrowAdapter := &fileSubAdapter{ev: narrow.EventChan, er: narrow.ErrChan}

	ev, derr := drainFileEvent(t, wideAdapter, 3*time.Second)
	assert.Nil(t, derr)
	assert.Equal(t, "/inbox", ev.Label)

	ev2, derr := drainFileEvent(t, narrowAdapter, 3*time.Second)
	assert.Nil(t, derr)
	assert.Equal(t, "/inbox", ev2.Label)

	srv.Shutdown()
	<-done
}
