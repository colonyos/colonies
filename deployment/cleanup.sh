#!/bin/bash

# Cleanup script to remove all Docker containers, networks, and volumes
# Use with caution - this will remove ALL Docker resources on your system

echo "========================================="
echo "Docker Cleanup Script"
echo "========================================="
echo ""

# Stop all running containers
echo "Stopping all running containers..."
docker stop $(docker ps -aq) 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ All containers stopped"
else
    echo "✓ No running containers found"
fi
echo ""

# Remove all containers
echo "Removing all containers..."
docker rm -f $(docker ps -aq) 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ All containers removed"
else
    echo "✓ No containers to remove"
fi
echo ""

# Remove all networks (except default ones)
echo "Removing all custom networks..."
docker network prune -f 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ All custom networks removed"
else
    echo "✓ No custom networks to remove"
fi
echo ""

# Remove all volumes
echo "Removing all volumes..."
docker volume prune -f 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ All volumes removed"
else
    echo "✓ No volumes to remove"
fi
echo ""

# Remove all images (optional - uncomment if you want to remove images too)
# echo "Removing all images..."
# docker image prune -a -f 2>/dev/null
# if [ $? -eq 0 ]; then
#     echo "✓ All images removed"
# else
#     echo "✓ No images to remove"
# fi
# echo ""

# System-wide cleanup
echo "Running system-wide cleanup..."
docker system prune -a -f --volumes 2>/dev/null
echo "✓ System cleanup complete"
echo ""

echo "========================================="
echo "Cleanup Complete!"
echo "========================================="
echo ""
echo "Summary:"
docker ps -a
echo ""
docker network ls
echo ""
docker volume ls
