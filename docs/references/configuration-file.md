# Configuration file

## `pageship.toml`

`pageship.toml` is the main configuration file used by Pageship. All paths
referenced in `pageship.toml` is resolved relative to its location.

### `app` section

The `app` section defines the app config when hosted in managed-sites mode
server.

- `app.id`: Unique ID of the app in the server.
- `app.team`: ACL rules controlling API access of app management.
    - `app.team[].access`: Access level of the actor matching the rule (highest level applies):
        - `reader`: read-only access to app metadata (e.g. list of deployments/sites)
        - `deployer`: access neccessary for deploying sites
        - `admin`: full access to the app
- `app.defaultSite`: The name of main site (default to `main`).
- `app.sites`: The available sites in the app, the main site can be accessed
  through the app domain, while other sites is accessed through a subdomain.
      - `app.sites[].name`: the site name, cannot be used with pattern.
      - `app.sites[].pattern`: the site name pattern, cannot be used with name.
  subdomain.
- `app.deployments`: Configuration for preview deployments
    - `access`: ACL rules controlling access of preview deployments.
    - `ttl`: the lifetime of a preview deployment (default to `24h`)
- `app.domains`: Configuration for custom domains
    - `domain`: The custom domain to use
    - `site`: The site name associated the custom domain

### `site` section

The `site` section defines the site config.
- `site.public`: The path to site directory
- `site.access`: ACL rules controlling access of site


## `sites.toml`

`sites.toml` is used for unmanaged-sites mode. It defines the location and
resolution of multiple apps.
- `sites.<name>`: The list of sites to serve
    - `sites.<name>.context`: directory of the site.
