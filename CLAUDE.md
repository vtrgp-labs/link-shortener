# Link Shortener

A simple link shortener web app written in Go. The goal is to serve as a learning project and a test bed for the AI Factory autonomous agent pipeline.

## What it does

- Accepts a long URL via a web form and returns a short code (e.g. `http://localhost:8080/abc123`)
- Redirects `GET /:code` to the original URL
- Stores all links in a local SQLite database
- Serves a minimal HTML frontend (no separate JS framework — plain HTML + a little CSS served by Go itself)

## Stack

- **Language:** Go (standard library only where possible)
- **Database:** SQLite via `modernc.org/sqlite` (pure Go driver, no CGO required)
- **HTTP:** `net/http` from stdlib
- **Tests:** `testing` package from stdlib

## Project structure

```
link-shortener/
├── main.go           # entry point, server setup
├── handler.go        # HTTP handlers (shorten, redirect, index)
├── store.go          # SQLite storage layer
├── store_test.go     # tests for the storage layer
├── handler_test.go   # tests for the HTTP handlers
├── static/
│   └── index.html    # the web form
└── go.mod
```

## Rules for agents

- Keep it simple. No frameworks, no ORMs, no unnecessary abstractions.
- All business logic must have tests.
- Short codes are 6-character random alphanumeric strings.
- The server listens on port 8080 by default, configurable via `PORT` env var.
- Return proper HTTP status codes (301 for redirects, 400 for bad input, 404 for unknown codes).
- Do not add features not described in the issue.
