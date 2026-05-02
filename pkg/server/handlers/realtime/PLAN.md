# Realtime File Subscriptions Plan

## Goal

Extend the realtime subscription primitive (today only `SubscribeProcess` /
`SubscribeProcesses`) with `SubscribeFiles`, so clients can be notified when
files are added, updated, or removed under a label or label prefix in
ColonyFS. The motivating use case is event-driven workflows that need to
fire when something lands in ColonyFS — for example, an inbox notifier
that surfaces new daemon-output files to a web UI without polling, or a
reconciler that runs whenever a config file changes.

## Why

Today the only way to learn that a ColonyFS label has new content is to
poll `GetFileData(colony, label)`. That is fine for low-frequency lookups
but it forces every event-driven consumer to invent its own poll loop and
swallow the latency budget that comes with it. Internal projects have
already started working around the missing primitive by submitting
sentinel processes whose only purpose is to fan out notifications — a
correct workaround, but it ties unrelated lifecycles together (every
notification consumes one process slot).

A first-class file subscription closes this gap. It uses the same
WebSocket plumbing the process subscriptions already use, so the marginal
cost on the server is small.

## Existing infrastructure to reuse

- `pkg/client/realtime_client.go`
  `SubscribeProcesses` / `SubscribeProcess` — WebSocket lifecycle, auth,
  read loop, message decoding. Same pattern can serve files.
- `pkg/rpc/` — payload type registry; we add a `SubscribeFilesPayloadType`.
- `pkg/server/handlers/realtime/handlers.go` — websocket handler entry
  point; adds a branch for the new payload type.
- `pkg/backends.RealtimeSubscription` — the in-memory subscription record
  the server tracks per active websocket. Extends with file-event filter
  fields.
- `pkg/fs/fs_client.go` and `pkg/server/handlers/file/data_handlers.go` —
  the write/delete paths that need to fan out file events to active
  subscribers. Hook is at the post-commit point in the file handlers.

## API shape

### Client (Go)

```
type FileEventKind int
const (
    FileAdded   FileEventKind = 1
    FileUpdated FileEventKind = 2  // new revision of an existing name
    FileRemoved FileEventKind = 3
)

type FileEvent struct {
    Kind       FileEventKind
    ColonyName string
    Label      string
    Name       string
    Size       int64
    Checksum   string
    Revision   int64
    Timestamp  time.Time
}

type FileSubscription struct {
    EventChan <-chan *FileEvent
    ErrorChan <-chan error
    Close     func()
}

func (c *ColoniesClient) SubscribeFiles(
    colonyName string,
    labelPrefix string,    // matches label == prefix OR label starts with prefix + "/"
    kinds       []FileEventKind, // empty = all kinds
    timeout     int,             // seconds; 0 = no timeout
    prvKey      string,
) (*FileSubscription, error)
```

`labelPrefix == ""` matches every label in the colony. `labelPrefix ==
"/home/root/inbox"` matches files written directly under that label and
also under any descendant label like `/home/root/inbox/2026`. This is
the same prefix-match convention `GetFileLabels` already exposes.

### Wire format

A new RPC payload type:

```
SubscribeFilesPayloadType = "subscribefilesmsg"

type SubscribeFilesMsg struct {
    ColonyName  string
    LabelPrefix string
    Kinds       []int
    Timeout     int
}
```

Replies on the websocket use a new envelope:

```
type FileEventMsg struct {
    Event FileEvent
}
```

### TypeScript / colonies-ts

Equivalent `subscribeFiles(...)` returning an `EventSource`-like object
with `onEvent` / `onError` / `close()`. Mirrors how `subscribeProcess`
is exposed today.

## Server-side dispatch

When a file write or remove succeeds, the file data handler emits an
event to the realtime backend:

```
realtimeBackend.PublishFileEvent(FileEvent{
    Kind: FileAdded,
    ColonyName: ..., Label: ..., Name: ..., Size: ..., Checksum: ...,
})
```

The realtime backend keeps an in-memory index of active file
subscriptions keyed by colony, with each entry holding the
`labelPrefix`, the kind filter, and the websocket writer. Publish walks
matching subscriptions and writes the event. This mirrors how process
events are currently fanned out and reuses the same per-subscription
backpressure handling (drop on full buffer, log).

Auth: the subscriber's prvKey must own a colony membership for
`colonyName`. Enforced at subscribe time, no per-event check needed.

## Filtering rules

- **Label prefix match**: `subscriptionLabel == eventLabel` OR
  `eventLabel starts with subscriptionLabel + "/"`. Empty prefix
  matches all.
- **Kind filter**: if `kinds` is empty, all kinds match. Otherwise
  event.Kind must be in the set.
- **Multi-tenancy**: subscriptions are colony-scoped; an event in
  colony A is never delivered to subscribers of colony B.

## Backpressure & lifecycle

- Per-subscription buffered channel (size 256). On overflow, the oldest
  events are dropped and a "lagged" marker is delivered the next time
  the writer drains. Subscribers can detect this and resync via
  `GetFileData`.
- Heartbeat / ping-pong inherits the WebSocket settings used for
  process subscriptions.
- On disconnect, the backend removes the subscription record.
- `Timeout` mirrors `SubscribeProcesses` — when set, the subscription
  auto-closes after that many seconds.

## Persistence

None. File subscriptions are in-memory, single-server, lost on restart
(matches the channel and process-subscription model). Consumers
reconnect and replay state from `GetFileData` if they need durable
delivery.

A future durable-delivery layer (offset-based replay) is out of scope
for v1 — it is a separate feature with its own PLAN.

## Test plan

- Unit: `pkg/server/handlers/realtime/file_handler_test.go` covering
  prefix matching, kind filter, multi-colony isolation, and overflow.
- Integration: `pkg/server/realtime_file_integration_test.go` —
  subscribe, write a file via `pkg/fs/fs_client`, assert event arrives.
- Security: `pkg/server/handlers/realtime/handler_security_test.go`
  gains a `TestSubscribeFilesSecurity` mirroring the existing
  `TestSubscribeProcessesSecurity` (invalid prvKey, foreign colony).
- Load: a thousand concurrent subscriptions, ten thousand writes, no
  leaks. Reuse the existing realtime load-test scaffolding if any.

## Phasing

1. **Wire format + server backend.** Add the RPC type, the server-side
   subscription registry entry, and the publish hook in the file
   handlers. No client API yet — verifiable via direct WebSocket calls
   or a tiny test client.
2. **Go client.** `SubscribeFiles` in `realtime_client.go`, mirroring
   `SubscribeProcesses`. Unit tests that round-trip events.
3. **TypeScript client.** Add to `colonies-ts`. Smoke test against a
   local colony.
4. **Documentation.** New section in `docs/Filesystem.md` (if it
   exists; otherwise extend `docs/Generators.md` or create
   `docs/FileSubscriptions.md`) describing the subscribe pattern with
   curl + Go examples.

## Out of scope (for now)

- Durable replay / offset-based delivery.
- Cross-server fan-out (file subscriptions are single-server like
  channels and process subs).
- Filtering by file metadata beyond label prefix and kind.
- Subscribing to label list changes (label created/deleted) as a
  separate event family — could be a follow-up.

## Open questions

1. **Update vs Add semantics.** ColonyFS supports multiple revisions of
   the same `(label, name)` pair. Is "add" only the first revision and
   subsequent writes are "updates", or do all writes emit `FileAdded`?
   Lean: first write = `FileAdded`, subsequent writes of the same name
   = `FileUpdated`. Subscribers who only care about novelty filter on
   `FileAdded` alone.

2. **Atomic-multi events.** A common write pattern is "drop a directory"
   — many files at once. Should the server batch deliver these as one
   event with an array of files, or fire one event per file? Lean: one
   event per file. Consumers that want batching can debounce
   client-side. Simpler delivery semantics.

3. **Path matching syntax.** Plain prefix today; no globbing. Worth a
   discussion whether to support globs (`/home/root/inbox/*/today.md`)
   later. Current lean: plain prefix only; let consumers do the rest
   client-side.
