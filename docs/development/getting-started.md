# Getting Started

## Prerequisites

Pageship is a plain Go application, so developing Pageship only requires a
modern version (min 1.20) of Go.

For database migration, we used
[golang-migrate/migrate](https://github.com/golang-migrate/migrate) as library.
Only creating new database migrations requires installing the tool.

## Setup environment

Copy `.env.example` to `.env` and adjust as needed.
We recommend to use [`direnv`](direnv.net) to help setup required environment variables.

Once you've setup direnv, please ensure to have following config to load dotenv.

```
# ~/.config/direnv/direnv.toml
[global]
load_dotenv = true
```

Load environment variables:

```
direnv allow .
```

## Database

By default, the database is `sqlite` and data is stored in `data.local` directory, please create the folder `./data.local/storage` before start development.

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
go run ./cmd/controller start --migrate
```

Setup pageship command to use `http://api.localtest.me:8001` as the API server.
