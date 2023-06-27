# Preview Deployment

Pageship supports preview deployments. When a deployment is created without
assigning to a site, it is treated as preview deployments.

Preview deployments has limited lifetime, and expires after a period not
assigned to a site. This period can be configured in `pageship.toml`:
```toml
[app.deployments]
ttl = "24h"     # expires after 24 hours.
```

An expired preview deployment is inaccessible and deleted automatically after
some time.

## Access Control

By default, preview deployments are not accessible. To enable access to
preview deployments, setup access control for preview deployment in `pageship.toml`:
```toml
[app.deployments]
access = [
    { ipRange="0.0.0.0/0" }     # preview deployments are accessible to anyone.
]
```
