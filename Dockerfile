# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Get build variables
ARG VERSION
ARG BUILDTIME
ENV VERSION=${VERSION}
ENV BUILDTIME=${BUILDTIME}

# Build the binary for the target architecture
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X 'main.BuildVersion=${VERSION}' -X 'main.BuildTime=${BUILDTIME}'" \
    -o /colonies \
    ./cmd/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /

# Copy the binary from builder
COPY --from=builder /colonies /bin/colonies

CMD ["colonies", "server", "start"]
