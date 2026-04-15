# test-go-resolver

A simple DNS resolution service exposed via HTTP.

## Usage

```
curl localhost:8080/resolve/google.com
```

## Docker

### Build

```bash
docker build --platform=linux/amd64 -t quay.io/matskiv/test-go-resolver:latest .
```

### Push

```bash
docker push quay.io/matskiv/test-go-resolver:latest
```
