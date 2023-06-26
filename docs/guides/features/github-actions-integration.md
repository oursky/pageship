# GitHub Actions Integration

Pageship detects if it is running in GitHub Actions environment, and would
authenticate with server automatically if possible.

To deploy from GitHub Actions, first configure the app to accept GitHub Actions
running in the repo as `deployer` permission. See [Access Control](./access-control.md)
guide for details.

Then, enable OIDC token in GitHub Actions workflow by granting `id-token`
permission to workflow jobs:
```yaml
jobs:
  <job-name>:
    permissions:
      contents: read
      id-token: write
```

Finally, install `pageship` command in workflow and deploy directly.

```
docker run --rm \
    -e PAGESHIP_API="..." \
    -e ACTIONS_ID_TOKEN_REQUEST_URL="$ACTIONS_ID_TOKEN_REQUEST_URL" \
    -e ACTIONS_ID_TOKEN_REQUEST_TOKEN="$ACTIONS_ID_TOKEN_REQUEST_TOKEN" \
    -v "$PWD:/var/pageship" \
    ghcr.io/oursky/pageship:v0.3.1 \
        deploy /var/pageship --site main -y
```
