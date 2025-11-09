# Blueprint Examples

This directory contains example JSON files for ColonyOS blueprints.

## Files

- `executor-deployment-definition.json` - BlueprintDefinition for ExecutorDeployment kind
- `executor-deployment-instance.json` - Example ExecutorDeployment blueprint instance

## Documentation

For complete documentation on blueprints, see:

- **[Blueprints.md](../../docs/Blueprints.md)** - Complete guide to blueprint management
- **[SchemaValidation.md](../../docs/SchemaValidation.md)** - Schema validation guide

## Quick Start

```bash
# Add the BlueprintDefinition (colony owner only)
export COLONIES_PRVKEY=${COLONIES_COLONY_PRVKEY}
colonies blueprint definition add --spec executor-deployment-definition.json

# Add a blueprint instance
colonies blueprint add --spec executor-deployment-instance.json

# Check blueprint status
colonies blueprint get --name web-server-deployment
```
