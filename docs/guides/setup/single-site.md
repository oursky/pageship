# Single site

Single site mode hosts a single static site in a server.

## Setup

First, prepare a directory to store the site data. The following instruction
assumes the data directory is located at `/var/pageship`.

Before continuing, copy the site files to the data directory. The directory
layout should look like this:

```
/var/pageship
├── pageship.toml
└── public
```

### Docker compose

```yaml
version: "3"
services:
  controller:
    image: ghcr.io/oursky/pageship-server
    volumes:
      - /var/pageship:/var/pageship
    environment:
      - PAGESHIP_HOST_PATTERN=http://localhost:8001
    ports:
      - "8001:8001"
```

## What's Next

- [Deploy your site](../deploying-sites.md#deploying-single-site)
- [Configure access control to the sites](../features/access-control.md#access-control-for-sites)
- [Configure automatic TLS](../features/automatic-tls.md)
