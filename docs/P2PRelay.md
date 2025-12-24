# LibP2P Relay Configuration

## Environment Variables

The `colonies p2p relay` command uses dedicated environment variables with the `COLONIES_P2P_RELAY_` prefix:

### Required Variables

- **`COLONIES_P2P_RELAY_PUBLIC_IP`** - Public IP address to announce to the DHT
  - Required for remote access
  - Example: `COLONIES_P2P_RELAY_PUBLIC_IP=46.62.173.145`

### Optional Variables

- **`COLONIES_P2P_RELAY_PORT`** - Port to listen on for both TCP and UDP/QUIC
  - Default: `4002`
  - Used for both TCP and QUIC transports
  - Example: `COLONIES_P2P_RELAY_PORT=4002`

- **`COLONIES_P2P_RELAY_IDENTITY`** - Hex-encoded LibP2P identity
  - Use this to provide the identity as hex-encoded environment variable
  - Takes precedence over base64 identity
  - Example: `COLONIES_P2P_RELAY_IDENTITY="080112..."`

- **`COLONIES_P2P_RELAY_IDENTITY_FILE`** - Base64-encoded LibP2P identity content
  - Alternative to hex-encoded identity
  - **Note:** Despite the name, this accepts the base64 content directly, NOT a file path
  - Example: `COLONIES_P2P_RELAY_IDENTITY_FILE="CAESQNhOpx..."`

## Quick Start

### 1. Generate a new identity

```bash
colonies p2p generate
```

This outputs:
- Peer ID
- Hex-encoded identity (for `COLONIES_P2P_RELAY_IDENTITY`)
- Base64-encoded identity (for saving to file)

### 2. Start the relay (Option A: Using hex-encoded identity)

```bash
export COLONIES_P2P_RELAY_PUBLIC_IP=46.62.173.145
export COLONIES_P2P_RELAY_IDENTITY="080112..."
colonies p2p relay
```

### 3. Start the relay (Option B: Using base64-encoded identity)

```bash
export COLONIES_P2P_RELAY_PUBLIC_IP=46.62.173.145
export COLONIES_P2P_RELAY_IDENTITY_FILE="CAESQNhOpx..."
colonies p2p relay
```

### 4. Configure clients to use your relay

The relay will output multiaddresses when it starts. Copy these to your `docker-compose.env`:

```bash
# Server configuration (TCP)
export COLONIES_SERVER_LIBP2P_BOOTSTRAP_PEERS="/ip4/46.62.173.145/tcp/4002/p2p/12D3Koo..."

# Client configuration (QUIC - recommended for 5G/mobile)
export COLONIES_CLIENT_LIBP2P_BOOTSTRAP_PEERS="/ip4/46.62.173.145/udp/4002/quic-v1/p2p/12D3Koo..."
```

## Deployment Example (Hetzner VM)

```bash
# 1. Generate identity once
colonies p2p generate

# 2. Copy the hex identity and save to environment file
cat > /etc/colonies/relay.env << 'EOF'
export COLONIES_P2P_RELAY_PUBLIC_IP=46.62.173.145
export COLONIES_P2P_RELAY_IDENTITY="080112..."  # Paste hex from generate output
EOF

# 3. Start relay
source /etc/colonies/relay.env
colonies p2p relay
```

## Systemd Service Example

Create `/etc/systemd/system/colonies-relay.service`:

```ini
[Unit]
Description=Colonies LibP2P Relay/Bootstrap Node
After=network.target

[Service]
Type=simple
User=colonies
Environment="COLONIES_P2P_RELAY_PUBLIC_IP=46.62.173.145"
Environment="COLONIES_P2P_RELAY_PORT=4002"
Environment="COLONIES_P2P_RELAY_IDENTITY=080112..."
ExecStart=/usr/local/bin/colonies p2p relay
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Then:
```bash
systemctl daemon-reload
systemctl enable colonies-relay
systemctl start colonies-relay
systemctl status colonies-relay
```

## Differences from Colonies Server Variables

| Purpose | Relay Variable | Server Variable |
|---------|---------------|-----------------|
| Identity (hex) | `COLONIES_P2P_RELAY_IDENTITY` | `COLONIES_SERVER_LIBP2P_IDENTITY` |
| Identity (base64) | `COLONIES_P2P_RELAY_IDENTITY_FILE` | (not used) |
| Public IP | `COLONIES_P2P_RELAY_PUBLIC_IP` | (uses ANNOUNCE_ADDRS) |
| Port | `COLONIES_P2P_RELAY_PORT` | `COLONIES_SERVER_LIBP2P_PORT` |

**Note:** `COLONIES_P2P_RELAY_IDENTITY_FILE` accepts base64-encoded identity content directly as an environment variable, NOT a file path.

This separation ensures relay nodes have their own independent configuration and identity, separate from the Colonies server.
