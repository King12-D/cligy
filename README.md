# cligy — File Hosting CDN

A lightweight, self-hosted file hosting CDN built in Go using Gin.

## Features

- Upload, serve, list, and delete files via REST API
- Immutable caching with ETag and `Cache-Control: public, max-age=31536000, immutable`
- In-memory ETag cache for 304 Not Modified responses
- Random hex file IDs (no sequential enumeration)
- Extensible storage and cache interfaces (swap local FS for S3, memory for Redis)
- Content-Type detection from upload headers
- Force-download endpoint

## Quick Start

```bash
go run ./cmd
# CDN listening on 0.0.0.0:8080
```

## Configuration

| Env Variable   | Default     | Description                |
|----------------|-------------|----------------------------|
| `PORT`         | `8080`      | Server port                |
| `STORAGE_PATH` | `./data`    | Directory for stored files |
| `DEBUG`        | `false`     | Enable Gin debug mode      |

## API

### Upload a file

```bash
curl -X POST http://localhost:8080/api/upload \
  -F "file=@photo.jpg"
```

### List all files

```bash
curl http://localhost:8080/api/files
```

### Serve a file

```bash
curl http://localhost:8080/raw/<file-id> --output photo.jpg
```

### Force download

```bash
curl http://localhost:8080/api/files/<file-id>/download --output photo.jpg
```

### Delete a file

```bash
curl -X DELETE http://localhost:8080/api/files/<file-id>
```

### Health check

```bash
curl http://localhost:8080/health
```

## Project Structure

```
cmd/main.go              # Entry point
internal/
  model/file.go          # FileInfo struct
  storage/storage.go     # Storage interface
  storage/local.go       # Local filesystem implementation
  cache/cache.go         # Cache interface
  cache/memory.go        # In-memory cache
  handler/               # HTTP handlers (upload, serve, admin)
  server/server.go       # Gin engine + router setup
```

## Extending

Implement `storage.Storage` to add S3, GCS, or any backend:

```go
type Storage interface {
    Save(name string, contentType string, reader io.Reader) (*model.FileInfo, error)
    Get(id string) (io.ReadCloser, *model.FileInfo, error)
    Delete(id string) error
    List() ([]*model.FileInfo, error)
}
```

Implement `cache.Cache` to use Redis, Memcached, etc.:

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
