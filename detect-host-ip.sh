#!/bin/bash
# This script attempts to detect the host machine's IP from within Docker

# Method 1: Use the default gateway (Docker host)
HOST_IP=$(ip route | grep default | awk '{print $3}')
echo "Method 1 (gateway): $HOST_IP"

# Method 2: Query external service (if internet available)
EXTERNAL_IP=$(curl -s ifconfig.me 2>/dev/null || curl -s icanhazip.com 2>/dev/null)
echo "Method 2 (external): $EXTERNAL_IP"

# Method 3: Get IP from DNS (if hostname resolves)
DNS_IP=$(getent hosts host.docker.internal 2>/dev/null | awk '{print $1}')
echo "Method 3 (host.docker.internal): $DNS_IP"

# Method 4: Parse from /proc/net/route
PROC_IP=$(awk '/^default/ {print $2}' /proc/net/route | xargs -I {} printf '%d.%d.%d.%d\n' 0x{} | cut -d'.' -f1-4)
echo "Method 4 (proc): $PROC_IP"
