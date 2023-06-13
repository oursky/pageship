# Getting Started

## Prerequisites

Pageship is a plain Go application, so developing Pageship only requires a
modern version (min 1.20) of Go.

For database migration, we used
[golang-migrate/migrate](https://github.com/golang-migrate/migrate) as library.
Only creating new database migrations requires installing the tool.

## Setup environment

Copy `.env.example` to `.env` and adjust as needed.
We used [`direnv`](direnv.net) to help setup required environment variables.

By default, the local data is stored in `data.local`.

## Running in single site mode

```sh
go run ./cmd/pageship serve examples/main
```

Open the site at `http://localtest.me:8000/`

## Running in unmanaged sites mode

```sh
go run ./cmd/pageship serve examples
```

Open the sites at `http://localtest.me:8000/` or `http://dev.localtest.me:8000/`

## Running in managed sites mode

```sh
go run ./cmd/controller start
```

Setup pageship command to use `http://api.localtest.me:8001` as the API server.
