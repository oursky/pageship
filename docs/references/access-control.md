# Access Control

Access control is configured by ACL rules of different types. A request/action
passes the access control check if it matches any of the applicable ACL rules.

A typical ACL would looks like this:
```toml
access = [
    { githubUser="username" },
    { ipRange="127.0.0.1/32" }
]
```

## Authentication

A GitHub user may authenticate through the `pageship login` command. Currently,
it will connect to the Pageship server through SSH protocol, and verify user's
identity through GitHub user's public key.

GitHub Actions jobs would be authenticate automatically when `pageship` command
detected running in CI environment. It authenticates through GitHub Actions
OIDC token.

## ACL Types

### GitHub user

```toml
{ githubUser = "username" }
```

Actions/requests from the specified GitHub user is allowed.

### GitHub Actions repository
```toml
{ gitHubRepositoryActions = "oursky/pageship" }
{ gitHubRepositoryActions = "oursky/*" }
{ gitHubRepositoryActions = "*" }
```

Actions/requests from the specified GitHub Action jobs is allowed. Wildcard can
be specified for all repository of a user/organization, or any repository.


### IP Range

```toml
{ ipRange = "127.0.0.1/32" }
{ ipRange = "192.168.0.0/16" }
{ ipRange = "0.0.0.0/0" }
{ ipRange = "::1/128" }
```

Actions/requests from the specified IP range (CIDR) is allowed.
IPv4 is mapped to IPv6 before matching.
