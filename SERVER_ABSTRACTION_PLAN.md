# Server Abstraction Layer Implementation Plan

## Overview

This document outlines a comprehensive plan to implement an abstraction layer for the ColonyOS server, enabling support for multiple server backends including HTTP frameworks (Gin, Echo, etc.), libP2P for peer-to-peer communication, gRPC, and other protocols. It also covers the corresponding client abstraction layer to support multiple communication backends.

## Current Architecture Issues

1. **Tight Coupling**: Handlers are tightly coupled to Gin's `*gin.Context`
2. **Single Backend**: Only HTTP/Gin backend is supported
3. **WebSocket Limitation**: WebSocket implementation is HTTP-specific
4. **Scalability**: Cannot run multiple protocols simultaneously
5. **Testing**: Hard to mock backend layer
6. **Client Limitations**: Client only supports HTTP communication

## Proposed Architecture

### Core Design Principles

1. **Backend Agnostic**: Business logic independent of server backend
2. **Multi-Backend Support**: Run HTTP, P2P, gRPC, WebSocket simultaneously  
3. **Client-Server Symmetry**: Both client and server support multiple backends
4. **Backward Compatibility**: Existing functionality continues during migration
5. **Extensibility**: Easy to add new backend protocols
6. **Performance**: Minimal overhead from abstraction

## Phase 1: Core Backend Abstraction (Week 1-2)

### 1.1 Create Server Backend Interfaces

**File**: `pkg/service/backend/interfaces.go`

```go
// Request represents an incoming request regardless of transport
type Request interface {
    GetBody() ([]byte, error)
    GetHeader(key string) string
    GetMethod() string
    GetPath() string
    GetRemoteAddr() string
    GetQuery(key string) string
    GetParam(key string) string
}

// Response represents an outgoing response
type Response interface {
    SetStatus(code int)
    SetHeader(key, value string)
    Write(data []byte) error
    WriteJSON(data interface{}) error
    WriteString(data string) error
    GetStatus() int
}

// Context provides request/response context abstraction
type Context interface {
    Request() Request
    Response() Response
    Set(key string, value interface{})
    Get(key string) (interface{}, bool)
    Abort()
    IsAborted() bool
    Clone() Context
}

// HandlerFunc represents a generic request handler
type HandlerFunc func(ctx Context)

// MiddlewareFunc represents middleware function
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Backend defines the server backend interface
type Backend interface {
    Start(addr string) error
    Stop() error
    Handle(method, path string, handler HandlerFunc)
    Use(middleware MiddlewareFunc)
    SetCORS(config CORSConfig)
    Name() string
    Type() BackendType
}

// BackendType defines the type of backend
type BackendType string

const (
    HTTPBackend   BackendType = "http"
    GRPCBackend   BackendType = "grpc"
    LibP2PBackend BackendType = "libp2p"
)

// CORSConfig defines CORS configuration
type CORSConfig struct {
    AllowOrigins     []string
    AllowMethods     []string
    AllowHeaders     []string
    AllowCredentials bool
}
```

### 1.2 Gin HTTP Backend Implementation

**File**: `pkg/service/backend/gin/gin_backend.go`

```go
type GinBackend struct {
    engine *gin.Engine
    server *http.Server
    addr   string
}

type GinContext struct {
    ginCtx *gin.Context
    req    *GinRequest
    resp   *GinResponse
}

type GinRequest struct {
    ctx *gin.Context
}

type GinResponse struct {
    ctx *gin.Context
}

func NewGinBackend() *GinBackend {
    return &GinBackend{
        engine: gin.Default(),
    }
}

func (g *GinBackend) Type() BackendType {
    return HTTPBackend
}

func (g *GinBackend) Start(addr string) error {
    g.addr = addr
    g.server = &http.Server{
        Addr:    addr,
        Handler: g.engine,
    }
    return g.server.ListenAndServe()
}

func (g *GinBackend) Handle(method, path string, handler HandlerFunc) {
    g.engine.Handle(method, path, func(c *gin.Context) {
        ctx := &GinContext{
            ginCtx: c,
            req:    &GinRequest{ctx: c},
            resp:   &GinResponse{ctx: c},
        }
        handler(ctx)
    })
}

// Implement all interface methods...
```

### 1.3 Backend Factory

**File**: `pkg/service/backend/factory.go`

```go
type BackendConfig struct {
    Type     BackendType            `yaml:"type"`
    Name     string                 `yaml:"name"`
    Address  string                 `yaml:"address"`
    Settings map[string]interface{} `yaml:"settings"`
}

type BackendFactory interface {
    CreateBackend(config BackendConfig) (Backend, error)
    SupportedTypes() []BackendType
}

type DefaultBackendFactory struct {
    creators map[BackendType]func(BackendConfig) (Backend, error)
}

func NewBackendFactory() *DefaultBackendFactory {
    factory := &DefaultBackendFactory{
        creators: make(map[BackendType]func(BackendConfig) (Backend, error)),
    }
    
    // Register built-in backend creators
    factory.RegisterCreator(HTTPBackend, createGinBackend)
    factory.RegisterCreator(GRPCBackend, createGRPCBackend)
    factory.RegisterCreator(LibP2PBackend, createLibP2PBackend)
    
    return factory
}

func (f *DefaultBackendFactory) RegisterCreator(backendType BackendType, creator func(BackendConfig) (Backend, error)) {
    f.creators[backendType] = creator
}

func (f *DefaultBackendFactory) CreateBackend(config BackendConfig) (Backend, error) {
    creator, exists := f.creators[config.Type]
    if !exists {
        return nil, fmt.Errorf("unsupported backend type: %s", config.Type)
    }
    return creator(config)
}
```

### 1.4 Client Backend Abstraction

**File**: `pkg/client/backend/interfaces.go`

```go
// ClientBackend defines the client-side backend interface
type ClientBackend interface {
    Connect() error
    Disconnect() error
    IsConnected() bool
    
    // Core API methods
    SendRequest(endpoint string, payload interface{}) (*Response, error)
    SendRequestWithContext(ctx context.Context, endpoint string, payload interface{}) (*Response, error)
    
    // Real-time communication
    Subscribe(topic string) (Subscription, error)
    Publish(topic string, message interface{}) error
    
    // Metadata
    Name() string
    Type() BackendType
    Address() string
    
    // Health check
    Ping() error
}

// Response represents a response from the server
type Response struct {
    StatusCode int
    Headers    map[string]string
    Body       []byte
    Error      error
}

// Subscription represents a real-time subscription
type Subscription interface {
    Receive() ([]byte, error)
    Close() error
    Topic() string
    ID() string
}

type ClientConfig struct {
    Type     BackendType            `yaml:"type"`
    Name     string                 `yaml:"name"`
    Address  string                 `yaml:"address"`
    Settings map[string]interface{} `yaml:"settings"`
}

type ClientFactory interface {
    CreateClient(config ClientConfig) (ClientBackend, error)
    SupportedTypes() []BackendType
}
```

### 1.5 HTTP Client Backend Implementation

**File**: `pkg/client/backend/http/http_client.go`

```go
type HTTPClientBackend struct {
    baseURL    string
    httpClient *http.Client
    config     ClientConfig
    connected  bool
    mutex      sync.RWMutex
}

func NewHTTPClientBackend(config ClientConfig) *HTTPClientBackend {
    return &HTTPClientBackend{
        baseURL: config.Address,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        config: config,
    }
}

func (h *HTTPClientBackend) Type() BackendType {
    return HTTPBackend
}

func (h *HTTPClientBackend) Connect() error {
    h.mutex.Lock()
    defer h.mutex.Unlock()
    
    // Test connection with health check
    if err := h.Ping(); err != nil {
        return fmt.Errorf("failed to connect to HTTP backend: %v", err)
    }
    
    h.connected = true
    return nil
}

func (h *HTTPClientBackend) SendRequest(endpoint string, payload interface{}) (*Response, error) {
    return h.SendRequestWithContext(context.Background(), endpoint, payload)
}

func (h *HTTPClientBackend) SendRequestWithContext(ctx context.Context, endpoint string, payload interface{}) (*Response, error) {
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    
    url := h.baseURL + endpoint
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := h.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    return &Response{
        StatusCode: resp.StatusCode,
        Headers:    convertHeaders(resp.Header),
        Body:       body,
    }, nil
}
```

### 1.6 gRPC Client Backend Implementation

**File**: `pkg/client/backend/grpc/grpc_client.go`

```go
type GRPCClientBackend struct {
    conn       *grpc.ClientConn
    client     pb.ColoniesServiceClient
    config     ClientConfig
    connected  bool
    mutex      sync.RWMutex
}

func NewGRPCClientBackend(config ClientConfig) *GRPCClientBackend {
    return &GRPCClientBackend{
        config: config,
    }
}

func (g *GRPCClientBackend) Type() BackendType {
    return GRPCBackend
}

func (g *GRPCClientBackend) Connect() error {
    g.mutex.Lock()
    defer g.mutex.Unlock()
    
    conn, err := grpc.Dial(g.config.Address, grpc.WithInsecure())
    if err != nil {
        return fmt.Errorf("failed to connect to gRPC backend: %v", err)
    }
    
    g.conn = conn
    g.client = pb.NewColoniesServiceClient(conn)
    g.connected = true
    
    return nil
}

func (g *GRPCClientBackend) SendRequestWithContext(ctx context.Context, endpoint string, payload interface{}) (*Response, error) {
    if !g.connected {
        return nil, errors.New("client not connected")
    }
    
    // Convert generic payload to protobuf message based on endpoint
    protoReq, err := g.convertToProtoMessage(endpoint, payload)
    if err != nil {
        return nil, err
    }
    
    // Call appropriate gRPC method based on endpoint
    protoResp, err := g.callGRPCMethod(ctx, endpoint, protoReq)
    if err != nil {
        return nil, err
    }
    
    // Convert protobuf response back to generic response
    return g.convertFromProtoMessage(protoResp)
}
```

### 1.7 LibP2P Client Backend Implementation

**File**: `pkg/client/backend/libp2p/p2p_client.go`

```go
type LibP2PClientBackend struct {
    host      host.Host
    protocols []string
    peers     map[peer.ID]*P2PConnection
    config    ClientConfig
    connected bool
    mutex     sync.RWMutex
}

type P2PConnection struct {
    stream network.Stream
    peerID peer.ID
}

func NewLibP2PClientBackend(config ClientConfig) (*LibP2PClientBackend, error) {
    h, err := libp2p.New()
    if err != nil {
        return nil, err
    }
    
    return &LibP2PClientBackend{
        host:      h,
        protocols: []string{"/colonies/1.0.0"},
        peers:     make(map[peer.ID]*P2PConnection),
        config:    config,
    }, nil
}

func (l *LibP2PClientBackend) Type() BackendType {
    return LibP2PBackend
}

func (l *LibP2PClientBackend) Connect() error {
    l.mutex.Lock()
    defer l.mutex.Unlock()
    
    // Connect to bootstrap peers or discover peers
    if err := l.discoverPeers(); err != nil {
        return fmt.Errorf("failed to discover peers: %v", err)
    }
    
    l.connected = true
    return nil
}

func (l *LibP2PClientBackend) SendRequestWithContext(ctx context.Context, endpoint string, payload interface{}) (*Response, error) {
    if !l.connected {
        return nil, errors.New("client not connected")
    }
    
    // Select best peer for request
    peerID, conn, err := l.selectBestPeer()
    if err != nil {
        return nil, err
    }
    
    // Send request over P2P stream
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    
    message := &P2PMessage{
        Type:     endpoint,
        Payload:  jsonData,
        ID:       generateMessageID(),
    }
    
    return l.sendP2PMessage(ctx, conn, message)
}
```

## Phase 2: Real-Time Communication Abstraction (Week 3-4)

### 2.1 Real-Time Interfaces

**File**: `pkg/service/realtime/interfaces.go`

```go
// Connection represents a bidirectional real-time connection
type Connection interface {
    ID() string
    Send(message []byte) error
    SendJSON(data interface{}) error
    Receive() ([]byte, error)
    Close() error
    IsClosed() bool
    RemoteAddr() string
    GetMetadata() map[string]interface{}
    SetMetadata(key string, value interface{})
    Type() ConnectionType
}

type ConnectionType string

const (
    WebSocketConnection ConnectionType = "websocket"
    LibP2PConnection    ConnectionType = "libp2p"
    GRPCConnection      ConnectionType = "grpc"
)

// ConnectionManager manages real-time connections
type ConnectionManager interface {
    AddConnection(conn Connection) error
    RemoveConnection(connID string) error
    GetConnection(connID string) (Connection, bool)
    GetConnections() []Connection
    GetConnectionsByType(connType ConnectionType) []Connection
    GetConnectionsByFilter(filter ConnectionFilter) []Connection
    Broadcast(message []byte) error
    BroadcastJSON(data interface{}) error
    BroadcastToFilter(message []byte, filter ConnectionFilter) error
    Stats() ConnectionStats
}

// ConnectionFilter defines filtering criteria
type ConnectionFilter func(conn Connection) bool

// ConnectionStats provides connection statistics
type ConnectionStats struct {
    TotalConnections     int
    ConnectionsByType    map[ConnectionType]int
    MessagesSent         int64
    MessagesReceived     int64
    BytesSent           int64
    BytesReceived       int64
}

// RealtimeTransport defines the real-time transport interface
type RealtimeTransport interface {
    Start() error
    Stop() error
    SetupEndpoint(path string, handler ConnectionHandler) error
    GetConnectionManager() ConnectionManager
    Name() string
    Type() ConnectionType
    Config() RealtimeConfig
}

// ConnectionHandler handles new connections
type ConnectionHandler func(conn Connection, manager ConnectionManager)

// MessageHandler handles incoming messages
type MessageHandler func(conn Connection, message []byte) error

// RealtimeConfig defines real-time transport configuration
type RealtimeConfig struct {
    MaxConnections    int           `yaml:"max_connections"`
    ReadTimeout      time.Duration `yaml:"read_timeout"`
    WriteTimeout     time.Duration `yaml:"write_timeout"`
    PingInterval     time.Duration `yaml:"ping_interval"`
    BufferSize       int           `yaml:"buffer_size"`
    EnableMetrics    bool          `yaml:"enable_metrics"`
}
```

### 2.2 WebSocket Implementation

**File**: `pkg/service/realtime/websocket/websocket_transport.go`

```go
type WebSocketTransport struct {
    upgrader    websocket.Upgrader
    manager     ConnectionManager
    endpoints   map[string]ConnectionHandler
    httpServer  *http.Server
    config      RealtimeConfig
    stats       *ConnectionStats
    mutex       sync.RWMutex
}

type WebSocketConnection struct {
    id        string
    conn      *websocket.Conn
    metadata  map[string]interface{}
    transport *WebSocketTransport
    closed    bool
    mutex     sync.RWMutex
    stats     struct {
        messagesSent     int64
        messagesReceived int64
        bytesSent       int64
        bytesReceived   int64
    }
}

func NewWebSocketTransport(config RealtimeConfig) *WebSocketTransport {
    return &WebSocketTransport{
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool { return true },
            BufferSize:  config.BufferSize,
        },
        manager:   NewDefaultConnectionManager(),
        endpoints: make(map[string]ConnectionHandler),
        config:    config,
        stats:     &ConnectionStats{},
    }
}

func (ws *WebSocketConnection) Send(message []byte) error {
    ws.mutex.Lock()
    defer ws.mutex.Unlock()
    
    if ws.closed {
        return errors.New("connection closed")
    }
    
    ws.conn.SetWriteDeadline(time.Now().Add(ws.transport.config.WriteTimeout))
    err := ws.conn.WriteMessage(websocket.TextMessage, message)
    if err == nil {
        atomic.AddInt64(&ws.stats.messagesSent, 1)
        atomic.AddInt64(&ws.stats.bytesSent, int64(len(message)))
    }
    return err
}

func (ws *WebSocketConnection) Receive() ([]byte, error) {
    ws.conn.SetReadDeadline(time.Now().Add(ws.transport.config.ReadTimeout))
    _, message, err := ws.conn.ReadMessage()
    if err == nil {
        atomic.AddInt64(&ws.stats.messagesReceived, 1)
        atomic.AddInt64(&ws.stats.bytesReceived, int64(len(message)))
    }
    return message, err
}
```

### 2.3 LibP2P PubSub Implementation

**File**: `pkg/service/realtime/libp2p/pubsub_transport.go`

```go
type LibP2PPubSubTransport struct {
    host       host.Host
    pubsub     *pubsub.PubSub
    topics     map[string]*pubsub.Topic
    manager    ConnectionManager
    handlers   map[string]ConnectionHandler
    config     RealtimeConfig
    ctx        context.Context
    cancel     context.CancelFunc
}

type LibP2PConnection struct {
    id           string
    peerID       peer.ID
    topic        *pubsub.Topic
    subscription *pubsub.Subscription
    metadata     map[string]interface{}
    transport    *LibP2PPubSubTransport
    closed       bool
    mutex        sync.RWMutex
}

func NewLibP2PPubSubTransport(config RealtimeConfig, libp2pConfig LibP2PConfig) (*LibP2PPubSubTransport, error) {
    ctx, cancel := context.WithCancel(context.Background())
    
    // Create libp2p host
    host, err := libp2p.New(
        libp2p.ListenAddrStrings(libp2pConfig.ListenAddresses...),
        libp2p.Identity(libp2pConfig.PrivateKey),
    )
    if err != nil {
        cancel()
        return nil, err
    }
    
    // Create pubsub instance
    ps, err := pubsub.NewGossipSub(ctx, host)
    if err != nil {
        cancel()
        host.Close()
        return nil, err
    }
    
    return &LibP2PPubSubTransport{
        host:     host,
        pubsub:   ps,
        topics:   make(map[string]*pubsub.Topic),
        manager:  NewDefaultConnectionManager(),
        handlers: make(map[string]ConnectionHandler),
        config:   config,
        ctx:      ctx,
        cancel:   cancel,
    }, nil
}

func (l *LibP2PConnection) Send(message []byte) error {
    l.mutex.RLock()
    defer l.mutex.RUnlock()
    
    if l.closed {
        return errors.New("connection closed")
    }
    
    return l.topic.Publish(l.transport.ctx, message)
}
```

## Phase 3: Handler Migration (Week 5-6)

### 3.1 Update Handler Registry

**File**: `pkg/service/registry/handler_registry.go`

```go
// Update HandlerFunc signature to use abstracted Context
type HandlerFunc func(ctx transport.Context, recoveredID string, payloadType string, jsonString string)
type HandlerFuncWithRawRequest func(ctx transport.Context, recoveredID string, payloadType string, jsonString string, rawRequest string)

type HandlerRegistry struct {
    handlers             map[string]HandlerFunc
    handlersWithRawReq   map[string]HandlerFuncWithRawRequest
    middleware          []transport.MiddlewareFunc
    mutex               sync.RWMutex
}

// Add middleware support
func (r *HandlerRegistry) Use(middleware transport.MiddlewareFunc) {
    r.middleware = append(r.middleware, middleware)
}

// Apply middleware chain
func (r *HandlerRegistry) applyMiddleware(handler transport.HandlerFunc) transport.HandlerFunc {
    for i := len(r.middleware) - 1; i >= 0; i-- {
        handler = r.middleware[i](handler)
    }
    return handler
}

func (r *HandlerRegistry) HandleRequestWithRaw(ctx transport.Context, recoveredID string, payloadType string, jsonString string, rawRequest string) bool {
    // First try handlers that need raw request access
    r.mutex.RLock()
    handlerWithRaw, exists := r.handlersWithRawReq[payloadType]
    r.mutex.RUnlock()
    
    if exists {
        wrappedHandler := r.applyMiddleware(func(ctx transport.Context) {
            handlerWithRaw(ctx, recoveredID, payloadType, jsonString, rawRequest)
        })
        wrappedHandler(ctx)
        return true
    }
    
    // Fall back to regular handlers
    r.mutex.RLock()
    handler, exists := r.handlers[payloadType]
    r.mutex.RUnlock()
    
    if exists {
        wrappedHandler := r.applyMiddleware(func(ctx transport.Context) {
            handler(ctx, recoveredID, payloadType, jsonString)
        })
        wrappedHandler(ctx)
        return true
    }
    
    return false
}
```

### 3.2 Update Server Interface

**File**: `pkg/service/server_interface.go`

```go
type ColoniesServer interface {
    // Replace Gin-specific methods with transport-agnostic ones
    HandleHTTPError(ctx transport.Context, err error, errorCode int) bool
    SendHTTPReply(ctx transport.Context, payloadType string, jsonString string)
    SendEmptyHTTPReply(ctx transport.Context, payloadType string)
    
    // Keep existing database and business logic interfaces
    Validator() security.Validator
    FileDB() database.FileDatabase
    UserDB() database.UserDatabase
    ColonyDB() database.ColonyDatabase
    ExecutorDB() database.ExecutorDatabase
    ProcessDB() database.ProcessDatabase
    AttributeDB() database.AttributeDatabase
    ProcessGraphDB() database.ProcessGraphDatabase
    GeneratorDB() database.GeneratorDatabase
    CronDB() database.CronDatabase
    LogDB() database.LogDatabase
    SnapshotDB() database.SnapshotDatabase
    SecurityDB() database.SecurityDatabase
    ProcessController() process.Controller
    ExclusiveAssign() bool
    TLS() bool
    GetServerID() (string, error)
}
```

### 3.3 Create Server Adapter

**File**: `pkg/service/server_adapter.go`

```go
type ServerAdapter struct {
    server *ColoniesServer
}

func NewServerAdapter(server *ColoniesServer) *ServerAdapter {
    return &ServerAdapter{server: server}
}

// Implement transport-agnostic methods
func (sa *ServerAdapter) HandleHTTPError(ctx transport.Context, err error, errorCode int) bool {
    if err != nil {
        rpcReplyMsg, err := sa.server.generateRPCErrorMsg(err, errorCode)
        if err != nil {
            log.WithFields(log.Fields{"Error": err}).Error("Failed to generate RPC error message")
        }
        
        rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
        if err != nil {
            log.WithFields(log.Fields{"Error": err}).Error("Failed to convert RPC reply to JSON")
        }
        
        ctx.Response().SetStatus(errorCode)
        ctx.Response().WriteString(rpcReplyMsgJSONString)
        return true
    }
    return false
}

func (sa *ServerAdapter) SendHTTPReply(ctx transport.Context, payloadType string, jsonString string) {
    rpcReplyMsg, err := rpc.CreateRPCReplyMsg(payloadType, jsonString)
    if sa.HandleHTTPError(ctx, err, http.StatusBadRequest) {
        return
    }
    
    rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
    if sa.HandleHTTPError(ctx, err, http.StatusBadRequest) {
        return
    }
    
    ctx.Response().SetStatus(http.StatusOK)
    ctx.Response().WriteString(rpcReplyMsgJSONString)
}

func (sa *ServerAdapter) SendEmptyHTTPReply(ctx transport.Context, payloadType string) {
    sa.SendHTTPReply(ctx, payloadType, "{}")
}

// Delegate all other methods to the underlying server
func (sa *ServerAdapter) Validator() security.Validator {
    return sa.server.validator
}

// ... implement all other interface methods as delegates
```

## Phase 4: Multi-Transport Server (Week 7-8)

### 4.1 Multi-Transport Server Architecture

**File**: `pkg/service/multi_server.go`

```go
type MultiTransportServer struct {
    transports          map[string]transport.Transport
    realtimeTransports  map[string]realtime.RealtimeTransport
    handlerRegistry     *registry.HandlerRegistry
    realtimeManager     *realtime.RealtimeManager
    controller          controllers.Controller
    serverAdapter       *ServerAdapter
    
    // Database interfaces
    userDB       database.UserDatabase
    colonyDB     database.ColonyDatabase
    executorDB   database.ExecutorDatabase
    // ... other database interfaces
    
    // Configuration
    config ServerConfig
    
    // Lifecycle
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

type ServerConfig struct {
    Transports         []TransportConfig         `yaml:"transports"`
    RealtimeTransports []RealtimeTransportConfig `yaml:"realtime_transports"`
    Database           DatabaseConfig            `yaml:"database"`
    Security           SecurityConfig            `yaml:"security"`
    Cluster            ClusterConfig             `yaml:"cluster"`
}

type RealtimeTransportConfig struct {
    Name     string                 `yaml:"name"`
    Type     string                 `yaml:"type"`
    Endpoint string                 `yaml:"endpoint"`
    Config   map[string]interface{} `yaml:"config"`
}

func NewMultiTransportServer(config ServerConfig, db database.Database) (*MultiTransportServer, error) {
    ctx, cancel := context.WithCancel(context.Background())
    
    server := &MultiTransportServer{
        transports:         make(map[string]transport.Transport),
        realtimeTransports: make(map[string]realtime.RealtimeTransport),
        handlerRegistry:    registry.NewHandlerRegistry(),
        realtimeManager:    realtime.NewRealtimeManager(),
        config:            config,
        ctx:               ctx,
        cancel:            cancel,
    }
    
    // Initialize database interfaces
    server.userDB = db
    server.colonyDB = db
    // ... set other database interfaces
    
    // Initialize controller
    server.controller = controllers.CreateColoniesController(db, /* other params */)
    
    // Initialize server adapter
    server.serverAdapter = NewServerAdapter(server)
    
    // Initialize transports
    if err := server.initializeTransports(); err != nil {
        return nil, err
    }
    
    // Initialize real-time transports
    if err := server.initializeRealtimeTransports(); err != nil {
        return nil, err
    }
    
    // Register handlers
    server.registerHandlers()
    
    return server, nil
}

func (s *MultiTransportServer) initializeTransports() error {
    factory := transport.NewTransportFactory()
    
    for _, config := range s.config.Transports {
        transport, err := factory.CreateTransport(config)
        if err != nil {
            return fmt.Errorf("failed to create transport %s: %v", config.Name, err)
        }
        
        s.transports[config.Name] = transport
        log.WithFields(log.Fields{
            "transport": config.Name,
            "type":      config.Type,
            "address":   config.Address,
        }).Info("Initialized transport")
    }
    
    return nil
}

func (s *MultiTransportServer) setupRoutes() {
    // Register the same routes across all transports
    for name, transport := range s.transports {
        log.WithFields(log.Fields{"transport": name}).Info("Setting up routes")
        
        // Main API endpoint
        transport.Handle("POST", "/api", func(ctx transport.Context) {
            s.handleAPIRequest(ctx)
        })
        
        // Health check endpoint
        transport.Handle("GET", "/health", func(ctx transport.Context) {
            s.handleHealthRequest(ctx)
        })
        
        // Add middleware
        transport.Use(s.loggingMiddleware)
        transport.Use(s.corsMiddleware)
        transport.Use(s.authMiddleware)
    }
}

func (s *MultiTransportServer) Start() error {
    s.setupRoutes()
    
    // Start all transports
    for name, transport := range s.transports {
        s.wg.Add(1)
        go func(transportName string, t transport.Transport) {
            defer s.wg.Done()
            
            log.WithFields(log.Fields{"transport": transportName}).Info("Starting transport")
            if err := t.Start(s.getTransportAddress(transportName)); err != nil {
                log.WithFields(log.Fields{
                    "transport": transportName,
                    "error":     err,
                }).Error("Transport failed to start")
            }
        }(name, transport)
    }
    
    // Start all real-time transports
    for name, rtTransport := range s.realtimeTransports {
        s.wg.Add(1)
        go func(transportName string, rt realtime.RealtimeTransport) {
            defer s.wg.Done()
            
            log.WithFields(log.Fields{"realtime_transport": transportName}).Info("Starting real-time transport")
            if err := rt.Start(); err != nil {
                log.WithFields(log.Fields{
                    "realtime_transport": transportName,
                    "error":             err,
                }).Error("Real-time transport failed to start")
            }
        }(name, rtTransport)
    }
    
    return nil
}

func (s *MultiTransportServer) Stop() error {
    log.Info("Stopping multi-transport server")
    
    s.cancel()
    
    // Stop all transports
    for name, transport := range s.transports {
        log.WithFields(log.Fields{"transport": name}).Info("Stopping transport")
        if err := transport.Stop(); err != nil {
            log.WithFields(log.Fields{
                "transport": name,
                "error":     err,
            }).Error("Failed to stop transport")
        }
    }
    
    // Stop all real-time transports
    for name, rtTransport := range s.realtimeTransports {
        log.WithFields(log.Fields{"realtime_transport": name}).Info("Stopping real-time transport")
        if err := rtTransport.Stop(); err != nil {
            log.WithFields(log.Fields{
                "realtime_transport": name,
                "error":             err,
            }).Error("Failed to stop real-time transport")
        }
    }
    
    // Stop controller
    s.controller.Stop()
    
    s.wg.Wait()
    return nil
}
```

### 4.2 Middleware Implementation

**File**: `pkg/service/middleware/middleware.go`

```go
func (s *MultiTransportServer) loggingMiddleware(next transport.HandlerFunc) transport.HandlerFunc {
    return func(ctx transport.Context) {
        start := time.Now()
        
        log.WithFields(log.Fields{
            "method": ctx.Request().GetMethod(),
            "path":   ctx.Request().GetPath(),
            "remote": ctx.Request().GetRemoteAddr(),
        }).Debug("Request started")
        
        next(ctx)
        
        log.WithFields(log.Fields{
            "method":   ctx.Request().GetMethod(),
            "path":     ctx.Request().GetPath(),
            "status":   ctx.Response().GetStatus(),
            "duration": time.Since(start),
        }).Debug("Request completed")
    }
}

func (s *MultiTransportServer) corsMiddleware(next transport.HandlerFunc) transport.HandlerFunc {
    return func(ctx transport.Context) {
        ctx.Response().SetHeader("Access-Control-Allow-Origin", "*")
        ctx.Response().SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        ctx.Response().SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if ctx.Request().GetMethod() == "OPTIONS" {
            ctx.Response().SetStatus(http.StatusOK)
            return
        }
        
        next(ctx)
    }
}

func (s *MultiTransportServer) authMiddleware(next transport.HandlerFunc) transport.HandlerFunc {
    return func(ctx transport.Context) {
        // Skip auth for health checks
        if ctx.Request().GetPath() == "/health" {
            next(ctx)
            return
        }
        
        // Add authentication logic here
        next(ctx)
    }
}
```

## Phase 5: Configuration and Deployment (Week 9-10)

### 5.1 Configuration Management

**File**: `config/server.yaml`

```yaml
# Multi-backend server configuration
backends:
  - name: http
    type: http
    address: ":8080"
    settings:
      framework: gin  # or echo, fiber, etc.
      cors:
        allow_origins: ["*"]
        allow_methods: ["GET", "POST", "PUT", "DELETE"]
        allow_headers: ["Content-Type", "Authorization"]
      tls:
        enabled: false
        cert_file: ""
        key_file: ""
      
  - name: grpc
    type: grpc
    address: ":8081"
    settings:
      max_recv_msg_size: 4194304  # 4MB
      max_send_msg_size: 4194304  # 4MB
      keepalive:
        time: 30s
        timeout: 5s

  - name: libp2p
    type: libp2p
    address: "/ip4/0.0.0.0/tcp/9000"
    settings:
      protocols: ["/colonies/1.0.0"]
      bootstrap_peers: []
      dht_enabled: true

realtime_transports:
  - name: websocket
    type: websocket
    endpoint: "/pubsub"
    config:
      max_connections: 1000
      read_timeout: 60s
      write_timeout: 10s
      ping_interval: 30s
      buffer_size: 1024
      enable_metrics: true
      
  - name: libp2p
    type: libp2p_pubsub
    endpoint: "colonies-events"
    config:
      listen_addresses:
        - "/ip4/0.0.0.0/tcp/9001"
        - "/ip6/::/tcp/9001"
      bootstrap_peers:
        - "/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"
      topics:
        - "colonies-events"
        - "colonies-control"
      max_connections: 100
      enable_metrics: true

database:
  type: postgresql
  connection_string: "postgres://user:pass@localhost/colonies?sslmode=disable"
  max_connections: 25
  max_idle_connections: 5
  connection_max_lifetime: 300s

security:
  enable_tls: false
  cert_file: ""
  key_file: ""
  require_client_certs: false

cluster:
  enabled: true
  node_name: "node1"
  etcd:
    client_port: 24100
    peer_port: 23100
    data_path: "/tmp/colonies/etcd"

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  port: 8090
  path: "/metrics"
```

### 5.2 Docker Support

**File**: `docker/Dockerfile.multi`

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# Build the multi-transport server
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o colonies-server-multi ./cmd/multi-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/colonies-server-multi .
COPY --from=builder /app/config/server.yaml ./config/

EXPOSE 8080 8081 9001

CMD ["./colonies-server-multi", "--config", "config/server.yaml"]
```

**File**: `docker-compose.multi.yml`

```yaml
version: '3.8'

services:
  colonies-multi:
    build:
      context: .
      dockerfile: docker/Dockerfile.multi
    ports:
      - "8080:8080"  # HTTP
      - "8081:8081"  # gRPC
      - "9001:9001"  # LibP2P
      - "8090:8090"  # Metrics
    environment:
      - LOG_LEVEL=debug
      - DB_CONNECTION=postgres://colonies:password@db:5432/colonies?sslmode=disable
    volumes:
      - ./config:/root/config
      - colonies_data:/tmp/colonies
    depends_on:
      - db
      - etcd

  db:
    image: timescale/timescaledb:latest-pg14
    environment:
      - POSTGRES_DB=colonies
      - POSTGRES_USER=colonies
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  etcd:
    image: quay.io/coreos/etcd:v3.5.0
    command:
      - /usr/local/bin/etcd
      - --data-dir=/etcd-data
      - --listen-client-urls=http://0.0.0.0:2379
      - --advertise-client-urls=http://etcd:2379
      - --listen-peer-urls=http://0.0.0.0:2380
      - --initial-advertise-peer-urls=http://etcd:2380
      - --initial-cluster=default=http://etcd:2380
      - --name=default
    volumes:
      - etcd_data:/etcd-data
    ports:
      - "2379:2379"
      - "2380:2380"

volumes:
  postgres_data:
  etcd_data:
  colonies_data:
```

## Phase 6: Testing and Validation (Week 11-12)

### 6.1 Integration Tests

**File**: `pkg/service/integration_test.go`

```go
func TestMultiTransportServer(t *testing.T) {
    // Test configuration
    config := ServerConfig{
        Transports: []TransportConfig{
            {
                Name:    "http",
                Type:    transport.GinTransport,
                Address: ":0", // Random port
            },
        },
        RealtimeTransports: []RealtimeTransportConfig{
            {
                Name:     "websocket",
                Type:     "websocket",
                Endpoint: "/pubsub",
            },
        },
    }
    
    // Create and start server
    server, err := NewMultiTransportServer(config, mockDB)
    require.NoError(t, err)
    
    go server.Start()
    defer server.Stop()
    
    // Test HTTP transport
    t.Run("HTTP Transport", func(t *testing.T) {
        // Test API endpoints
        testHTTPAPI(t, server)
    })
    
    // Test WebSocket transport
    t.Run("WebSocket Transport", func(t *testing.T) {
        // Test real-time communication
        testWebSocketAPI(t, server)
    })
}

func TestLibP2PTransport(t *testing.T) {
    // Test P2P communication
    config := ServerConfig{
        RealtimeTransports: []RealtimeTransportConfig{
            {
                Name: "libp2p",
                Type: "libp2p_pubsub",
                Config: map[string]interface{}{
                    "listen_addresses": []string{"/ip4/127.0.0.1/tcp/0"},
                    "topics":          []string{"test-topic"},
                },
            },
        },
    }
    
    server, err := NewMultiTransportServer(config, mockDB)
    require.NoError(t, err)
    
    // Test P2P message publishing and subscription
    testP2PMessaging(t, server)
}
```

### 6.2 Performance Benchmarks

**File**: `pkg/service/benchmark_test.go`

```go
func BenchmarkMultiTransportThroughput(b *testing.B) {
    server := createTestServer(b)
    defer server.Stop()
    
    b.Run("HTTP", func(b *testing.B) {
        benchmarkHTTPThroughput(b, server)
    })
    
    b.Run("gRPC", func(b *testing.B) {
        benchmarkGRPCThroughput(b, server)
    })
    
    b.Run("WebSocket", func(b *testing.B) {
        benchmarkWebSocketThroughput(b, server)
    })
    
    b.Run("LibP2P", func(b *testing.B) {
        benchmarkLibP2PThroughput(b, server)
    })
}

func BenchmarkConcurrentConnections(b *testing.B) {
    server := createTestServer(b)
    defer server.Stop()
    
    connections := []int{10, 100, 1000, 10000}
    
    for _, connCount := range connections {
        b.Run(fmt.Sprintf("Connections_%d", connCount), func(b *testing.B) {
            benchmarkConcurrentConnections(b, server, connCount)
        })
    }
}
```

## Implementation Timeline

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| Phase 1 | Week 1-2 | Core transport abstraction, Gin adapter |
| Phase 2 | Week 3-4 | Real-time abstraction, WebSocket + LibP2P |
| Phase 3 | Week 5-6 | Handler migration, registry updates |
| Phase 4 | Week 7-8 | Multi-transport server architecture |
| Phase 5 | Week 9-10 | Configuration, deployment, Docker |
| Phase 6 | Week 11-12 | Testing, validation, documentation |

## Migration Strategy

### Backward Compatibility

1. **Phase 1-2**: No breaking changes, abstraction runs alongside existing code
2. **Phase 3**: Gradual handler migration with feature flags
3. **Phase 4**: Optional multi-transport mode
4. **Phase 5**: Full migration with backward compatibility

### Feature Flags

```go
type FeatureFlags struct {
    EnableMultiTransport  bool `yaml:"enable_multi_transport"`
    EnableLibP2P         bool `yaml:"enable_libp2p"`
    EnableGRPC           bool `yaml:"enable_grpc"`
    MigratedHandlers     []string `yaml:"migrated_handlers"`
}
```

### Rollback Plan

1. Feature flags allow disabling new transports
2. Legacy Gin-based handlers remain available
3. Configuration-driven transport selection
4. Gradual handler migration with fallback

## Success Metrics

1. **Performance**: No more than 5% overhead from abstraction
2. **Compatibility**: 100% backward compatibility during migration
3. **Reliability**: 99.9% uptime across all transports
4. **Scalability**: Support 10x more concurrent connections with P2P
5. **Maintainability**: 50% reduction in transport-specific code

## Risk Mitigation

1. **Performance Impact**: Benchmark at each phase
2. **Complexity**: Start with simple abstractions, evolve gradually
3. **Testing**: Comprehensive integration tests for all transports
4. **Dependencies**: Use stable, well-maintained transport libraries
5. **Migration Risk**: Feature flags and rollback mechanisms

## Future Extensions

1. **Additional Transports**: MQTT, NATS, Apache Kafka
2. **Protocol Translation**: HTTP to gRPC automatic conversion
3. **Load Balancing**: Transport-aware request routing
4. **Service Mesh**: Integration with Istio/Linkerd
5. **Monitoring**: Transport-specific metrics and alerting

---

This plan provides a comprehensive roadmap for implementing a flexible, multi-transport server architecture while maintaining backward compatibility and ensuring high performance.