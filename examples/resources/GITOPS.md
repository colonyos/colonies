# GitOps for ColonyOS Resources

This guide explains how to use GitOps to manage ColonyOS resources from Git repositories.

## Overview

GitOps enables you to define resources in a Git repository and automatically sync them to your ColonyOS cluster. This provides:

- Version control for your infrastructure
- Declarative resource management
- Automated synchronization
- Audit trail through Git history

## Setup

### 1. Add GitOps Configuration to ResourceDefinition

Add a `gitops` section to your ResourceDefinition spec:

```json
{
  "metadata": {
    "name": "deployments.example.io"
  },
  "spec": {
    "group": "example.io",
    "version": "v1",
    "names": {
      "kind": "Deployment",
      "plural": "deployments",
      "singular": "deployment"
    },
    "scope": "Namespaced",
    "handler": {
      "executorType": "deployment-controller",
      "functionName": "reconcile"
    },
    "gitops": {
      "repoURL": "https://github.com/example/deployments.git",
      "branch": "main",
      "path": "/resources",
      "interval": 300
    }
  }
}
```

### GitOps Configuration Fields

- **repoURL** (required): Git repository URL (HTTPS or SSH)
- **branch** (optional): Git branch to sync from (default: "main")
- **path** (optional): Path within the repository to sync (default: "/")
- **secretName** (optional): Name of secret containing Git credentials for private repos
- **interval** (optional): Sync interval in seconds (default: 300)

### 2. Create Resource Files in Git

In your Git repository, create JSON files defining your resources:

```json
{
  "kind": "Deployment",
  "metadata": {
    "name": "my-app-deployment",
    "namespace": "production"
  },
  "spec": {
    "image": "myapp:v1.2.3",
    "replicas": 3
  }
}
```

### 3. Sync Resources from Git

Use the CLI to sync resources:

```bash
# Dry run to see what would be synced
colonies resource sync --definition deployments.example.io --dry-run

# Sync resources from Git
colonies resource sync --definition deployments.example.io
```

## How It Works

1. The sync command fetches the ResourceDefinition from ColonyOS
2. It clones or pulls the Git repository specified in the GitOps configuration
3. It scans the specified path for resource files (*.json, *.yaml, *.yml)
4. Resources matching the ResourceDefinition's kind are loaded
5. Each resource is validated against the schema
6. Resources are created or updated in ColonyOS
7. The Git commit SHA and sync time are recorded in each resource's status

## Git Repository Structure

Your Git repository can be structured in several ways:

### Single Directory
```
/
├── app1-deployment.json
├── app2-deployment.json
└── app3-deployment.json
```

### Organized by Environment
```
/
├── production/
│   ├── app1-deployment.json
│   └── app2-deployment.json
└── staging/
    └── app1-deployment.json
```

In this case, set `path: "/production"` or `path: "/staging"` in the GitOps config.

### Multiple Resource Types
```
/
├── deployments/
│   ├── app1.json
│   └── app2.json
└── services/
    ├── service1.json
    └── service2.json
```

Each ResourceDefinition can have its own path.

## Private Repositories

For private Git repositories, you have several options:

### SSH Keys
Use SSH URL format and ensure the executor has SSH keys configured:
```json
{
  "gitops": {
    "repoURL": "git@github.com:example/deployments.git"
  }
}
```

### HTTPS with Token (Future Enhancement)
Store credentials in a secret and reference it:
```json
{
  "gitops": {
    "repoURL": "https://github.com/example/deployments.git",
    "secretName": "git-credentials"
  }
}
```

## Best Practices

1. **Use branches** for different environments (main, staging, dev)
2. **Tag releases** in Git to track deployments
3. **Enable branch protection** to prevent accidental changes
4. **Review changes** through pull requests before merging
5. **Set appropriate sync intervals** (300s for production, 60s for development)
6. **Use path filtering** to sync only relevant resources

## Monitoring Sync Status

Each synced resource includes GitSync metadata:

```json
{
  "gitSync": {
    "lastSyncTime": "2025-10-30T10:30:00Z",
    "lastCommitSHA": "abc123def456",
    "syncError": ""
  }
}
```

Check this status to verify when a resource was last synced and from which commit.

## Example Workflow

1. Create a ResourceDefinition with GitOps config:
   ```bash
   colonies resource definition add --spec gitops-example-definition.json
   ```

2. Create resource files in your Git repository

3. Commit and push to Git:
   ```bash
   git add resources/
   git commit -m "Add new deployments"
   git push origin main
   ```

4. Sync to ColonyOS:
   ```bash
   colonies resource sync --definition deployments.example.io
   ```

5. Verify resources were created:
   ```bash
   colonies resource ls --kind Deployment
   ```

## Troubleshooting

### Sync fails with "git clone failed"
- Check that the Git URL is correct
- Ensure you have network connectivity
- For private repos, verify SSH keys or credentials are configured

### Resources not being created
- Check that resource files are valid JSON/YAML
- Verify the `kind` field matches the ResourceDefinition
- Run with `--dry-run` to see what would be synced
- Check that resources pass schema validation

### Sync is slow
- Use `--depth 1` shallow clones (default behavior)
- Reduce the sync interval
- Use path filtering to limit files scanned
