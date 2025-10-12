#!/bin/bash

# Colonies Relay/Bootstrap Node Deployment Script
# This script builds and deploys a relay/bootstrap node for Colonies LibP2P networking

set -e

echo "════════════════════════════════════════════════════════════════"
echo "Colonies Relay/Bootstrap Node Deployment"
echo "════════════════════════════════════════════════════════════════"
echo

# Check if go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

echo "✓ Go found: $(go version)"
echo

# Build the relay node
echo "Building relay/bootstrap node..."
cd "$(dirname "$0")"
CGO_ENABLED=0 go build -o bin/relay-bootstrap ./cmd/relay-bootstrap/main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✓ Built: bin/relay-bootstrap"
echo

# Create systemd service file
echo "Creating systemd service file..."
cat > relay-bootstrap.service << 'EOF'
[Unit]
Description=Colonies Relay/Bootstrap Node
After=network.target

[Service]
Type=simple
User=colonies
WorkingDirectory=/opt/colonies-relay
ExecStart=/opt/colonies-relay/relay-bootstrap
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

echo "✓ Created relay-bootstrap.service"
echo

# Print deployment instructions
cat << 'INSTRUCTIONS'
════════════════════════════════════════════════════════════════
DEPLOYMENT STEPS
════════════════════════════════════════════════════════════════

1. On your PUBLIC-FACING server (VPS/cloud instance), run:

   # Create deployment directory
   sudo mkdir -p /opt/colonies-relay
   sudo useradd -r -s /bin/false colonies || true

2. Copy the binary to your server:

   scp bin/relay-bootstrap user@your-server:/tmp/
   ssh user@your-server "sudo mv /tmp/relay-bootstrap /opt/colonies-relay/ && sudo chmod +x /opt/colonies-relay/relay-bootstrap"

3. Copy the systemd service:

   scp relay-bootstrap.service user@your-server:/tmp/
   ssh user@your-server "sudo mv /tmp/relay-bootstrap.service /etc/systemd/system/"

4. Open firewall ports on your server:

   sudo ufw allow 4001/tcp
   sudo ufw allow 4001/udp

5. Start the service:

   sudo chown -R colonies:colonies /opt/colonies-relay
   sudo systemctl daemon-reload
   sudo systemctl enable relay-bootstrap
   sudo systemctl start relay-bootstrap

6. Check status and get configuration:

   sudo journalctl -u relay-bootstrap -f

   Look for the "CONFIGURATION INSTRUCTIONS" section in the logs.
   Copy the BOOTSTRAP_PEERS lines to your docker-compose.env file.

7. Important: Update the IP address in the BOOTSTRAP_PEERS to your PUBLIC IP!

   Example:
   If the log shows: /ip4/10.0.0.5/tcp/4001/p2p/12D3Koo...
   Change to:         /ip4/YOUR_PUBLIC_IP/tcp/4001/p2p/12D3Koo...

════════════════════════════════════════════════════════════════
QUICK TEST (Run relay locally first)
════════════════════════════════════════════════════════════════

You can test the relay locally before deploying:

   ./bin/relay-bootstrap

This will start the relay and show you the configuration strings.
Press Ctrl+C to stop.

════════════════════════════════════════════════════════════════
INSTRUCTIONS

echo
echo "Deployment package ready!"
echo "Binary: bin/relay-bootstrap"
echo "Service: relay-bootstrap.service"
echo
