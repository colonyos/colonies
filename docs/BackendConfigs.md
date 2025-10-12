# Backend Configuration Examples

# ============================================================================
# DEPLOYMENT SCENARIOS
# ============================================================================
#
# Scenario 1: Local Development (Fast, Simple)
# -----------------------------------------------
# Client:
# export COLONIES_CLIENT_BACKENDS="http"
# export COLONIES_CLIENT_HTTP_HOST="localhost"
# export COLONIES_CLIENT_HTTP_PORT="50080"
# export COLONIES_CLIENT_HTTP_INSECURE="true"
# Server:
# export COLONIES_SERVER_BACKENDS="http"
# export COLONIES_SERVER_HTTP_HOST="0.0.0.0"
# export COLONIES_SERVER_HTTP_PORT="50080"
#
# Scenario 2: High-Performance gRPC Deployment
# ----------------------------------------------
# Server:
# export COLONIES_SERVER_BACKENDS="grpc"
# export COLONIES_SERVER_GRPC_PORT="50051"
# export COLONIES_SERVER_GRPC_INSECURE="true"  # For development
# Client:
# export COLONIES_CLIENT_BACKENDS="grpc"
# export COLONIES_CLIENT_GRPC_HOST="localhost"
# export COLONIES_CLIENT_GRPC_PORT="50051"
# export COLONIES_CLIENT_GRPC_INSECURE="true"
# Benefits:
# - Higher performance than HTTP (binary protocol, multiplexing)
# - Lower latency for RPC calls
# - Efficient serialization with Protocol Buffers
# - Ideal for microservices and high-throughput systems
# Note: Each backend has its own port variable to avoid conflicts:
#   - HTTP: COLONIES_SERVER_HTTP_PORT (50080) / COLONIES_CLIENT_HTTP_PORT (50080)
#   - gRPC: COLONIES_SERVER_GRPC_PORT (50051) / COLONIES_CLIENT_GRPC_PORT (50051)
#   - LibP2P: COLONIES_SERVER_LIBP2P_PORT (5000)
#   - CoAP: COLONIES_SERVER_COAP_PORT (5683) / COLONIES_CLIENT_COAP_PORT (5683)
#
# Scenario 3: IoT/Constrained Devices with CoAP (Lightweight)
# ------------------------------------------------------------
# Server:
# export COLONIES_SERVER_BACKENDS="coap"
# export COLONIES_SERVER_COAP_PORT="5683"  # Default CoAP port
# Client:
# export COLONIES_CLIENT_BACKENDS="coap"
# export COLONIES_CLIENT_COAP_HOST="localhost"
# export COLONIES_CLIENT_COAP_PORT="5683"
# Benefits:
# - Extremely lightweight protocol designed for constrained devices
# - UDP-based (lower overhead than TCP)
# - Perfect for IoT sensors, embedded systems, edge devices
# - Minimal memory and power consumption
# - RFC 7252 standard protocol
#
# Scenario 4: Edge Device with Intermittent Connectivity (Resilient)
# -------------------------------------------------------------------
# Client:
# export COLONIES_CLIENT_BACKENDS="libp2p,http"  # Try P2P first, fallback to HTTP
# export COLONIES_CLIENT_LIBP2P_HOST="dht"
# export COLONIES_CLIENT_LIBP2P_BOOTSTRAP_PEERS="/dns/localhost/tcp/5000/p2p/..."
# export COLONIES_CLIENT_HTTP_HOST="localhost"
# export COLONIES_CLIENT_HTTP_PORT="50080"
# Server:
# export COLONIES_SERVER_BACKENDS="libp2p,http"
#
# Scenario 5: Behind Firewall/NAT (P2P Only, DHT Discovery)
# ----------------------------------------------------------
# Client:
# export COLONIES_CLIENT_BACKENDS="libp2p"
# export COLONIES_CLIENT_LIBP2P_HOST="dht"
# export COLONIES_CLIENT_LIBP2P_BOOTSTRAP_PEERS="<public-bootstrap-nodes>"
# Server:
# export COLONIES_SERVER_BACKENDS="libp2p"
#
# Scenario 6: Autonomous System with Network Partitions (Maximum Resilience)
# ---------------------------------------------------------------------------
# Client:
# export COLONIES_CLIENT_BACKENDS="libp2p,http"  # Multiple transports
# export COLONIES_CLIENT_LIBP2P_HOST="dht"
# export COLONIES_CLIENT_LIBP2P_BOOTSTRAP_PEERS="<hybrid-mode>"  # DHT + direct peers
# export COLONIES_CLIENT_HTTP_HOST="localhost"
# export COLONIES_CLIENT_HTTP_PORT="50080"
# Server:
# export COLONIES_SERVER_BACKENDS="libp2p,http"
# This enables:
# - Work during network partitions (local autonomy)
# - Automatic reconnection when network restored
# - Multiple paths to reach server (P2P + HTTP)
# - Perfect for drones, robots, edge AI, distributed sensors
#
# Scenario 7: Universal Gateway (All Protocols)
# ----------------------------------------------
# Server runs all backends simultaneously for maximum compatibility:
# export COLONIES_SERVER_BACKENDS="http,grpc,libp2p,coap"
# export COLONIES_SERVER_HTTP_PORT="50080"
# export COLONIES_SERVER_GRPC_PORT="50051"
# export COLONIES_SERVER_LIBP2P_PORT="5000"
# export COLONIES_SERVER_COAP_PORT="5683"
# This enables:
# - Web clients use HTTP
# - High-performance services use gRPC
# - P2P clients use LibP2P
# - IoT devices use CoAP
# - All connected to same backend/database
#
# ============================================================================

ColonyOS supports four independent backends: HTTP (gin), gRPC, LibP2P, and CoAP. Each backend has its own separate configuration with no shared settings.

## HTTP Backend (Gin)

**Server Configuration:**
```bash
export COLONIES_SERVER_BACKENDS="http"
export COLONIES_SERVER_HTTP_HOST="0.0.0.0"      # Bind address
export COLONIES_SERVER_HTTP_PORT="50080"         # HTTP port
export COLONIES_SERVER_HTTP_TLS="false"          # TLS enabled/disabled
```

**Client Configuration:**
```bash
export COLONIES_CLIENT_BACKENDS="http"
export COLONIES_CLIENT_HTTP_HOST="localhost"     # Server hostname
export COLONIES_CLIENT_HTTP_PORT="50080"         # HTTP port (must match server)
export COLONIES_CLIENT_HTTP_INSECURE="true"      # Run without TLS
# export COLONIES_CLIENT_HTTP_SKIP_TLS_VERIFY="false"  # Skip TLS verification
```

## gRPC Backend

**Server Configuration:**
```bash
export COLONIES_SERVER_BACKENDS="grpc"
export COLONIES_SERVER_GRPC_PORT="50051"         # gRPC port (REQUIRED)
export COLONIES_SERVER_GRPC_INSECURE="true"      # Run without TLS
# For production with TLS:
# export COLONIES_SERVER_GRPC_INSECURE="false"
# export COLONIES_SERVER_GRPC_TLS_CERT="/path/to/cert.pem"
# export COLONIES_SERVER_GRPC_TLS_KEY="/path/to/key.pem"
```

**Client Configuration:**
```bash
export COLONIES_CLIENT_BACKENDS="grpc"
export COLONIES_CLIENT_GRPC_HOST="localhost"     # Server hostname
export COLONIES_CLIENT_GRPC_PORT="50051"         # gRPC port (must match server)
export COLONIES_CLIENT_GRPC_INSECURE="true"      # Run without TLS
# export COLONIES_CLIENT_GRPC_SKIP_TLS_VERIFY="false"  # Skip TLS verification
```

## LibP2P Backend

**Server Configuration:**
```bash
export COLONIES_SERVER_BACKENDS="libp2p"
export COLONIES_SERVER_HTTP_PORT="50080"         # HTTP fallback port
export COLONIES_SERVER_LIBP2P_PORT="5000"        # LibP2P TCP port
export COLONIES_SERVER_LIBP2P_IDENTITY="..."     # Optional: predefined identity
export COLONIES_SERVER_LIBP2P_BOOTSTRAP_PEERS="..." # Bootstrap peers
```

**Client Configuration:**
```bash
export COLONIES_CLIENT_BACKENDS="libp2p"
export COLONIES_CLIENT_LIBP2P_HOST="dht"         # DHT discovery
export COLONIES_CLIENT_LIBP2P_BOOTSTRAP_PEERS="..." # Bootstrap peers
# Or direct connection with multiaddress:
# export COLONIES_CLIENT_LIBP2P_HOST="/dns/localhost/tcp/5000/p2p/12D3KooW..."
```

## CoAP Backend

**Server Configuration:**
```bash
export COLONIES_SERVER_BACKENDS="coap"
export COLONIES_SERVER_COAP_PORT="5683"          # CoAP port (UDP, default CoAP port)
```

**Client Configuration:**
```bash
export COLONIES_CLIENT_BACKENDS="coap"
export COLONIES_CLIENT_COAP_HOST="localhost"     # Server hostname
export COLONIES_CLIENT_COAP_PORT="5683"          # CoAP port (must match server)
# Optional TLS/DTLS settings:
# export COLONIES_CLIENT_COAP_INSECURE="true"
# export COLONIES_CLIENT_COAP_SKIP_TLS_VERIFY="false"
```

**Benefits:**
- Extremely lightweight protocol (RFC 7252)
- UDP-based transport (lower overhead than TCP)
- Designed for constrained devices and IoT
- Minimal memory footprint and power consumption
- Perfect for sensors, embedded systems, battery-powered devices

## Running Multiple Backends Simultaneously

You can run multiple backends on the same server by setting multiple backend types:

**Server (Two Backends):**
```bash
export COLONIES_SERVER_BACKENDS="http,grpc"      # Run both HTTP and gRPC
export COLONIES_SERVER_HTTP_PORT="50080"
export COLONIES_SERVER_HTTP_TLS="false"
export COLONIES_SERVER_GRPC_PORT="50051"
export COLONIES_SERVER_GRPC_INSECURE="true"
```

**Server (All Four Backends):**
```bash
export COLONIES_SERVER_BACKENDS="http,grpc,libp2p,coap"  # Universal gateway
export COLONIES_SERVER_HTTP_PORT="50080"
export COLONIES_SERVER_GRPC_PORT="50051"
export COLONIES_SERVER_LIBP2P_PORT="5000"
export COLONIES_SERVER_COAP_PORT="5683"
```

**Client (with fallback):**
```bash
export COLONIES_CLIENT_BACKENDS="grpc,http"      # Try gRPC first, fallback to HTTP
export COLONIES_CLIENT_GRPC_HOST="localhost"
export COLONIES_CLIENT_GRPC_PORT="50051"
export COLONIES_CLIENT_GRPC_INSECURE="true"
export COLONIES_CLIENT_HTTP_HOST="localhost"
export COLONIES_CLIENT_HTTP_PORT="50080"
export COLONIES_CLIENT_HTTP_INSECURE="true"
```

**Client (Multi-protocol fallback chain):**
```bash
export COLONIES_CLIENT_BACKENDS="coap,grpc,http"  # Try CoAP first, then gRPC, then HTTP
export COLONIES_CLIENT_COAP_HOST="localhost"
export COLONIES_CLIENT_COAP_PORT="5683"
export COLONIES_CLIENT_GRPC_HOST="localhost"
export COLONIES_CLIENT_GRPC_PORT="50051"
export COLONIES_CLIENT_GRPC_INSECURE="true"
export COLONIES_CLIENT_HTTP_HOST="localhost"
export COLONIES_CLIENT_HTTP_PORT="50080"
export COLONIES_CLIENT_HTTP_INSECURE="true"
```

## Port Summary

| Backend | Protocol | Server Variable | Client Variable | Default Port |
|---------|----------|----------------|-----------------|--------------|
| HTTP    | TCP      | COLONIES_SERVER_HTTP_PORT | COLONIES_CLIENT_HTTP_PORT | 50080 |
| gRPC    | TCP      | COLONIES_SERVER_GRPC_PORT | COLONIES_CLIENT_GRPC_PORT | 50051 (required) |
| LibP2P  | TCP      | COLONIES_SERVER_LIBP2P_PORT | N/A (uses multiaddr) | 5000 |
| CoAP    | UDP      | COLONIES_SERVER_COAP_PORT | COLONIES_CLIENT_COAP_PORT | 5683 |

**Protocol Characteristics:**

| Backend | Transport | Use Case | Overhead | Best For |
|---------|-----------|----------|----------|----------|
| HTTP    | TCP       | Web APIs, REST clients | Medium | General purpose, web browsers |
| gRPC    | TCP       | High-performance RPC | Low | Microservices, high-throughput |
| LibP2P  | TCP/QUIC  | P2P networks, NAT traversal | Medium | Edge computing, autonomous systems |
| CoAP    | UDP       | IoT, constrained devices | Very Low | Sensors, embedded systems, battery-powered |

**Important Notes:**
- Each backend has completely separate client and server configuration
- Client uses `COLONIES_CLIENT_*` variables, server uses `COLONIES_SERVER_*`
- No configuration is shared between backends
- Backends can run independently or simultaneously
- CoAP uses UDP (note the `/udp` in docker-compose port mapping)
- Backward compatibility: client variables fall back to old `COLONIES_SERVER_*` names if client-specific variables are not set

## Protocol Selection Guide

**Choose HTTP when:**
- Building web applications or REST APIs
- Need compatibility with existing HTTP tooling
- Using standard web browsers or curl
- TLS/SSL security is important

**Choose gRPC when:**
- Need high performance and low latency
- Building microservices architecture
- Want efficient binary serialization
- Using Protocol Buffers for data exchange

**Choose LibP2P when:**
- Operating behind firewalls or NAT
- Need P2P connectivity without central server
- Building autonomous or distributed systems
- Want DHT-based service discovery
- Network partitions are expected

**Choose CoAP when:**
- Working with IoT devices or sensors
- Device has limited memory or CPU
- Battery power is a constraint
- Need minimal network overhead (UDP)
- Following RFC 7252 IoT standards
