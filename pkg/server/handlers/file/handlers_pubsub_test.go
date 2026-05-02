package file

import (
	"context"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/stretchr/testify/assert"
)

// drainEvent reads one event from the bus channel with a generous timeout.
func drainEvent(t *testing.T, ch <-chan *core.FileEvent, d time.Duration) (*core.FileEvent, bool) {
	t.Helper()
	select {
	case ev, ok := <-ch:
		if !ok {
			return nil, false
		}
		return ev, true
	case <-time.After(d):
		return nil, false
	}
}

func expectNoBusEvent(t *testing.T, ch <-chan *core.FileEvent, d time.Duration) {
	t.Helper()
	select {
	case ev := <-ch:
		t.Fatalf("expected no event, got %+v", ev)
	case <-time.After(d):
	}
}

// createMockServerWithBus builds the standard mock plus a real
// in-memory FileEventBus so we can observe what the handlers publish.
// The MockFileDB starts empty so the first AddFile produces a FileAdded
// event; subsequent adds of the same name produce FileUpdated.
func createMockServerWithBus() (*MockServer, *MockContext, backends.FileEventBus) {
	fileDB := &MockFileDB{}
	validator := &MockValidator{}
	bus := backends.NewInMemoryFileEventBus()

	server := &MockServer{
		fileDB:    fileDB,
		validator: validator,
		bus:       bus,
	}

	return server, &MockContext{}, bus
}

// HandleAddFile must publish a FileAdded event for a brand-new
// (colony, label, name) tuple, with all the file metadata populated.
func TestHandleAddFile_PublishesFileAdded(t *testing.T) {
	server, ctx, bus := createMockServerWithBus()
	handlers := NewHandlers(server)

	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, _ := bus.Subscribe("test-colony", "/test", nil, subCtx)

	file := &core.File{
		ColonyName:  "test-colony",
		Label:       "/test/label",
		Name:        "first.txt",
		Size:        128,
		Checksum:    "cs",
		ChecksumAlg: "sha256",
	}
	jsonString, _ := rpc.CreateAddFileMsg(file).ToJSON()
	handlers.HandleAddFile(ctx, "test-user", rpc.AddFilePayloadType, jsonString)
	assert.Nil(t, server.lastError)

	ev, ok := drainEvent(t, evCh, time.Second)
	assert.True(t, ok)
	assert.Equal(t, core.FileAdded, ev.Kind)
	assert.Equal(t, "test-colony", ev.ColonyName)
	assert.Equal(t, "/test/label", ev.Label)
	assert.Equal(t, "first.txt", ev.Name)
	assert.NotEmpty(t, ev.FileID, "FileID must be populated for FileAdded events")
	assert.Equal(t, int64(128), ev.Size)
	assert.Equal(t, "cs", ev.Checksum)
	assert.Equal(t, "sha256", ev.ChecksumAlg)
}

// A second AddFile with the same (colony, label, name) is a new revision
// and must publish FileUpdated rather than FileAdded — subscribers that
// only care about novelty rely on this discrimination.
func TestHandleAddFile_PublishesFileUpdatedOnNewRevision(t *testing.T) {
	server, ctx, bus := createMockServerWithBus()
	handlers := NewHandlers(server)

	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, _ := bus.Subscribe("test-colony", "", nil, subCtx)

	file := &core.File{ColonyName: "test-colony", Label: "/l", Name: "doc.md", Size: 1}
	jsonString, _ := rpc.CreateAddFileMsg(file).ToJSON()
	handlers.HandleAddFile(ctx, "user", rpc.AddFilePayloadType, jsonString)

	ev, _ := drainEvent(t, evCh, time.Second)
	assert.Equal(t, core.FileAdded, ev.Kind)

	// Second add — same name, same label.
	file2 := &core.File{ColonyName: "test-colony", Label: "/l", Name: "doc.md", Size: 2}
	jsonString2, _ := rpc.CreateAddFileMsg(file2).ToJSON()
	handlers.HandleAddFile(&MockContext{}, "user", rpc.AddFilePayloadType, jsonString2)

	ev2, ok := drainEvent(t, evCh, time.Second)
	assert.True(t, ok)
	assert.Equal(t, core.FileUpdated, ev2.Kind, "second revision should be FileUpdated")
	assert.Equal(t, "doc.md", ev2.Name)
}

// HandleRemoveFile (by name) publishes a FileRemoved event with the
// (colony, label, name) coordinates and no file-identifying fields
// (which would be misleading after the file is gone).
func TestHandleRemoveFile_ByName_PublishesFileRemoved(t *testing.T) {
	server, ctx, bus := createMockServerWithBus()
	handlers := NewHandlers(server)

	// Pre-populate with a file so the remove succeeds.
	server.fileDB.files = []*core.File{{
		ID: "id1", ColonyName: "test-colony", Label: "/l", Name: "doomed.md",
	}}

	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, _ := bus.Subscribe("test-colony", "", nil, subCtx)

	jsonString, _ := rpc.CreateRemoveFileMsg("test-colony", "", "/l", "doomed.md").ToJSON()
	handlers.HandleRemoveFile(ctx, "user", rpc.RemoveFilePayloadType, jsonString)
	assert.Nil(t, server.lastError)

	ev, ok := drainEvent(t, evCh, time.Second)
	assert.True(t, ok)
	assert.Equal(t, core.FileRemoved, ev.Kind)
	assert.Equal(t, "test-colony", ev.ColonyName)
	assert.Equal(t, "/l", ev.Label)
	assert.Equal(t, "doomed.md", ev.Name)
	assert.Empty(t, ev.FileID, "FileRemoved must omit FileID")
	assert.Empty(t, ev.Checksum, "FileRemoved must omit Checksum")
}

// HandleRemoveFile (by ID) looks up the file's coordinates BEFORE
// removing so the published event carries the right label+name.
func TestHandleRemoveFile_ByID_PublishesFileRemovedWithCoordinates(t *testing.T) {
	server, ctx, bus := createMockServerWithBus()
	handlers := NewHandlers(server)

	server.fileDB.files = []*core.File{{
		ID: "id1", ColonyName: "test-colony", Label: "/route", Name: "x.md",
	}}

	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, _ := bus.Subscribe("test-colony", "", nil, subCtx)

	jsonString, _ := rpc.CreateRemoveFileMsg("test-colony", "id1", "", "").ToJSON()
	handlers.HandleRemoveFile(ctx, "user", rpc.RemoveFilePayloadType, jsonString)
	assert.Nil(t, server.lastError)

	ev, ok := drainEvent(t, evCh, time.Second)
	assert.True(t, ok)
	assert.Equal(t, core.FileRemoved, ev.Kind)
	assert.Equal(t, "/route", ev.Label)
	assert.Equal(t, "x.md", ev.Name)
}

// When the server's FileEventBus() returns nil, the handlers must
// no-op cleanly — no panics, no goroutines started. Older deployments
// that haven't enabled the bus stay functional.
func TestHandleAddFile_NilBusIsNoop(t *testing.T) {
	server, ctx := createMockServer()
	server.bus = nil
	handlers := NewHandlers(server)

	file := createTestFile()
	jsonString, _ := rpc.CreateAddFileMsg(file).ToJSON()
	// Just shouldn't panic.
	handlers.HandleAddFile(ctx, "user", rpc.AddFilePayloadType, jsonString)
	assert.Nil(t, server.lastError)
}

func TestHandleRemoveFile_NilBusIsNoop(t *testing.T) {
	server, ctx := createMockServer()
	server.bus = nil
	handlers := NewHandlers(server)

	jsonString, _ := rpc.CreateRemoveFileMsg("test-colony", "", "/test/label", "test-file.txt").ToJSON()
	handlers.HandleRemoveFile(ctx, "user", rpc.RemoveFilePayloadType, jsonString)
	assert.Nil(t, server.lastError)
}

// Subscribers in colony A must not see events from colony B. This is the
// security boundary at the bus-routing level (the auth boundary lives in
// the websocket handler).
func TestPublishCrossColonyIsolation(t *testing.T) {
	server, ctx, bus := createMockServerWithBus()
	handlers := NewHandlers(server)

	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	colonyA, _ := bus.Subscribe("colony-A", "", nil, subCtx)
	colonyB, _ := bus.Subscribe("colony-B", "", nil, subCtx)

	jsonString, _ := rpc.CreateAddFileMsg(&core.File{
		ColonyName: "colony-A", Label: "/l", Name: "n",
	}).ToJSON()
	handlers.HandleAddFile(ctx, "user", rpc.AddFilePayloadType, jsonString)

	_, ok := drainEvent(t, colonyA, time.Second)
	assert.True(t, ok, "colony-A subscriber must receive event")
	expectNoBusEvent(t, colonyB, 30*time.Millisecond)
}

// A subscription scoped to a deeper label must only see events for that
// label tree, not for sibling labels in the same colony.
func TestPublishLabelPrefixScopingThroughHandler(t *testing.T) {
	server, ctx, bus := createMockServerWithBus()
	handlers := NewHandlers(server)

	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inboxCh, _ := bus.Subscribe("test-colony", "/home/root/inbox", nil, subCtx)

	// Sibling label — must not be delivered.
	siblingMsg, _ := rpc.CreateAddFileMsg(&core.File{
		ColonyName: "test-colony", Label: "/home/root/outbox", Name: "n",
	}).ToJSON()
	handlers.HandleAddFile(ctx, "user", rpc.AddFilePayloadType, siblingMsg)
	expectNoBusEvent(t, inboxCh, 30*time.Millisecond)

	// In-tree event — must be delivered.
	inTreeMsg, _ := rpc.CreateAddFileMsg(&core.File{
		ColonyName: "test-colony", Label: "/home/root/inbox/2026", Name: "msg.md",
	}).ToJSON()
	handlers.HandleAddFile(&MockContext{}, "user", rpc.AddFilePayloadType, inTreeMsg)

	ev, ok := drainEvent(t, inboxCh, time.Second)
	assert.True(t, ok)
	assert.Equal(t, "/home/root/inbox/2026", ev.Label)
}

// Failing the AddFile path (membership rejected) must NOT publish an
// event — we only publish after the DB write succeeds.
func TestHandleAddFile_DoesNotPublishOnFailure(t *testing.T) {
	server, ctx, bus := createMockServerWithBus()
	server.validator.membershipErr = assertNotNil // any non-nil error rejects the call
	handlers := NewHandlers(server)

	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, _ := bus.Subscribe("test-colony", "", nil, subCtx)

	jsonString, _ := rpc.CreateAddFileMsg(&core.File{
		ColonyName: "test-colony", Label: "/l", Name: "n",
	}).ToJSON()
	handlers.HandleAddFile(ctx, "user", rpc.AddFilePayloadType, jsonString)
	expectNoBusEvent(t, evCh, 30*time.Millisecond)
}

// assertNotNil is a sentinel error used in TestHandleAddFile_DoesNotPublishOnFailure
// — any non-nil error works for the validator-rejection path.
var assertNotNil = errOf("membership rejected")

type sentinelErr string

func (e sentinelErr) Error() string { return string(e) }
func errOf(s string) error          { return sentinelErr(s) }
