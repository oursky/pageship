# Automatic TLS

Pageship supports automatic TLS through [certmagic](https://github.com/caddyserver/certmagic) library.

Automatic TLS can be activated by passing `--tls` command line parameter.
Certificates would be obtained from Let's Encrypt when a site domain is accessed
for the first time. It is recommnded to provide a email to receive notifications
from certificate issuer using `--tls-acme-email` command line parameter.

## Certificate Persistence

In single-site & unmanaged-sites mode, certificate data is stored on the default
filesystem directory specified by `certmagic` library
(`${XDG_DATA_HOME}/certmagic`) in plain-text. Care should be taken to secure
the key materials.

In managed sites mode, certificate data is stored in database. Optionally, an
encryption key can be specified through `--tls-protect-key` parameter to
encrypt the certificate data at rest using NaCL secretbox.
