# Getting Started

## Prerequisites

Pageship is a plain Go application, so developing Pageship only requires a
modern version (min 1.20) of Go.

For database migration, we used
[golang-migrate/migrate](https://github.com/golang-migrate/migrate) as library.
Only creating new database migrations requires installing the tool.

## Setup environment

Copy `.env.example` to `.env` and adjust as needed. Then run it via [`docker-compose`](../../docker-compose.yaml).

To run it without `docker-compose`, we recommend to use [`direnv`](https://github.com/direnv/direnv) to help setup required environment variables. Otherwise, expose `.env` variables to your system environment.

By default, the local data is stored in `data.local`.

And default domain is `http://*.localtest.me:8001` as value of PAGESHIP_HOST_PATTERN set in .env file

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

method 1: Run with docker-compose

```sh
docker-compose up -d
```

method 2: Run with go command

```sh
go run ./cmd/controller start
```

Setup pageship command to use `http://api.localtest.me:8001` as the API server with your github account.

(Note: Must enter absolute path for `SSH Key file`)

```sh
go run ./cmd/pageship login

GitHub user name: <your github user name>
API server: http://api.localtest.me:8001
SSH key file: /Home/yourUsername/.ssh/id_rsa

```

If you enter incorrect info and want to restore whole process please enter command:

```sh
go run ./cmd/pageship config reset

```
