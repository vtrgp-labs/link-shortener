# Link Shortener

A simple link shortener web app written in Go.

## What it does

- Accepts a long URL via a web form and returns a short code (e.g. `http://localhost:8080/abc123`)
- Redirects short codes to the original URL
- Stores all links in a local SQLite database
- Serves a minimal HTML frontend built with plain HTML and CSS

## Stack

- **Language:** Go (standard library only where possible)
- **Database:** SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **HTTP:** `net/http` from stdlib

## Running

```bash
go run .
```

The server listens on port 8080 by default. Override with the `PORT` environment variable:

```bash
PORT=9090 go run .
```

## Testing

```bash
go test ./...
```
