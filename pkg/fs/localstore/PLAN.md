# ColonyFS: Embedded File Storage Plan

## Goal

Replace the MinIO/S3 dependency with an embedded file storage backend, making ColonyOS fully self-contained (no external services required when using the embedded database + embedded file storage).

## Architecture

```
  Protocol: "s3"
  Client (Go/TS) --metadata--> Colonies Server           Client --direct--> MinIO

  Protocol: "coloniesfs"
  Client (Go/TS) --metadata + file data--> Colonies Server --> local filesystem
```

When `Protocol = "coloniesfs"`, file data flows through the Colonies server instead of directly to S3. The server stores files on its local filesystem. No external dependencies needed.

Both protocols coexist: the server can serve files from local storage while clients can still use S3 for other files.

## Current Status

### Done

- **LocalObjectStore** (`pkg/fs/localstore/localstore.go`)
  - Filesystem-backed storage with colony isolation
  - Put, Get, GetRange, Remove, Exists, Stat, List, DiskUsage
  - Chunked upload support (PutChunk, AssembleChunks, GetChunkStatus)
  - Path traversal protection
  - 32 KB buffer streaming (constant memory)
  - ObjectStore interface defined in `pkg/fs/localstore/objectstore.go`
  - Unit tests in `localstore_test.go`

- **Server-side HTTP endpoints** (`pkg/server/handlers/file/data_handlers.go`)
  - `PUT /api/fs/{objectName}` - upload (single request, streamed)
  - `GET /api/fs/{objectName}` - download (streamed)
  - `DELETE /api/fs/{objectName}` - delete
  - ECDSA signature verification via `X-Colonies-Payload` / `X-Colonies-Signature` headers
  - Colony membership validation

- **Reference model** (`pkg/core/file.go`)
  - `Reference.IsColonyFS()` helper
  - `Protocol = "coloniesfs"` recognized throughout

- **FSClient** (`pkg/fs/fs_client.go`)
  - `CreateFSClientWithColonyFS()` constructor
  - Upload/download/remove via coloniesfs protocol
  - Signed payload auth for all coloniesfs HTTP requests
  - S3 client lazy-initialized (only created when S3 operations are needed)
  - S3 code moved to `pkg/fs/s3/` subpackage
  - Tests that need S3 skip with `requireS3(t)` when `COLONIES_FILE_STORAGE_TYPE=coloniesfs`

- **Server configuration** (`pkg/server/server.go`)
  - `COLONIES_FILE_STORAGE_TYPE` env var (`s3` or `coloniesfs`)
  - `COLONIES_FILE_STORAGE_DIR` env var (local path for coloniesfs)
  - Server auto-creates LocalObjectStore and registers routes when coloniesfs is configured

- **CLI** (`internal/cli/fs.go`)
  - Detects `COLONIES_FILE_STORAGE_TYPE` and creates appropriate FSClient
  - `standalone.env` for standalone (no external services) mode

- **Embedded DB defaults** (`internal/cli/server.go`, `internal/cli/db.go`)
  - `DataDir` defaults to `~/.colonies` when using embedded DB
  - `EtcdDataDir` defaults to `~/.colonies/etcd`
  - Server auto-sets server ID for embedded DB on first startup

- **Docker compose** (`docker-compose-embedded.yml`)
  - Single container, single volume, no PostgreSQL, no MinIO

- **Makefile**
  - `pkg/fs` and `pkg/fs/localstore` tests always run
  - `pkg/fs/s3` tests only run when not using coloniesfs

### Key Files

| File | Role |
|------|------|
| `pkg/fs/localstore/objectstore.go` | ObjectStore interface + ObjectInfo |
| `pkg/fs/localstore/localstore.go` | Filesystem-backed ObjectStore |
| `pkg/fs/localstore/localstore_test.go` | LocalObjectStore unit tests |
| `pkg/fs/s3/s3.go` | S3Client (MinIO wrapper) |
| `pkg/fs/s3/s3_test.go` | S3Client tests (need MinIO) |
| `pkg/fs/fs_client.go` | FSClient - sync logic, supports both protocols |
| `pkg/fs/fs_client_test.go` | FSClient tests (S3-dependent ones skip without MinIO) |
| `pkg/server/handlers/file/data_handlers.go` | ColonyFS upload/download/delete HTTP handlers |
| `pkg/server/server.go` | Server setup, registers ColonyFS routes |
| `pkg/core/file.go` | File, Reference, S3Object structs + IsColonyFS() |
| `internal/cli/fs.go` | CLI file commands, protocol-aware FSClient creation |
| `standalone.env` | Standalone mode environment (embedded DB + coloniesfs) |
| `docker-compose-embedded.yml` | Docker compose for embedded mode |

### Not Yet Implemented

- **Chunked upload HTTP protocol** - init/chunk/complete flow via HTTP endpoints (LocalObjectStore supports it, but no HTTP handlers or client-side logic yet)
- **Range/resumable downloads** - LocalObjectStore has GetRange, but HTTP handler doesn't expose Range header support
- **Nonce/timestamp replay protection** - not implemented in HTTP handlers
- **Disk space checks / quota enforcement** - not implemented
- **Stale upload cleanup** - background goroutine to clean abandoned staging dirs
- **Health check extensions** - coloniesfs status in health endpoint
- **Migration tooling** - `colonies fs migrate` CLI command
- **TypeScript client** - upload/download methods for coloniesfs protocol

## HTTP Protocol

### Implemented Endpoints

```
PUT    /api/fs/{objectName}           # upload file (single request, streamed)
GET    /api/fs/{objectName}           # download file (streamed)
DELETE /api/fs/{objectName}           # delete file
```

### Authentication

Authentication uses the same ECDSA signing mechanism as all other ColonyOS requests. The signed payload is carried in HTTP headers since the body is binary data.

**Request headers:**
```
X-Colonies-Payload:   <base64 encoded JSON>
X-Colonies-Signature: <hex ECDSA signature of the payload>
```

### Future Endpoints (Not Implemented)

```
POST   /api/fs/{objectName}/init      # initiate chunked upload
PUT    /api/fs/{objectName}/{chunk}   # upload a chunk
POST   /api/fs/{objectName}/complete  # finalize upload, verify checksum
GET    /api/fs/{objectName}/status    # query chunk upload status
```

## Future Work

### Chunked Uploads

Large files should be uploaded in fixed-size chunks (default 64 MB) for constant memory usage and resume support. The LocalObjectStore already implements PutChunk/AssembleChunks/GetChunkStatus - HTTP handlers and client-side logic are needed.

### Replay Protection

Every request should include a client-generated nonce and timestamp. Server maintains a short-lived nonce cache to reject replayed requests.

### Disk Space Management

- Pre-upload disk space check (reject with HTTP 507)
- Optional per-colony quotas
- Optional max file size limit

### Migration Tooling

CLI commands for migrating files between S3 and coloniesfs backends:

```bash
colonies fs migrate --from s3 --to coloniesfs --colony my-colony
colonies fs migrate --from coloniesfs --to s3 --colony my-colony --dry-run
```

### TypeScript Client

Add upload/download methods to `colonies-ts/src/client.ts` using the same ECDSA signing the client already implements.
