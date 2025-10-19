#!/bin/bash
set -e

echo "═══════════════════════════════════════════════════════════════"
echo "Colonies Docker Compose Startup Script"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Detect various IP addresses
echo "Detecting network configuration..."
echo ""

# Method 1: Local network IP (most reliable for LAN access)
LOCAL_IP=$(hostname -I | awk '{print $1}')
echo "  Local IP (hostname -I):     $LOCAL_IP"

# Method 2: IP used for external routing
ROUTE_IP=$(ip route get 1.1.1.1 2>/dev/null | grep -oP 'src \K\S+' || echo "$LOCAL_IP")
echo "  Route IP (primary interface): $ROUTE_IP"

# Method 3: External/Public IP (if accessible)
echo -n "  External IP (public):        "
EXTERNAL_IP=$(timeout 3 curl -s ifconfig.me 2>/dev/null || timeout 3 curl -s icanhazip.com 2>/dev/null || echo "")
if [ -z "$EXTERNAL_IP" ]; then
    echo "(not detected - no internet or behind NAT)"
else
    echo "$EXTERNAL_IP"
fi

# Method 4: Docker gateway (for docker-to-host communication)
DOCKER_GATEWAY=$(ip route | grep docker0 2>/dev/null | awk '{print $9}' || echo "172.17.0.1")
echo "  Docker Gateway:              $DOCKER_GATEWAY"

echo ""
echo "─────────────────────────────────────────────────────────────"
echo "📋 Configuration Selection"
echo "─────────────────────────────────────────────────────────────"

# Determine which IP to use for announce addresses
# Priority: ROUTE_IP (best for most scenarios)
HOST_IP="${ROUTE_IP}"

echo ""
echo "  Using IP for LibP2P announce: $HOST_IP"
echo ""
echo "  This IP will be announced to the DHT so peers can find and"
echo "  connect to the Colonies server."
echo ""

# Export as environment variables
export HOST_IP="$HOST_IP"
export LOCAL_IP="$LOCAL_IP"
export ROUTE_IP="$ROUTE_IP"
export EXTERNAL_IP="${EXTERNAL_IP:-$HOST_IP}"
export DOCKER_GATEWAY="$DOCKER_GATEWAY"

# Set the announce addresses for LibP2P
export COLONIES_SERVER_LIBP2P_ANNOUNCE_ADDRS="/ip4/${HOST_IP}/tcp/50000,/ip4/${HOST_IP}/udp/50001/quic-v1"

echo "─────────────────────────────────────────────────────────────"
echo "🔧 LibP2P Configuration"
echo "─────────────────────────────────────────────────────────────"
echo ""
echo "  COLONIES_SERVER_LIBP2P_ANNOUNCE_ADDRS:"
echo "    ${COLONIES_SERVER_LIBP2P_ANNOUNCE_ADDRS}"
echo ""

# Load the docker-compose.env file (but HOST_IP will override)
if [ -f docker-compose.env ]; then
    echo " Loading docker-compose.env..."
    # Source it but allow our HOST_IP to take precedence
    set -a
    source docker-compose.env
    set +a

    # Re-export our detected values to override
    export HOST_IP="$HOST_IP"
    export COLONIES_SERVER_LIBP2P_ANNOUNCE_ADDRS="/ip4/${HOST_IP}/tcp/50000,/ip4/${HOST_IP}/udp/50001/quic-v1"

    echo "   Loaded (with HOST_IP override: $HOST_IP)"
else
    echo "   Warning: docker-compose.env not found"
fi

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo " Starting Docker Compose"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Check if we should run in detached mode
DETACHED=""
if [ "$1" = "-d" ] || [ "$1" = "--detach" ]; then
    DETACHED="-d"
    echo "Running in detached mode..."
else
    echo "Running in foreground mode (use -d for detached)"
    echo "Press Ctrl+C to stop..."
fi

echo ""

# Start docker-compose with the env file and our overrides
docker-compose --env-file docker-compose.env up $DETACHED

# If running in detached mode, show helpful info
if [ -n "$DETACHED" ]; then
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "✓ Colonies started in background"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    echo "Useful commands:"
    echo "  docker-compose logs -f                  # View all logs"
    echo "  docker-compose logs -f colonies-server  # View server logs"
    echo "  docker-compose ps                       # Check status"
    echo "  docker-compose down                     # Stop all services"
    echo ""
    echo "Server should be accessible at:"
    echo "  LibP2P: /ip4/${HOST_IP}/tcp/50000"
    echo "  HTTP:   http://${HOST_IP}:50080"
    echo ""
fi
