# Unmanaged sites

Unmanaged sites mode hosts multiple static sites in a server.

## Setup

First, prepare a directory to store the site data. The following instruction
assumes the data directory is located at `/var/pageship`.

Before continuing, create a directory for each sites in the data directory.
The directory layout should look like this:

```
/var/pageship
├── main                    # main
│   ├── pageship.toml
│   └── public
└── blogs                   # blogs
    ├── pageship.toml
    ├── public
    └── user                # blogs/user
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
      - PAGESHIP_HOST_PATTERN=http://*.localhost:8001
      - PAGESHIP_DEFAULT_SITE=main
    ports:
      - "8001:8001"
```

## Configuration

The host pattern (`PAGESHIP_HOST_PATTERN`) specify how Pageship should map the
request host to a site. The wildcard part would be extracted as the site name.

By default, the sites are resolved ad-hoc using the directory layout, and the
name of each site directories is used as the site name. A valid site must
contain a `pageship.toml` config file in the site directory.

For example, for the above shown directory layout, the sites would be reachable
at:
- `main`: http://localhost:8001 (matches default site `PAGESHIP_DEFAULT_SITE`)
- `blogs`: http://blogs.localhost:8001
- `blogs/user`: http://user.blogs.localhost:8001

Optionally, a static config file (`sites.toml`) can be created to specify the
available sites. For example:

```toml
# /var/pageship/sites.toml

[sites."main"]      # default site: http://localhost:8001
context="main"      # site directory: /var/pageship/main/

[sites."blogs"]         # http://blogs.localhost:8001
context="blogs/main"    # site directory: /var/pageship/blogs/main

[sites."user.blogs"]    # http://user.blogs.localhost:8001
context="blogs/user"    # site directory: /var/pageship/blogs/user

```

The site directory (`context`) is resolved relative to the location of config
file.

## What's Next

- [Deploy your site](../deploying-sites.md#deploying-single-site)
- [Configure access control to the sites](../features/access-control.md#access-control-for-sites)
- [Configure automatic TLS](../features/automatic-tls.md)
