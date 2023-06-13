# Setup Server

Pageship supports three modes of deployment:


## Single Site

Single site mode is the simplest deployment mode.
It hosts a single static site in a server, and pages can be updated by directly
copying the files to the data directory.

If you'd like to use the `pageship` command for deployment, or uses advanced
features like preview deployment, please use managed sites mode.

See the [setup guide for single site mode](./setup/single-site.md)


## Unmanaged Sites

Unmanaged sites mode hosts multiple static site on different sub-domains in a
server. Static sites can be created by creating a site directory in the 
data directory, and pages can be updated by copying the files to the site
directory.

If you'd like to use the `pageship` command for deployment, or uses advanced
features like preview deployment, please use managed sites mode.

See the [setup guide for unmanaged sites mode](./setup/unmanaged-sites.md)


## Managed Sites

Managed sites mode provides advanced features like preview deployment, GitHub
Actions integration, but is more complex to setup.

You can use `pageship` command to deploy to server and manage sites using this
mode.

See the [setup guide for managed sites mode](./setup/unmanaged-sites.md)

