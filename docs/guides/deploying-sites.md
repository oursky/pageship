# Deploying Sites

## Deploying using `pageship` command

For managed sites mode, you may deploy a site through `pageship` command.

## Configure app

## Deploy site

## Deploying single site

For single site/unmanaged sites mode, you may deploy a site by copying the site
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
