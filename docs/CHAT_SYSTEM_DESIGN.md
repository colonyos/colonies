# ColonyOS Chat System Design

## Architecture Overview

```
┌─────────────────────────────────────────┐
│            Chat Server                  │
│  - User auth, sessions, UI              │
│  - Own database (PostgreSQL/MongoDB)    │
│  - Calls ColonyOS API                   │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│            ColonyOS                     │
│  - Process orchestration                │
│  - Stream transport (stdin/stdout)      │
│  - Executor scheduling                  │
└────────────────┬────────────────────────┘
                 │
    ┌────────────┼────────────┐
    ▼            ▼            ▼
┌────────┐  ┌────────┐  ┌────────┐
│Ollama  │  │Ollama  │  │Ollama  │
│Edge    │  │Cloud   │  │On-prem │
│(GPU)   │  │(A100)  │  │(RTX)   │
└────────┘  └────────┘  └────────┘
```

## Design Principles

- **Loose coupling**: ColonyOS knows nothing about chat, users, or sessions
- **Process streams**: Only feature addition to ColonyOS
- **Stateless executors**: All state in DB streams, executor can crash/restart
- **Horizontal scaling**: Ollama executors across compute continuum

---

## 1. ColonyOS Implementation Plan

### 1.1 Database Schema

```sql
CREATE TABLE process_streams (
    id SERIAL,
    process_id TEXT NOT NULL,
    stream TEXT NOT NULL,           -- 'stdin', 'stdout', 'stderr'
    sequence BIGINT NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    data BYTEA NOT NULL,
    PRIMARY KEY (process_id, stream, sequence)
);

CREATE INDEX idx_process_streams_process ON process_streams(process_id);
CREATE INDEX idx_process_streams_timestamp ON process_streams(timestamp);

-- Convert to hypertable for time-series optimization
SELECT create_hypertable('process_streams', 'timestamp');
```

### 1.2 Core Package Changes

**pkg/core/stream.go**
```go
type StreamEntry struct {
    ProcessID string    `json:"processId"`
    Stream    string    `json:"stream"`    // stdin, stdout, stderr
    Sequence  int64     `json:"sequence"`
    Timestamp time.Time `json:"timestamp"`
    Data      []byte    `json:"data"`
}

type StreamPosition struct {
    ProcessID string `json:"processId"`
    Stream    string `json:"stream"`
    Sequence  int64  `json:"sequence"`
}
```

### 1.3 Database Interface

**pkg/database/streams.go**
```go
type StreamDatabase interface {
    WriteStream(processID, stream string, data []byte) (int64, error)
    ReadStream(processID, stream string, fromSeq int64, limit int) ([]*StreamEntry, error)
    GetStreamPosition(processID, stream string) (int64, error)
    DeleteStreams(processID string) error
}
```

### 1.4 RPC Messages

ColonyOS uses signed RPC messages over multiple transports (HTTP/Gin, gRPC, CoAP, LibP2P).

**pkg/rpc/stream_msgs.go**
```go
// Write to stdin
type WriteStdinMsg struct {
    ProcessID string `json:"processid"`
    Data      []byte `json:"data"`
}

// Read stdout/stderr
type ReadStreamMsg struct {
    ProcessID string `json:"processid"`
    Stream    string `json:"stream"`  // "stdout" or "stderr"
    FromSeq   int64  `json:"fromseq"`
    Limit     int    `json:"limit"`
}

// Stream entries response
type StreamEntriesMsg struct {
    Entries []*core.StreamEntry `json:"entries"`
}

// Subscribe to stream (WebSocket)
type SubscribeStreamMsg struct {
    ProcessID string `json:"processid"`
    Stream    string `json:"stream"`
}
```

**PayloadTypes:**
- `writestdinmsg` - Write data to stdin
- `readstdoutmsg` - Read stdout entries
- `readstderrmsg` - Read stderr entries
- `subscribestdoutmsg` - WebSocket subscription for stdout
- `subscribestderrmsg` - WebSocket subscription for stderr

### 1.5 Handler Implementation

**pkg/server/handlers/stream/handlers.go**
```go
func (h *Handlers) HandleWriteStdin(rpcMsg *rpc.RPCMsg, recoveredID string) (*rpc.RPCReplyMsg, error) {
    msg, err := rpc.CreateWriteStdinMsgFromJSON(rpcMsg.DecodePayload())
    // ... validate, write to DB, notify
}

func (h *Handlers) HandleReadStdout(rpcMsg *rpc.RPCMsg, recoveredID string) (*rpc.RPCReplyMsg, error) {
    msg, err := rpc.CreateReadStreamMsgFromJSON(rpcMsg.DecodePayload())
    // ... validate, read from DB, return entries
}
```

### 1.6 WebSocket Subscription

Uses existing WebSocket pattern in `pkg/server/handlers/realtime/`:
```go
func (h *Handlers) HandleSubscribeStdout(rpcMsg *rpc.RPCMsg, recoveredID string, ws *websocket.Conn) error {
    // Subscribe to PostgreSQL NOTIFY
    // Forward stream entries to WebSocket
}
```

### 1.6 PostgreSQL NOTIFY for Real-time

```sql
-- Trigger on insert
CREATE OR REPLACE FUNCTION notify_stream_insert()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('stream_' || NEW.process_id, NEW.stream || ':' || NEW.sequence);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER stream_insert_trigger
AFTER INSERT ON process_streams
FOR EACH ROW EXECUTE FUNCTION notify_stream_insert();
```

### 1.7 CLI Commands

```bash
# Write to stdin
colonies process stdin --processid <pid> --data "Hello"
colonies process stdin --processid <pid> --file input.txt

# Read stdout
colonies process stdout --processid <pid>
colonies process stdout --processid <pid> --follow
colonies process stdout --processid <pid> --from 100 --limit 50

# Read stderr
colonies process stderr --processid <pid>
```

### 1.8 Client SDK

**pkg/client/streams.go**
```go
func (c *Client) WriteStdin(processID string, data []byte, prvKey string) (int64, error)
func (c *Client) ReadStdout(processID string, fromSeq int64, limit int, prvKey string) ([]*StreamEntry, error)
func (c *Client) ReadStderr(processID string, fromSeq int64, limit int, prvKey string) ([]*StreamEntry, error)
func (c *Client) StreamStdout(processID string, prvKey string) (<-chan *StreamEntry, error)
func (c *Client) ConnectStreams(processID string, prvKey string) (*StreamConnection, error)
```

### 1.9 Implementation Tasks

- [ ] Create database migration for process_streams table
- [ ] Implement StreamDatabase interface for PostgreSQL
- [ ] Add stream handlers to server
- [ ] Implement WebSocket handler with NOTIFY/LISTEN
- [ ] Add CLI commands
- [ ] Update client SDK
- [ ] Write tests
- [ ] Update documentation

---

## 2. Ollama Executor Implementation Plan

### 2.1 Executor Structure

```
ollama-executor/
├── cmd/
│   └── main.go
├── pkg/
│   └── executor/
│       ├── executor.go
│       ├── ollama_client.go
│       └── chat_handler.go
├── Dockerfile
├── go.mod
└── README.md
```

### 2.2 Supported Functions

| Function | Description | KwArgs |
|----------|-------------|--------|
| `chat` | Interactive chat session | `model`, `system_prompt` |
| `generate` | One-shot generation | `model`, `prompt` |
| `embeddings` | Generate embeddings | `model`, `text` |

### 2.3 Chat Handler Implementation

**pkg/executor/chat_handler.go**
```go
func (e *Executor) handleChat(process *core.Process) {
    model := process.FunctionSpec.KwArgs["model"].(string)
    systemPrompt, _ := process.FunctionSpec.KwArgs["system_prompt"].(string)

    // Read existing conversation from stdout (for crash recovery)
    history := e.reconstructHistory(process.ID)

    // Build initial context
    messages := []OllamaMessage{}
    if systemPrompt != "" {
        messages = append(messages, OllamaMessage{Role: "system", Content: systemPrompt})
    }
    messages = append(messages, history...)

    // Main chat loop
    for {
        // Wait for user input from stdin
        userInput, err := e.waitForStdin(process.ID)
        if err != nil {
            break // Session closed
        }

        // Add user message
        messages = append(messages, OllamaMessage{Role: "user", Content: userInput})

        // Call Ollama with streaming
        response := ""
        for chunk := range e.ollamaClient.ChatStream(model, messages) {
            response += chunk
            // Stream each token to stdout
            e.client.WriteStdin(process.ID, []byte(chunk))
        }

        // Mark end of response
        e.client.WriteStdin(process.ID, []byte("\n[END_RESPONSE]\n"))

        // Add to context
        messages = append(messages, OllamaMessage{Role: "assistant", Content: response})
    }

    e.client.Close(process.ID, true) // Success
}
```

### 2.4 Ollama Client

**pkg/executor/ollama_client.go**
```go
type OllamaClient struct {
    baseURL string
    client  *http.Client
}

type OllamaMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func (c *OllamaClient) ChatStream(model string, messages []OllamaMessage) <-chan string {
    ch := make(chan string)

    go func() {
        defer close(ch)

        req := map[string]interface{}{
            "model":    model,
            "messages": messages,
            "stream":   true,
        }

        resp, _ := c.client.Post(c.baseURL+"/api/chat", "application/json", toJSON(req))
        defer resp.Body.Close()

        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            var chunk struct {
                Message struct {
                    Content string `json:"content"`
                } `json:"message"`
                Done bool `json:"done"`
            }
            json.Unmarshal(scanner.Bytes(), &chunk)

            if chunk.Message.Content != "" {
                ch <- chunk.Message.Content
            }
            if chunk.Done {
                break
            }
        }
    }()

    return ch
}
```

### 2.5 Crash Recovery

```go
func (e *Executor) reconstructHistory(processID string) []OllamaMessage {
    // Read all stdout entries
    entries, _ := e.client.ReadStdout(processID, 0, 10000)

    // Parse conversation from stream
    messages := []OllamaMessage{}
    var currentRole string
    var currentContent strings.Builder

    for _, entry := range entries {
        text := string(entry.Data)
        // Parse markers to reconstruct roles and content
        // ... parsing logic
    }

    return messages
}
```

### 2.6 Configuration

```yaml
executor:
  name: "ollama-executor"
  type: "ollama"
  colony: "dev"

ollama:
  url: "http://localhost:11434"

models:
  - llama3
  - codellama
  - mistral
```

### 2.7 Implementation Tasks

- [ ] Create executor project structure
- [ ] Implement Ollama client with streaming
- [ ] Implement chat handler with stdin/stdout
- [ ] Add crash recovery (history reconstruction)
- [ ] Add generate and embeddings functions
- [ ] Create Dockerfile
- [ ] Write tests
- [ ] Create example blueprints

---

## 3. Chat Application Implementation Plan

### 3.1 Application Structure

```
chat-app/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers.go
│   │   ├── websocket.go
│   │   └── middleware.go
│   ├── auth/
│   │   └── auth.go
│   ├── session/
│   │   ├── session.go
│   │   └── repository.go
│   └── colonies/
│       └── client.go
├── web/                    # Frontend (optional)
├── migrations/
├── docker-compose.yml
└── README.md
```

### 3.2 Database Schema

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Chat sessions
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    process_id TEXT,                    -- ColonyOS process ID
    title TEXT,
    model TEXT NOT NULL,
    system_prompt TEXT,
    status TEXT DEFAULT 'active',       -- active, closed
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_process ON sessions(process_id);

-- Session messages (for quick listing, not full content)
CREATE TABLE session_summaries (
    session_id UUID REFERENCES sessions(id),
    message_count INT DEFAULT 0,
    last_message_preview TEXT,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 3.3 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register user |
| POST | `/api/auth/login` | Login, get JWT |
| GET | `/api/sessions` | List user's sessions |
| POST | `/api/sessions` | Create new session |
| GET | `/api/sessions/{id}` | Get session details |
| DELETE | `/api/sessions/{id}` | Close/delete session |
| POST | `/api/sessions/{id}/message` | Send message |
| WS | `/api/sessions/{id}/stream` | Real-time stream |

### 3.4 Session Creation Flow

```go
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserID(r.Context())

    var req CreateSessionRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 1. Create ColonyOS process
    funcSpec := core.CreateEmptyFunctionSpec()
    funcSpec.FuncName = "chat"
    funcSpec.Conditions.ExecutorType = "ollama"
    funcSpec.KwArgs = map[string]interface{}{
        "model":         req.Model,
        "system_prompt": req.SystemPrompt,
    }

    process, err := h.coloniesClient.Submit(funcSpec, h.prvKey)

    // 2. Create session record
    session := &Session{
        UserID:       userID,
        ProcessID:    process.ID,
        Title:        req.Title,
        Model:        req.Model,
        SystemPrompt: req.SystemPrompt,
    }
    h.sessionRepo.Create(session)

    json.NewEncoder(w).Encode(session)
}
```

### 3.5 Message Handling

```go
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
    sessionID := chi.URLParam(r, "id")

    var req SendMessageRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Get session
    session, _ := h.sessionRepo.Get(sessionID)

    // Verify ownership
    if session.UserID != auth.GetUserID(r.Context()) {
        http.Error(w, "forbidden", 403)
        return
    }

    // Write to ColonyOS stdin
    h.coloniesClient.WriteStdin(session.ProcessID, []byte(req.Message), h.prvKey)

    // Update summary
    h.sessionRepo.UpdateSummary(sessionID, req.Message)

    w.WriteHeader(http.StatusAccepted)
}
```

### 3.6 WebSocket Streaming

```go
func (h *Handler) StreamSession(w http.ResponseWriter, r *http.Request) {
    sessionID := chi.URLParam(r, "id")
    session, _ := h.sessionRepo.Get(sessionID)

    // Upgrade to WebSocket
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()

    // Connect to ColonyOS streams
    streamConn, _ := h.coloniesClient.ConnectStreams(session.ProcessID, h.prvKey)
    defer streamConn.Close()

    // Bidirectional proxy
    go func() {
        // Client -> ColonyOS (stdin)
        for {
            _, msg, err := conn.ReadMessage()
            if err != nil {
                break
            }
            streamConn.WriteStdin(msg)
        }
    }()

    // ColonyOS -> Client (stdout)
    for entry := range streamConn.Stdout() {
        conn.WriteMessage(websocket.TextMessage, entry.Data)
    }
}
```

### 3.7 Frontend Considerations

**React/Vue/Svelte component structure:**
```
components/
├── SessionList.vue         # List of user's chats
├── ChatWindow.vue          # Main chat interface
├── MessageBubble.vue       # Individual message
├── StreamingResponse.vue   # Token-by-token display
└── ModelSelector.vue       # Choose model
```

**WebSocket handling:**
```javascript
const ws = new WebSocket(`wss://chat.example.com/api/sessions/${sessionId}/stream`);

ws.onmessage = (event) => {
    // Append token to current response
    currentResponse.value += event.data;
};

function sendMessage(text) {
    ws.send(JSON.stringify({ type: 'message', content: text }));
}
```

### 3.8 Implementation Tasks

- [ ] Set up project structure
- [ ] Implement database migrations
- [ ] Implement auth (JWT)
- [ ] Implement session CRUD
- [ ] Implement ColonyOS client wrapper
- [ ] Implement WebSocket proxy
- [ ] Build API handlers
- [ ] Create frontend (optional)
- [ ] Write tests
- [ ] Dockerize
- [ ] Write documentation

---

## 4. Deployment Architecture

```yaml
# docker-compose.yml
version: '3.8'

services:
  # Chat application
  chat-app:
    build: ./chat-app
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgres://...
      - COLONIES_URL=http://colonies-server:8080
      - COLONIES_PRVKEY=...
    depends_on:
      - postgres
      - colonies-server

  # Chat app database
  postgres:
    image: postgres:15
    volumes:
      - chat_data:/var/lib/postgresql/data

  # ColonyOS
  colonies-server:
    image: colonyos/colonies:latest
    ports:
      - "8080:8080"
    depends_on:
      - timescaledb

  # ColonyOS database
  timescaledb:
    image: timescale/timescaledb:latest-pg16
    volumes:
      - colonies_data:/var/lib/postgresql/data

  # Ollama executor
  ollama-executor:
    build: ./ollama-executor
    environment:
      - COLONIES_URL=http://colonies-server:8080
      - OLLAMA_URL=http://ollama:11434
    depends_on:
      - colonies-server
      - ollama

  # Ollama server
  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_models:/root/.ollama
    deploy:
      resources:
        reservations:
          devices:
            - capabilities: [gpu]

volumes:
  chat_data:
  colonies_data:
  ollama_models:
```

---

## 5. Future Enhancements

- **RAG support**: Add vector database (pgvector) for document retrieval
- **Multi-model**: Route to different executors based on model (Claude, GPT via API)
- **Function calling**: Let LLM call ColonyOS functions
- **Conversation branching**: Fork conversations
- **Sharing**: Share sessions between users
- **Export**: Export conversations to markdown/PDF

---

## 6. Implementation Order

1. **Phase 1: ColonyOS Streams** (1-2 weeks)
   - Database schema
   - Core stream handlers
   - CLI commands
   - Client SDK

2. **Phase 2: Ollama Executor** (1 week)
   - Basic chat function
   - Streaming support
   - Crash recovery

3. **Phase 3: Chat App MVP** (1-2 weeks)
   - Auth
   - Session management
   - WebSocket streaming
   - Basic UI

4. **Phase 4: Polish** (ongoing)
   - Error handling
   - Monitoring
   - Documentation
   - Additional features
