# ColonyOS Channels Design

## Overview

Channels provide bidirectional message passing between process submitters and executors, with ColonyOS acting as a relay server. This enables interactive patterns like chat, Jupyter kernels, and any request/response or streaming workload.

## Architecture

```
+----------+         +--------------+         +----------+
|  User/   | ------> |   ColonyOS   | ------> | Executor |
|  Client  | <------ |    Server    | <------ |          |
+----------+         +--------------+         +----------+
                           |
                    In-memory only
                    No persistence
                    Single server per channel
```

**Note**: Channels are scoped to a single ColonyOS server. Both the client and executor must connect to the same server for channel communication. In Kubernetes deployments, use sticky sessions or consistent routing to ensure this.

## Core Concepts

### Channels as Append-Only Logs
- Messages are immutable once appended
- Client assigns sequence numbers (per-sender ordering)
- InReplyTo field enables request-response correlation
- Multiple readers, tagged senders
- Automatic cleanup when process completes

### Process Integration
- Channels defined in FunctionSpec as string array
- Auto-created when process is submitted
- Deterministic channel IDs: `processID_channelName`
- Scoped to process lifecycle
- Accessed by name via process ID

---

## Data Structures

### FunctionSpec with Channels

```go
type FunctionSpec struct {
    FuncName    string   `json:"funcname"`
    // ... existing fields
    Channels    []string `json:"channels,omitempty"`
}
```

### Channel

```go
type Channel struct {
    ID          string      `json:"id"`          // Deterministic: processID_channelName
    ProcessID   string      `json:"processid"`
    Name        string      `json:"name"`
    SubmitterID string      `json:"submitterid"` // Process submitter
    ExecutorID  string      `json:"executorid"`  // Assigned executor
    Sequence    int64       `json:"sequence"`
    Log         []*MsgEntry `json:"log"`
}

type MsgEntry struct {
    Sequence  int64     `json:"sequence"`            // Client-assigned sequence number
    InReplyTo int64     `json:"inreplyto,omitempty"` // References sequence from other sender
    Timestamp time.Time `json:"timestamp"`
    SenderID  string    `json:"senderid"`
    Payload   []byte    `json:"payload"`
}
```

---

## API Design

### Channel Operations

```go
// Append message with client-assigned sequence
ChannelAppend(processID string, channelName string, sequence int64, inReplyTo int64, payload []byte) error

// Read entries after a given index (position in log)
// limit=0 means no limit
ChannelRead(processID string, channelName string, afterIndex int64, limit int) ([]*MsgEntry, error)

// Subscribe for push notifications (server-side only)
Subscribe(channelID string, callerID string) (chan *MsgEntry, error)
```

### Automatic Channel Creation

When a process is submitted, channels are auto-created with deterministic IDs:

```go
func (controller *ColoniesController) AddProcess(process *core.Process) error {
    // ... add process to database ...

    // Auto-create channels from spec
    if process.FunctionSpec.Channels != nil {
        for _, channelName := range process.FunctionSpec.Channels {
            ch := &channel.Channel{
                ID:          process.ID + "_" + channelName, // Deterministic ID
                ProcessID:   process.ID,
                Name:        channelName,
                SubmitterID: process.InitiatorID,
                ExecutorID:  "", // Set when process is assigned
            }
            controller.channelRouter.Create(ch)
        }
    }

    return nil
}
```

---

## RPC Messages

### ChannelAppendMsg

```go
type ChannelAppendMsg struct {
    ProcessID string `json:"processid"`
    Name      string `json:"name"`
    Sequence  int64  `json:"sequence"`            // Client-assigned sequence number
    InReplyTo int64  `json:"inreplyto,omitempty"` // References sequence from other sender
    Payload   []byte `json:"payload"`
    MsgType   string `json:"msgtype"`
}
```

**PayloadType:** `channelappendmsg`

### ChannelReadMsg

```go
type ChannelReadMsg struct {
    ProcessID string `json:"processid"`
    Name      string `json:"name"`
    AfterSeq  int64  `json:"afterseq"`
    Limit     int    `json:"limit"`
    MsgType   string `json:"msgtype"`
}
```

**PayloadType:** `channelreadmsg`

---

## Server Implementation

### Channel Router

The router manages channels in memory:

```go
type Router struct {
    mu          sync.RWMutex
    channels    map[string]*Channel
    byProcess   map[string][]string          // processID -> []channelID
    subMu       sync.RWMutex
    subscribers map[string][]*Subscriber     // channelID -> subscribers
}
```

### Key Operations

```go
// Append with client-assigned sequence and optional reply reference
func (r *Router) Append(channelID string, senderID string, sequence int64, inReplyTo int64, payload []byte) error {
    // Check authorization
    if err := r.authorize(channel, senderID); err != nil {
        return err
    }

    entry := &MsgEntry{
        Sequence:  sequence, // Client-assigned
        InReplyTo: inReplyTo,
        Timestamp: time.Now(),
        SenderID:  senderID,
        Payload:   payload,
    }
    channel.Log = append(channel.Log, entry)

    // Keep sorted by (SenderID, Sequence) for causal ordering
    sort.Slice(channel.Log, func(i, j int) bool {
        if channel.Log[i].SenderID == channel.Log[j].SenderID {
            return channel.Log[i].Sequence < channel.Log[j].Sequence
        }
        return channel.Log[i].Timestamp.Before(channel.Log[j].Timestamp)
    })

    // Notify subscribers
    r.notifySubscribers(channelID, entry)

    return nil
}

// ReadAfter reads entries starting from a given index
func (r *Router) ReadAfter(channelID string, callerID string, afterIndex int64, limit int) ([]*MsgEntry, error) {
    // Check authorization
    if err := r.authorize(channel, callerID); err != nil {
        return nil, err
    }

    // Return entries from afterIndex to afterIndex+limit
    return channel.Log[afterIndex:endIndex], nil
}
```

### Cleanup on Process Completion

```go
func (r *Router) CleanupProcess(processID string) {
    // Remove all channels for process
    for _, id := range r.byProcess[processID] {
        delete(r.channels, id)
    }
    delete(r.byProcess, processID)

    // Close subscriber channels
    for _, id := range channelIDs {
        for _, sub := range r.subscribers[id] {
            close(sub.ch)
        }
        delete(r.subscribers, id)
    }
}
```

---

## Push-Based Notifications

Subscribers receive real-time push notifications via Go channels:

```go
// Subscribe returns a buffered channel for receiving entries
func (r *Router) Subscribe(channelID string, callerID string) (chan *MsgEntry, error) {
    if err := r.authorize(channel, callerID); err != nil {
        return nil, err
    }

    ch := make(chan *MsgEntry, 100)
    r.subscribers[channelID] = append(r.subscribers[channelID], &Subscriber{ch: ch})
    return ch, nil
}

// Non-blocking notification to avoid slow subscribers blocking writers
func (r *Router) notifySubscribers(channelID string, entry *MsgEntry) {
    for _, sub := range r.subscribers[channelID] {
        select {
        case sub.ch <- entry:
            // Sent
        default:
            // Channel full, skip to avoid blocking
        }
    }
}
```

---

## Client SDK

### Go Client

```go
// Append message to channel
func (client *ColoniesClient) ChannelAppend(
    processID string,
    channelName string,
    sequence int64,
    inReplyTo int64,
    payload []byte,
    prvKey string,
) error

// Read messages from channel
func (client *ColoniesClient) ChannelRead(
    processID string,
    channelName string,
    afterIndex int64,
    limit int,
    prvKey string,
) ([]*channel.MsgEntry, error)
```

### Usage Pattern

```go
// Client-side: maintain sequence counter and poll for responses
sequenceCounter := int64(0)
readIndex := int64(0)

// Send message
sequenceCounter++
client.ChannelAppend(processID, "chat", sequenceCounter, 0, []byte("Hello"), prvKey)

// Poll for response
for {
    entries, _ := client.ChannelRead(processID, "chat", readIndex, 100, prvKey)
    for _, entry := range entries {
        if entry.SenderID != myID {
            fmt.Print(string(entry.Payload))
        }
        readIndex++
    }
    time.Sleep(100 * time.Millisecond)
}
```

---

## Causal Ordering

Messages maintain causal ordering using client-assigned sequence numbers:

1. Each sender maintains their own sequence counter
2. Messages sorted by (SenderID, Sequence) within each sender
3. Timestamps used as tiebreaker between different senders
4. InReplyTo field references another sender's sequence for correlation

### Example Flow

```
Client seq 1:                  "What is 2+2?"
Executor seq 1 (InReplyTo: 1): "4"
Client seq 2:                  "What is 3+3?"
Executor seq 2 (InReplyTo: 2): "6"
```

---

## Use Cases

### 1. Chat with Ollama (Streaming)

```go
// Submit with channels
funcSpec := &core.FunctionSpec{
    FuncName: "ollama-chat",
    Channels: []string{"chat", "control"},
    Kwargs: map[string]interface{}{
        "model": "llama3",
    },
}
process, _ := client.Submit(funcSpec, prvKey)

// Send message
client.ChannelAppend(process.ID, "chat", 1, 0, []byte("Hello!"), prvKey)

// Stream response tokens
readIndex := int64(0)
for {
    entries, _ := client.ChannelRead(process.ID, "chat", readIndex, 100, prvKey)
    for _, entry := range entries {
        if entry.SenderID != myID {
            fmt.Print(string(entry.Payload)) // Print streaming tokens
        }
        readIndex++
    }

    // Check for done signal on control channel
    controlEntries, _ := client.ChannelRead(process.ID, "control", 0, 10, prvKey)
    for _, e := range controlEntries {
        if string(e.Payload) == "done" {
            // Send ack and exit
            client.ChannelAppend(process.ID, "control", 1, e.Sequence, []byte("ack"), prvKey)
            return
        }
    }

    time.Sleep(100 * time.Millisecond)
}
```

**Executor side:**

```go
func (e *Executor) handleChat(process *core.Process) {
    readIndex := int64(0)
    seqCounter := int64(0)

    for {
        entries, _ := e.client.ChannelRead(process.ID, "chat", readIndex, 10, e.prvKey)

        for _, entry := range entries {
            if entry.SenderID == process.InitiatorID {
                // Stream response tokens
                for token := range e.ollama.Stream(string(entry.Payload)) {
                    seqCounter++
                    e.client.ChannelAppend(process.ID, "chat", seqCounter, entry.Sequence,
                        []byte(token), e.prvKey)
                }

                // Send done signal
                seqCounter++
                e.client.ChannelAppend(process.ID, "control", seqCounter, 0,
                    []byte("done"), e.prvKey)
            }
            readIndex++
        }

        time.Sleep(100 * time.Millisecond)
    }
}
```

### 2. Jupyter Kernel

```go
funcSpec := &core.FunctionSpec{
    FuncName: "python-kernel",
    Channels: []string{"shell", "iopub", "control"},
}
process, _ := client.Submit(funcSpec, prvKey)

// Execute code
client.ChannelAppend(process.ID, "shell", 1, 0,
    []byte(`{"type":"execute_request","code":"print('hello')"}`), prvKey)

// Stream outputs from iopub
// ...
```

### 3. Remote Terminal

```go
funcSpec := &core.FunctionSpec{
    FuncName: "terminal",
    Channels: []string{"stdin", "stdout", "stderr"},
}
process, _ := client.Submit(funcSpec, prvKey)

// Send command
client.ChannelAppend(process.ID, "stdin", 1, 0, []byte("ls -la\n"), prvKey)

// Read output from stdout/stderr
// ...
```

---

## Security Model

### Crypto-Based Authorization

ColonyOS uses cryptographic signatures for access control.

**Process ownership:**
```go
type Channel struct {
    SubmitterID string  // Derived from signature on submit
    ExecutorID  string  // Assigned executor
}
```

**Access rules:**
- Only submitter and assigned executor can access channels
- Identity derived from cryptographic signature
- No ACL configuration needed

**Authorization check:**
```go
func (r *Router) authorize(channel *Channel, callerID string) error {
    if callerID != channel.SubmitterID && callerID != channel.ExecutorID {
        return ErrUnauthorized
    }
    return nil
}
```

### Additional Security Considerations

1. **Rate limiting**: Prevent message flooding per identity
2. **Payload size**: Max message size limit
3. **Channel limits**: Max channels per process
4. **Memory limits**: Max log size per channel

---

## Deployment Considerations

### Single Server
Channels work out of the box with single-server deployments.

### Multi-Server / Kubernetes
For multi-server deployments, ensure client and executor connect to the same server:

```yaml
# Traefik example - sticky sessions
http:
  services:
    colonies:
      loadBalancer:
        sticky:
          cookie:
            name: colony_affinity
```

Or use consistent hashing based on process ID to route both parties to the same server.

---

## Key Files

```
pkg/channel/
    types.go           - Core data structures (Channel, MsgEntry)
    router.go          - In-memory channel management

pkg/rpc/
    channel_append_msg.go - Append message type
    channel_read_msg.go   - Read message type

pkg/server/handlers/channel/
    handlers.go        - HTTP request handlers

pkg/server/controllers/
    colonies_controller.go - Channel lifecycle (create, assign, cleanup)

pkg/client/
    channel_client.go  - Go SDK channel methods
```

---

## Design Characteristics

1. **Append-Only Logs**: Channels are immutable message logs
2. **Client-Assigned Sequences**: Clients control ordering for their messages
3. **InReplyTo Correlation**: Request-response patterns via sequence references
4. **Bidirectional Communication**: Both submitter and executor can send/receive
5. **Causal Ordering**: Per-sender sequences maintain message causality
6. **In-Memory**: Fast local access, no persistence
7. **Authorization Built-In**: Access control at channel level
8. **Push Notifications**: Real-time updates via Go channels to subscribers
9. **Deterministic IDs**: Enable consistent routing in multi-server setups
10. **Polling-Based Clients**: Stateless HTTP polling, no WebSocket state
