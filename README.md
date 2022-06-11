# go-bookmarks
A standalone bookmark / url shortening app.

# Build
go build

# Run
./go-bookmarks

# Usage

## via curl

### Create a bookmark
```bash
curl -X POST http://0:4912/api/v1/create -d name=golang-getting-started -d tags=golang,tutorial -d url=https://gobyexample.com/
```
### List all tags
```bash
curl http://0:4912/api/v1/tags
```
### List tag by name
eg: list all bookmarks tagged golang

```bash
curl http://0:4912/api/v1/tags/golang
```

## Roadmap
1. CLI
2. UI (standalone frontend in react or vue)
