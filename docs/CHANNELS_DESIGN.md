# ColonyOS Channels Design

## Overview

Channels provide bidirectional message passing between users and executors, with ColonyOS acting as a relay server. This enables interactive patterns like chat, Jupyter kernels, and any request/response or streaming workload.

## Architecture

```
┌──────────┐         ┌──────────────┐         ┌──────────┐
│  User/   │ ──────► │   ColonyOS   │ ──────► │ Executor │
│  Client  │ ◄────── │   (relay)    │ ◄────── │          │
└──────────┘         └──────────────┘         └──────────┘
                           │
                    In-memory only
                    No persistence
                    Replicated across servers
```

## Core Concepts

### Channels as Append-Only Logs
- Messages are immutable once appended
- Server assigns sequence numbers
- Multiple readers, tagged senders
- Simple replication (just append to all replicas)
- Automatic cleanup when process completes

### Process Integration
- Channels defined in FunctionSpec
- Auto-created when process is submitted
- Scoped to process lifecycle
- Accessed by name via Process

---

## Data Structures

### FunctionSpec with Channels

```go
type FunctionSpec struct {
    FuncName    string        `json:"funcname"`
    // ... existing fields
    Channels    []ChannelSpec `json:"channels,omitempty"`
}

type ChannelSpec struct {
    Name string `json:"name"`
}
```

### Channel

```go
type Channel struct {
    ID        string
    ProcessID string
    Name      string
    Sequence  int64       // Server-assigned counter
    Log       []LogEntry
    mu        sync.RWMutex
}

type LogEntry struct {
    Sequence  int64     `json:"sequence"`
    Timestamp time.Time `json:"timestamp"`
    SenderID  string    `json:"senderid"`
    Payload   []byte    `json:"payload"`
}
```

### Process with Channels

```go
type Process struct {
    ID           string
    FunctionSpec FunctionSpec
    // ... existing fields
    Channels     map[string]string `json:"channels"` // name → channelID
}
```

---

## API Design

### Channel Operations

```go
// Get channel by process and name
GetChannel(processID string, name string) → channelID

// Append message (server assigns sequence)
Append(channelID string, payload []byte) → sequence int64

// Read entries after sequence
ReadAfter(channelID string, afterSeq int64, limit int) → []LogEntry
```

### Automatic Channel Creation

When process is submitted, channels are auto-created:

```go
func (h *Handlers) HandleSubmit(funcSpec *FunctionSpec) (*Process, error) {
    process := createProcess(funcSpec)
    process.Channels = make(map[string]string)

    // Auto-create channels from spec
    for _, chSpec := range funcSpec.Channels {
        channelID := generateID()
        h.channelRouter.Create(&Channel{
            ID:        channelID,
            ProcessID: process.ID,
            Name:      chSpec.Name,
        })
        process.Channels[chSpec.Name] = channelID
    }

    return process, nil
}
```

---

## RPC Messages

### GetChannelMsg

```go
type GetChannelMsg struct {
    ProcessID string `json:"processid"`
    Name      string `json:"name"`
}

type GetChannelReplyMsg struct {
    ChannelID string `json:"channelid"`
}
```

**PayloadType:** `getchannelmsg`

### AppendMsg

```go
type AppendMsg struct {
    ChannelID string `json:"channelid"`
    Payload   []byte `json:"payload"`
}

type AppendReplyMsg struct {
    Sequence int64 `json:"sequence"`
}
```

**PayloadType:** `appendmsg`

### ReadAfterMsg

```go
type ReadAfterMsg struct {
    ChannelID string `json:"channelid"`
    AfterSeq  int64  `json:"afterseq"`
    Limit     int    `json:"limit"`
}

type ReadAfterReplyMsg struct {
    Entries []LogEntry `json:"entries"`
}
```

**PayloadType:** `readaftermsg`

---

## Server Implementation

### Channel Router

```go
type ChannelRouter struct {
    mu       sync.RWMutex
    channels map[string]*Channel
}

func (r *ChannelRouter) Create(channel *Channel) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.channels[channel.ID] = channel
}

func (r *ChannelRouter) Append(channelID string, senderID string, payload []byte) (int64, error) {
    r.mu.Lock()
    channel := r.channels[channelID]
    r.mu.Unlock()

    if channel == nil {
        return 0, ErrChannelNotFound
    }

    channel.mu.Lock()
    defer channel.mu.Unlock()

    // Server assigns sequence
    channel.Sequence++
    entry := LogEntry{
        Sequence:  channel.Sequence,
        Timestamp: time.Now(),
        SenderID:  senderID,
        Payload:   payload,
    }
    channel.Log = append(channel.Log, entry)

    // Replicate to peers
    go r.replicateToPeers(channelID, entry)

    return channel.Sequence, nil
}

func (r *ChannelRouter) ReadAfter(channelID string, afterSeq int64, limit int) ([]LogEntry, error) {
    r.mu.RLock()
    channel := r.channels[channelID]
    r.mu.RUnlock()

    if channel == nil {
        return nil, ErrChannelNotFound
    }

    channel.mu.RLock()
    defer channel.mu.RUnlock()

    var result []LogEntry
    for _, entry := range channel.Log {
        if entry.Sequence > afterSeq {
            result = append(result, entry)
            if limit > 0 && len(result) >= limit {
                break
            }
        }
    }

    return result, nil
}
```

### Cleanup on Process Completion

```go
func (r *ChannelRouter) CleanupProcess(processID string) {
    r.mu.Lock()
    defer r.mu.Unlock()

    for id, channel := range r.channels {
        if channel.ProcessID == processID {
            delete(r.channels, id)
        }
    }

    // Notify peers to cleanup
    r.replicateCleanup(processID)
}
```

---

## Distributed Replication

### Server-Assigned Sequences

Since server assigns sequence numbers, replication is simple:

```go
func (r *ChannelRouter) replicateToPeers(channelID string, entry LogEntry) {
    for _, peer := range r.peers {
        go func(p *Peer) {
            p.ReplicateEntry(channelID, entry)
        }(peer)
    }
}

// Peer receives replicated entry
func (r *ChannelRouter) ReplicateEntry(channelID string, entry LogEntry) {
    r.mu.Lock()
    channel := r.channels[channelID]
    r.mu.Unlock()

    if channel == nil {
        // Create channel if doesn't exist (late replication)
        channel = &Channel{ID: channelID}
        r.channels[channelID] = channel
    }

    channel.mu.Lock()
    defer channel.mu.Unlock()

    // Idempotent - check if already have this sequence
    for _, e := range channel.Log {
        if e.Sequence == entry.Sequence {
            return // Already have it
        }
    }

    channel.Log = append(channel.Log, entry)

    // Keep sorted by sequence
    sort.Slice(channel.Log, func(i, j int) bool {
        return channel.Log[i].Sequence < channel.Log[j].Sequence
    })
}
```

### Handling Multiple ColonyOS Servers

Challenge: Client writes to Server A, executor reads from Server B.

Solution: **Leader per channel with routing**

```go
// On channel creation, register leader in etcd
func (r *ChannelRouter) Create(channel *Channel) {
    r.channels[channel.ID] = channel
    etcd.Put("channels/"+channel.ID+"/leader", r.serverID)
}

// On append, route to leader
func (s *Server) HandleAppend(channelID string, payload []byte) (int64, error) {
    leader := etcd.Get("channels/" + channelID + "/leader")

    if leader == s.id {
        // We're the leader - append locally
        return s.router.Append(channelID, senderID, payload)
    } else {
        // Forward to leader
        return s.forwardAppend(leader, channelID, payload)
    }
}
```

This ensures:
- One server assigns sequences (no conflicts)
- All reads can go to any replica
- Simple replication

---

## Client SDK

### Go Client

```go
func (c *Client) GetChannel(processID string, name string, prvKey string) (string, error) {
    msg := &rpc.GetChannelMsg{ProcessID: processID, Name: name}
    reply, err := c.sendRPC("getchannelmsg", msg, prvKey)
    return reply.ChannelID, err
}

func (c *Client) Append(channelID string, payload []byte, prvKey string) (int64, error) {
    msg := &rpc.AppendMsg{ChannelID: channelID, Payload: payload}
    reply, err := c.sendRPC("appendmsg", msg, prvKey)
    return reply.Sequence, err
}

func (c *Client) ReadAfter(channelID string, afterSeq int64, limit int, prvKey string) ([]LogEntry, error) {
    msg := &rpc.ReadAfterMsg{ChannelID: channelID, AfterSeq: afterSeq, Limit: limit}
    reply, err := c.sendRPC("readaftermsg", msg, prvKey)
    return reply.Entries, err
}
```

### Channel Wrapper

```go
type Channel struct {
    id     string
    cursor int64
    client *Client
    prvKey string
}

func (ch *Channel) Send(payload interface{}) (int64, error) {
    data, _ := json.Marshal(payload)
    seq, err := ch.client.Append(ch.id, data, ch.prvKey)
    return seq, err
}

func (ch *Channel) Receive(limit int) ([]LogEntry, error) {
    entries, err := ch.client.ReadAfter(ch.id, ch.cursor, limit, ch.prvKey)
    if len(entries) > 0 {
        ch.cursor = entries[len(entries)-1].Sequence
    }
    return entries, err
}

func (ch *Channel) WaitForMessage(timeout time.Duration) (*LogEntry, error) {
    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        entries, err := ch.Receive(1)
        if err != nil {
            return nil, err
        }
        if len(entries) > 0 {
            return &entries[0], nil
        }
        time.Sleep(50 * time.Millisecond)
    }

    return nil, ErrTimeout
}
```

### JavaScript Client

```javascript
class ColoniesClient {
    async getChannel(processId, name) {
        const reply = await this.rpc('getchannelmsg', {
            processid: processId,
            name: name
        });
        return reply.channelid;
    }

    async append(channelId, payload) {
        const reply = await this.rpc('appendmsg', {
            channelid: channelId,
            payload: JSON.stringify(payload)
        });
        return reply.sequence;
    }

    async readAfter(channelId, afterSeq, limit = 100) {
        const reply = await this.rpc('readaftermsg', {
            channelid: channelId,
            afterseq: afterSeq,
            limit: limit
        });
        return reply.entries.map(e => ({
            sequence: e.sequence,
            senderId: e.senderid,
            payload: JSON.parse(e.payload),
            timestamp: new Date(e.timestamp)
        }));
    }
}

// Convenience wrapper
class Channel {
    constructor(client, channelId) {
        this.client = client;
        this.channelId = channelId;
        this.cursor = 0;
    }

    async send(payload) {
        return await this.client.append(this.channelId, payload);
    }

    async receive(limit = 100) {
        const entries = await this.client.readAfter(
            this.channelId,
            this.cursor,
            limit
        );
        if (entries.length > 0) {
            this.cursor = entries[entries.length - 1].sequence;
        }
        return entries;
    }

    async poll(callback, interval = 100) {
        while (true) {
            const entries = await this.receive();
            for (const entry of entries) {
                callback(entry);
            }
            await sleep(interval);
        }
    }
}
```

---

## CLI Commands

```bash
# Get channel ID
colonies channel get --processid <pid> --name input

# Send message
colonies channel send --channelid <cid> --data '{"text":"hello"}'
colonies channel send --processid <pid> --name input --data '{"text":"hello"}'

# Read messages
colonies channel read --channelid <cid>
colonies channel read --channelid <cid> --after 10 --limit 50

# Follow (poll continuously)
colonies channel follow --channelid <cid>
colonies channel follow --processid <pid> --name output
```

---

## Use Cases

### 1. Chat with Ollama

```go
// Submit with channels
funcSpec := &FunctionSpec{
    FuncName: "chat",
    Channels: []ChannelSpec{
        {Name: "input"},
        {Name: "output"},
    },
    KwArgs: map[string]interface{}{
        "model": "llama3",
    },
}
process := client.Submit(funcSpec)

// Get channels
inputCh := client.GetChannel(process.ID, "input")
outputCh := client.GetChannel(process.ID, "output")

// Send message
client.Append(inputCh, {text: "Hello!"})

// Read response
cursor := int64(0)
for {
    entries := client.ReadAfter(outputCh, cursor, 100)
    for _, entry := range entries {
        fmt.Print(entry.Payload.text)
        cursor = entry.Sequence
    }
    time.Sleep(100 * time.Millisecond)
}
```

**Executor:**

```go
func (e *Executor) handleChat(process *Process) {
    inputCh := e.client.GetChannel(process.ID, "input")
    outputCh := e.client.GetChannel(process.ID, "output")

    cursor := int64(0)
    for {
        entries := e.client.ReadAfter(inputCh, cursor, 10)

        for _, entry := range entries {
            msg := parseMessage(entry.Payload)

            // Stream response tokens
            for token := range e.ollama.Stream(msg.text) {
                e.client.Append(outputCh, {type: "token", text: token})
            }
            e.client.Append(outputCh, {type: "done"})

            cursor = entry.Sequence
        }

        time.Sleep(100 * time.Millisecond)
    }
}
```

### 2. Jupyter Kernel

```go
// Submit kernel with multiple channels
funcSpec := &FunctionSpec{
    FuncName: "python-kernel",
    Channels: []ChannelSpec{
        {Name: "shell"},    // execute requests
        {Name: "iopub"},    // outputs
        {Name: "control"},  // interrupt/shutdown
    },
}
process := client.Submit(funcSpec)

// Execute code
shell := client.GetChannel(process.ID, "shell")
iopub := client.GetChannel(process.ID, "iopub")

client.Append(shell, {
    type: "execute_request",
    code: "print('hello')\n2+2",
})

// Stream outputs
cursor := int64(0)
for {
    entries := client.ReadAfter(iopub, cursor, 100)
    for _, entry := range entries {
        switch entry.Payload.type {
        case "stream":
            fmt.Print(entry.Payload.text)
        case "execute_result":
            display(entry.Payload.data)
        }
        cursor = entry.Sequence
    }
}
```

### 3. Remote Terminal

```go
funcSpec := &FunctionSpec{
    FuncName: "terminal",
    Channels: []ChannelSpec{
        {Name: "stdin"},
        {Name: "stdout"},
        {Name: "stderr"},
    },
}
process := client.Submit(funcSpec)

stdin := client.GetChannel(process.ID, "stdin")
stdout := client.GetChannel(process.ID, "stdout")

// Send command
client.Append(stdin, []byte("ls -la\n"))

// Read output
// ...
```

---

## Implementation Plan

### Phase 1: Core Infrastructure (1 week)

**Tasks:**
- [ ] Add Channels field to FunctionSpec
- [ ] Add Channels field to Process
- [ ] Define RPC message types in `pkg/rpc/channel_msgs.go`
- [ ] Implement ChannelRouter in `pkg/server/channel_router.go`
- [ ] Create handlers in `pkg/server/handlers/channel/`
- [ ] Auto-create channels on process submit
- [ ] Cleanup channels on process completion
- [ ] Unit tests

**Files:**
```
pkg/rpc/channel_msgs.go
pkg/rpc/channel_msgs_test.go
pkg/server/channel_router.go
pkg/server/handlers/channel/handlers.go
pkg/server/handlers/channel/handlers_test.go
```

### Phase 2: Distributed Replication (1 week)

**Tasks:**
- [ ] Implement leader election per channel (etcd)
- [ ] Implement routing to leader
- [ ] Implement async replication to followers
- [ ] Handle leader failover
- [ ] Integration tests with multiple servers

### Phase 3: Client SDKs (1 week)

**Tasks:**
- [ ] Go client in `pkg/client/channels.go`
- [ ] JavaScript/TypeScript client
- [ ] Python client (if exists)
- [ ] Client tests

### Phase 4: CLI Commands (2-3 days)

**Tasks:**
- [ ] `colonies channel get`
- [ ] `colonies channel send`
- [ ] `colonies channel read`
- [ ] `colonies channel follow`

**Files:**
```
internal/cli/channel.go
```

### Phase 5: Ollama Executor (1 week)

**Tasks:**
- [ ] Create executor project
- [ ] Implement Ollama client with streaming
- [ ] Implement chat handler using channels
- [ ] Dockerfile
- [ ] Examples

### Phase 6: Chat Application (1-2 weeks)

**Tasks:**
- [ ] User auth
- [ ] Session management
- [ ] WebSocket proxy to channels
- [ ] REST API
- [ ] Basic UI

---

## Security Model

### Crypto-Based Authorization

ColonyOS uses cryptographic signatures for access control - no ACLs needed.

**Process ownership:**
```go
type Process struct {
    ID          string
    SubmitterID string  // Derived from signature on submit
    ExecutorID  string  // Assigned executor
    Channels    map[string]string
}
```

**Access rules:**
- Only **submitter** and **assigned executor** can access channels
- Identity derived from cryptographic signature
- No configuration needed - identity IS the authorization

**Authorization check:**
```go
func (h *Handlers) authorizeChannelAccess(channelID string, callerID string) error {
    channel := h.router.Get(channelID)
    process := h.db.GetProcess(channel.ProcessID)

    // Only submitter or assigned executor can access
    if callerID != process.SubmitterID && callerID != process.ExecutorID {
        return ErrUnauthorized
    }

    return nil
}
```

**Flow:**
1. User submits function spec (signed with private key)
2. Server recovers user ID from signature → `SubmitterID`
3. Process assigned to executor → `ExecutorID`
4. All channel operations verify: `callerID == SubmitterID || callerID == ExecutorID`

### Additional Security Considerations

1. **Rate limiting**: Prevent message flooding per identity
2. **Payload size**: Max message size limit (e.g., 1MB)
3. **Channel limits**: Max channels per process
4. **Memory limits**: Max log size per channel

---

## Timeline Summary

| Phase | Duration | Description |
|-------|----------|-------------|
| 1 | 1 week | Core infrastructure |
| 2 | 1 week | Distributed replication |
| 3 | 1 week | Client SDKs |
| 4 | 2-3 days | CLI commands |
| 5 | 1 week | Ollama executor |
| 6 | 1-2 weeks | Chat application |

**Total: 6-8 weeks**
