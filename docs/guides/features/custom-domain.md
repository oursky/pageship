# Custom Domains

Pageship supports custom domains for serving pages for an app. We assume a
cooperative model for custom domain association, so domain ownership
verification is not required.

To enable custom domain, configure `pageship.toml` and specify the site to serve
from the domain:

```toml
# 2 sites for the app: 'main' & 'dev'
[[app.sites]]
name = "main"

[[app.sites]]
name = "dev"

# For 'main' site, serve it at 'example.com'. Traffic to default domain is
# redirected to the configured domain automatically.
[[app.domains]]
domain="example.com"
site="main"
```

If the domain name is already in-use by other apps, the custom domain would not
be activated automatically when first added to the configuration. It can be
activated/deactivated manually using `pageship domains activate <domain name>`/
`pageship domains deactivate <domain name>` command.

Custom domains of the app can be listed with `pageship domains` command.
Additional setup instruction (e.g. DNS setup) would be shown if provided by
server operator.
