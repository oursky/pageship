# Deploying Sites

## Deploying using `pageship` command

For managed-sites mode, you may deploy a site through `pageship` command.

## Configure app

If it's the first time deploying the app on server, create the at with the
server through `pageship apps create` command.

```
$ pageship apps create
API Server: https://api.example.com
App "..." is created.
$
```

Then apply the configuration file (`pageship.toml`) through `pageship apps configure`.

```
$ pageship apps configure
Configured app "...".
```

You can reset the client side config using `pageship config reset`

```
$ pageship config reset
Reset client config: y
INFO   Client config reset.
```

## Deploy site

By default, each app has a default `main` site. Other sites can be configured in
`pageship.toml`.

To deploy to a site, use `pageship deploy` command with `site` parameter.

```
$ pageship deploy --site main
Deploy to site "main" of app "...": y
  INFO   Collecting files...
  INFO   69 files found. Tarball size: 1.0 MB
  INFO   Setting up deployment 'tmytb2i'...
uploading 100%
  INFO   Activating deployment...
  INFO   You can access the deployment at: ...
  INFO   Done!
```

To deploy as a preview deployment, omit the `site` parameter. For details,
refer to [Preview Deployment](./features/preview-deployment.md) guide.

```
$ pageship deploy --site main
Deploy to app "...": y
  INFO   Collecting files...
  INFO   69 files found. Tarball size: 1.0 MB
  INFO   Setting up deployment 'ztyflzy'...
  INFO   Site not specified; deployment would not be assigned to site
uploading 100%
  INFO   You can access the deployment at: ...
  INFO   Done!
```

## Deploying single site

For single-site/unmanaged-sites mode, you may deploy a site by copying the site
files to the site directory.

You may copy the site files using `rsync` with SSH access to the server.
Assuming the current directory contains the site, and the site directory is
located at `/var/pageship` on the server:
```
rsync -avh site/ /var/pageship/ --delete
```

Note that the deployment is not atomic - a visitor of the site may see
inconsistent content during the deployment. For atomic deployment, a
[server in managed sites mode](setup/managed-sites.md) is needed.
