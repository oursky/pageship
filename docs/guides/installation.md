# Installation

## Binary Release

Download latest binary release from [GitHub](https://github.com/oursky/pageship/releases).

```
curl -sSL https://raw.githubusercontent.com/oursky/pageship/main/install.sh | sh -s -- -b .
sudo mv ./pageship /usr/local/bin
```

## Runner with Docker

Docker images are available from GitHub Packages:

```sh
docker pull ghcr.io/oursky/pageship:v0.4.0
docker pull ghcr.io/oursky/pageship-controller:v0.4.0
```

## Go install

```sh
go install github.com/oursky/pageship@v0.4.0
```
