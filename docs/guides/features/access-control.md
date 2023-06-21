# Access Control

## Users & Credentials

In pageship, users are mostly just ID for internal reference. For purpose
of access control, each user/request is associated with a set of credentials, and
credentials are matched with ACL for permission check.

Pageship currently recognizes following credentials:

### GitHub user

The user/request is associated with a specific GitHub user.

Users can be authenticated as GitHub user through SSH login with
`pageship login` command:
```sh
$ pageship login
GitHub user name: *****
SSH key file: *****
Logged in as *****.
$
```

### [Github repository actions](./github-actions-integration.md)

The user/request is originated from GitHub Actions running in a specific
GitHub repository.

`pageship` command would authenticate as GitHub repository actions automatically
when it detected running in GitHub Actions environment.

### IP address

The user/request is originated from a specific IP address. All users/requests
are automatically associated with an IP address credential.

## Site Access

Site access can be specified through ACL in the `access` field:
```toml
[site]
access = [
    { ipRange="127.0.0.1/32" }
]
```

## App Management Access

App management access can be specified through ACL in the `team` field:
```toml
[app]
team = [
    { githubUser="...", access="admin" },
    { gitHubRepositoryActions="oursky/pageship", access="deployer" }
]
```

There are three levels of access for management:
- `reader`: read-only access to app metadata (e.g. list of deployments/sites)
- `deployer`: access neccessary for deploying sites
- `admin`: full access to the app

In addition, the creator user of an app is considered as the owner of the app,
and always has full access to the app.

## API Access Control

The server API may be protected from unwanted access by specifying an ACL file
in `PAGESHIP_API_ACL` environment variable.

```toml
[[access]]
githubUser="..."

[[access]]
gitHubRepositoryActions="oursky/pageship"

```
