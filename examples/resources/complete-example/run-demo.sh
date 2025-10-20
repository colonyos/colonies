#!/bin/bash

set -e

echo "===================================="
echo "ExecutorDeployment CRD Demo"
echo "===================================="
echo ""

# Build the example
echo "Building example..."
go build -o crd-example .

echo ""
echo "Running demo mode..."
echo ""
./crd-example -mode demo

echo ""
echo ""
echo "===================================="
echo "Demo Complete!"
echo "===================================="
echo ""
echo "To run with a real ColonyOS server:"
echo ""
echo "1. Start ColonyOS with Docker Compose:"
echo "   docker-compose up -d"
echo ""
echo "2. Wait for services to be healthy:"
echo "   docker-compose ps"
echo ""
echo "3. Register the CRD:"
echo "   ./crd-example -mode register-crd"
echo ""
echo "4. Start the controller (in another terminal):"
echo "   ./crd-example -mode controller"
echo ""
echo "5. Submit a CustomResource:"
echo "   ./crd-example -mode submit"
echo ""
echo "6. Check the controller logs to see reconciliation"
echo ""
