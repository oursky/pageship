# Managed-sites mode

Managed-sites mode hosts a multiple static sites in a server. The sites can
be managed/deployed through `pageship` command.

## Prerequisites

Pageship in managed-sites mode requires a database to store management metadata,
and object storage to store the actual site content.

For database, we supports:
- SQLite
- PostgreSQL

For object storage, we supports:
- Filesystem
- Azure Blob
- GCS
- S3

For simplicity, Pageship uses SQLite with Filesystem storage by default. You
may configure alternate database/object storage through configuration.

## Docker compose

```yaml
version: "3"
services:
  controller:
    image: ghcr.io/oursky/pageship-controller
    volumes:
      - data:/var/pageship
    environment:
      - PAGESHIP_MIGRATE=true
      - PAGESHIP_DATABASE_URL=sqlite:///var/pageship/data.db
      - PAGESHIP_STORAGE_URL=file:///var/pageship/storage?create_dir=true
      - PAGESHIP_HOST_PATTERN=http://*.localhost:8001
    ports:
      - "8001:8001"
```

## Configuration

The host pattern (`PAGESHIP_HOST_PATTERN`) specify how Pageship should map the
request host to a site. The wildcard part would be extracted as the site name.

By default, database schema is upgraded automatically on start. To disable it,
set `PAGESHIP_MIGRATE` to false. You may run migration manually using the
`migrate` subcommnad.

The database can be specified by `PAGESHIP_DATABASE_URL`.
- For SQLite, provide the path to the database file like `sqlite:///var/pageship/data.db`.
- For PostgreSQL: provide the DSN like `postgres://postgres:postgres@db:5432/postgres?sslmode=disable`.

The object storage can be specified by `PAGESHIP_STORAGE_URL`. Refer to
documentation of [gocloud](https://gocloud.dev/howto/blob/) for URL format of
different providers.

Refer to [Server configuration](../../references/server-configuration.md) for
detailed reference on configuration.

## What's Next

- [Deploy your site](../deploying-sites.md#deploying-using-pageship-command)
- [Configure access control](../features/access-control.md)
- [Configure automatic TLS](../features/automatic-tls.md)
