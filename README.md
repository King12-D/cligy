# cligy — File Hosting CDN

A lightweight, self-hosted file hosting CDN built in Go using Gin. Supports local filesystem and S3-compatible storage backends.

## Features

- Upload, serve, list, delete files via REST API
- Immutable caching — ETag + `Cache-Control: public, max-age=31536000, immutable`
- 304 Not Modified responses via in-memory ETag cache
- Random hex file IDs (no sequential enumeration)
- Pluggable storage — local FS or S3 (MinIO, AWS, DigitalOcean Spaces, etc.)
- Pluggable cache — swap memory for Redis, Memcached
- Content-Type detection from upload headers
- Force-download endpoint

## Quick Start

### Local filesystem

```bash
go run ./cmd
# CDN listening on 0.0.0.0:8080 (storage: ./data)
```

### S3-compatible (MinIO)

```bash
S3_BUCKET=my-bucket \
  S3_ENDPOINT=http://localhost:9000 \
  S3_REGION=us-east-1 \
  go run ./cmd
# CDN listening on 0.0.0.0:8080 (storage: s3://my-bucket)
```

### AWS S3

```bash
export AWS_ACCESS_KEY_ID=AKIA...
export AWS_SECRET_ACCESS_KEY=...
S3_BUCKET=my-cdn-bucket S3_REGION=us-west-2 go run ./cmd
```

## Configuration

| Env Variable    | Default     | Description                                |
|-----------------|-------------|--------------------------------------------|
| `PORT`          | `8080`      | Server port                                |
| `STORAGE_PATH`  | `./data`    | Local storage directory                    |
| `S3_BUCKET`     | —           | S3 bucket name (enables S3 backend)        |
| `S3_PREFIX`     | —           | S3 key prefix (e.g. `cdn/production`)      |
| `S3_ENDPOINT`   | —           | Custom S3 endpoint (MinIO, DO Spaces, etc) |
| `S3_REGION`     | —           | AWS region (e.g. `us-east-1`)              |
| `DEBUG`         | `false`     | Enable Gin debug mode (verbose logging)    |

Set `S3_BUCKET` to use S3; omit it for local filesystem storage. AWS credentials follow the [standard SDK chain](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials): env vars → `~/.aws/credentials` → IAM role.

## API

### Upload a file

```bash
curl -X POST http://localhost:8080/api/upload \
  -F "file=@photo.jpg"
```

Response `201`:
```json
{
  "id": "a1b2c3d4e5f6...",
  "name": "photo.jpg",
  "size": 284567,
  "content_type": "image/jpeg",
  "etag": "\"abcd1234...\"",
  "created_at": "2026-07-30T17:23:11.219Z",
  "url": "localhost:8080/raw/a1b2c3d4e5f6..."
}
```

### List all files

```bash
curl http://localhost:8080/api/files
```

Response `200`:
```json
{
  "files": [
    {
      "id": "a1b2c3d4e5f6...",
      "name": "photo.jpg",
      "size": 284567,
      "content_type": "image/jpeg",
      "etag": "\"abcd1234...\"",
      "created_at": "2026-07-30T17:23:11.219Z"
    }
  ]
}
```

### Serve a file (with caching headers)

```bash
curl http://localhost:8080/raw/<file-id> --output photo.jpg
```

Returns `ETag`, `Cache-Control: public, max-age=31536000, immutable`, `Content-Type`, `Content-Length`. Sends `304 Not Modified` on matching `If-None-Match`.

### Force download

```bash
curl http://localhost:8080/api/files/<file-id>/download --output photo.jpg
```

Same as serve but with `Content-Disposition: attachment`.

### Delete a file

```bash
curl -X DELETE http://localhost:8080/api/files/<file-id>
```

Response `204 No Content`.

### Health check

```bash
curl http://localhost:8080/health
```

Response `200`:
```json
{"status": "ok"}
```

## Project Structure

```
cmd/main.go                Entry point
internal/
  model/file.go            FileInfo model
  storage/
    storage.go             Storage interface
    local.go               Local filesystem implementation
    s3.go                  S3 implementation
  cache/
    cache.go               Cache interface
    memory.go              In-memory ETag cache
  handler/
    upload.go              POST /api/upload
    serve.go               GET /raw/:id, /api/files/:id, /api/files/:id/download
    admin.go               GET /api/files, DELETE /api/files/:id, /health
  server/
    server.go              Gin engine, routing, backend selection
```

## Adding a Storage Backend

Implement the `Storage` interface:

```go
type Storage interface {
    Save(name string, contentType string, reader io.Reader) (*model.FileInfo, error)
    Get(id string) (io.ReadCloser, *model.FileInfo, error)
    Delete(id string) error
    List() ([]*model.FileInfo, error)
}
```

Create `internal/storage/gcs.go` (example):

```go
package storage

import "cloud.google.com/go/storage"

type GCSStorage struct {
    client *storage.Client
    bucket string
}

func NewGCSStorage(client *storage.Client, bucket string) *GCSStorage {
    return &GCSStorage{client: client, bucket: bucket}
}

// implement Save, Get, Delete, List using GCS APIs
```

Then wire it in `internal/server/server.go` by passing it to `handler.New()`.

## Adding a Cache Backend

Implement the `Cache` interface:

```go
type Cache interface {
    Get(key string) ([]byte, bool)
    Set(key string, data []byte)
    Delete(key string)
    Clear()
}
```

## License

MIT
