# GitHub Actions Integration

Pageship detects if it is running in GitHub Actions environment, and would
authenticate with server automatically if possible.

To deploy from GitHub Actions, first configure the app to accept GitHub Actions
running in the repo as `deployer` permission. See [Access Control](./access-control.md)
guide for details.

After configuring access control, install `pageship` command in workflow and
deploy directly.

```
docker run --rm \
    -e PAGESHIP_API="..." \
    -e ACTIONS_ID_TOKEN_REQUEST_URL="$ACTIONS_ID_TOKEN_REQUEST_URL" \
    -e ACTIONS_ID_TOKEN_REQUEST_TOKEN="$ACTIONS_ID_TOKEN_REQUEST_TOKEN" \
    -v "$PWD:/var/pageship" \
    ghcr.io/oursky/pageship:v0.3.1 \
        deploy /var/pageship --site main -y
```
