# Service Examples

This directory contains example JSON files for ColonyOS services.

## Files

- `executor-deployment-definition.json` - ServiceDefinition for ExecutorDeployment kind
- `executor-deployment-instance.json` - Example ExecutorDeployment service instance

## Documentation

For complete documentation on services, see:

- **[Services.md](../../docs/Services.md)** - Complete guide to service management
- **[SchemaValidation.md](../../docs/SchemaValidation.md)** - Schema validation guide

## Quick Start

```bash
# Add the ServiceDefinition (colony owner only)
export COLONIES_PRVKEY=${COLONIES_COLONY_PRVKEY}
colonies service definition add --spec executor-deployment-definition.json

# Add a service instance
colonies service add --spec executor-deployment-instance.json

# Check service status
colonies service get --name web-server-deployment
```
