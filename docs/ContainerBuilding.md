# Container Building Guide

This guide explains how to build Docker containers for the Colonies server, including multi-platform builds for different architectures.

## Prerequisites

- Docker 20.10 or later
- Docker buildx plugin (for multi-platform builds)
- Git (for version information)

## Quick Start

### Building for Local Architecture

Build a container image for your current platform (fastest option):

```bash
make container
```

This creates an image tagged as `colonyos/colonies` for your local architecture (typically `linux/amd64`).

### Building Multi-Platform Images

Build for both AMD64 and ARM64 architectures:

```bash
make container-multiplatform
```

This builds images for `linux/amd64` and `linux/arm64` using Docker buildx.

### Building and Pushing to Registry

Build multi-platform images and push to Docker Hub or your registry:

```bash
make container-multiplatform-push
```

This builds for both architectures and pushes with tags:
- `colonyos/colonies` (latest)
- `colonyos/colonies:v1.9.0` (version from Makefile)

## Docker Buildx Setup

Multi-platform builds require Docker buildx with QEMU support. If buildx is not available, follow these setup steps:

### 1. Install Docker Buildx

**Method 1: Using Package Manager (if available)**
```bash
# On Ubuntu/Debian
sudo apt-get install docker-buildx-plugin
```

**Method 2: Manual Installation**
```bash
# Create plugin directory
mkdir -p ~/.docker/cli-plugins

# Download buildx (replace version as needed)
curl -SL https://github.com/docker/buildx/releases/download/v0.18.0/buildx-v0.18.0.linux-amd64 \
  -o ~/.docker/cli-plugins/docker-buildx

# Make executable
chmod +x ~/.docker/cli-plugins/docker-buildx

# Verify installation
docker buildx version
```

### 2. Install QEMU for Cross-Platform Emulation

Enable building for ARM64 and other architectures on x86 machines:

```bash
docker run --privileged --rm tonistiigi/binfmt --install all
```

This installs QEMU emulators for multiple architectures including:
- linux/arm64
- linux/arm/v7
- linux/riscv64
- linux/ppc64le
- linux/s390x

### 3. Create Buildx Builder

Create a dedicated builder instance with multi-platform support:

```bash
# Create and bootstrap the builder
docker buildx create --name multiplatform-builder --driver docker-container --bootstrap --use

# Verify available platforms
docker buildx inspect multiplatform-builder
```

You should see output showing support for multiple platforms:
```
Platforms: linux/amd64, linux/amd64/v2, linux/amd64/v3, linux/arm64, linux/riscv64, ...
```

## Build Process Details

### Dockerfile Structure

The Colonies container uses a multi-stage build:

1. **Builder Stage** - Builds the Go binary
   - Based on `golang:1.24-alpine`
   - Downloads dependencies with `go mod download`
   - Compiles with build version and timestamp
   - Produces a statically-linked binary (CGO_ENABLED=0)

2. **Runtime Stage** - Minimal runtime image
   - Based on `alpine:latest`
   - Contains only the compiled binary
   - Optimized for small image size

### Build Arguments

The build process accepts two arguments for versioning:

- `VERSION` - Git commit hash (from `git rev-parse --short HEAD`)
- `BUILDTIME` - ISO 8601 timestamp (from `date -u '+%Y-%m-%dT%H:%M:%SZ'`)

These are embedded into the binary via `-ldflags` during compilation.

### Build Performance

Build times vary by platform:

- **linux/amd64**: ~10-15 seconds (native build)
- **linux/arm64**: ~90-120 seconds (emulated on x86)

Subsequent builds are faster due to Docker layer caching.

## Advanced Usage

### Custom Image Tags

Override the default image name and tags:

```bash
# Custom build image name
BUILD_IMAGE=myregistry/colonies make container

# Custom push tags
BUILD_IMAGE=myregistry/colonies PUSH_IMAGE=myregistry/colonies:v2.0.0 make container-multiplatform-push
```

### Building for Specific Platforms

Build for a single non-native platform:

```bash
docker buildx build \
  --platform linux/arm64 \
  --build-arg VERSION=$(git rev-parse --short HEAD) \
  --build-arg BUILDTIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ') \
  -t colonyos/colonies:arm64 \
  --load \
  .
```

Note: `--load` loads the image into your local Docker daemon (only works with single platform).

### Building for Additional Platforms

Build for more architectures beyond amd64/arm64:

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64,linux/arm/v7,linux/riscv64 \
  --build-arg VERSION=$(git rev-parse --short HEAD) \
  --build-arg BUILDTIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ') \
  -t colonyos/colonies \
  --push \
  .
```

### Inspecting Built Images

View information about multi-platform images in the build cache:

```bash
# List builders and their platforms
docker buildx ls

# Inspect a specific builder
docker buildx inspect multiplatform-builder

# View build cache
docker buildx du
```

### Exporting Images

Export multi-platform images to OCI format:

```bash
# Export to tar archive
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t colonyos/colonies \
  -o type=oci,dest=colonies-multi.tar \
  .

# Extract to directory
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t colonyos/colonies \
  -o type=oci,dest=colonies-oci/ \
  .
```

## Troubleshooting

### "buildx is not a docker command"

Docker buildx is not installed. Follow the [Docker Buildx Setup](#docker-buildx-setup) section.

### "failed to solve: platform not supported"

QEMU emulation is not installed. Run:

```bash
docker run --privileged --rm tonistiigi/binfmt --install all
```

### "no builder instance found"

Create a buildx builder:

```bash
docker buildx create --name multiplatform-builder --driver docker-container --bootstrap --use
```

### Build Cache Issues

Clear the build cache if experiencing issues:

```bash
# Prune build cache
docker buildx prune -a -f

# Remove and recreate builder
docker buildx rm multiplatform-builder
docker buildx create --name multiplatform-builder --driver docker-container --bootstrap --use
```

### ARM64 Build Extremely Slow

This is expected when building ARM64 images on x86 hardware due to QEMU emulation. Consider:

- Using a native ARM64 builder (e.g., AWS Graviton, Apple Silicon)
- Setting up a remote buildx builder on ARM64 hardware
- Using cloud build services like Docker Hub automated builds

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build Multi-Platform Container

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up QEMU
        uses: docker/setup-qemu-action@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Build and push
        run: make container-multiplatform-push
```

### GitLab CI Example

```yaml
build-multiplatform:
  image: docker:latest
  services:
    - docker:dind
  before_script:
    - apk add --no-cache git make curl
    - docker run --privileged --rm tonistiigi/binfmt --install all
    - docker buildx create --name mybuilder --use
  script:
    - make container-multiplatform-push
  only:
    - tags
```

## Best Practices

1. **Use Multi-Stage Builds** - Keep final images small by separating build and runtime stages
2. **Cache Dependencies** - Copy `go.mod` and `go.sum` before source code for better layer caching
3. **Statically Link Binaries** - Use `CGO_ENABLED=0` for portable binaries
4. **Version Everything** - Embed version information in binaries
5. **Tag Consistently** - Use semantic versioning for image tags
6. **Test All Platforms** - Run basic tests on each architecture after building
7. **Automate Builds** - Set up CI/CD for automatic multi-platform builds on releases

## See Also

- [Installation Guide](Installation.md) - Installing Colonies server
- [Configuration](Configuration.md) - Configuring the server
- [HADeployment](HADeployment.md) - Production deployment with Docker Compose and Kubernetes
- [Makefile](../Makefile) - Complete build targets reference
