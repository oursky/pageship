package app

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	Version = "1.2.3"

	var testFiles fstest.MapFS = make(map[string]*fstest.MapFile)
	testFiles["pageship.toml"] = &fstest.MapFile{Data: []byte(`[app]
id = "pageship-test"

team = []

[app.deployments]
# ttl = "24h"
# access = []

[[app.sites]]
name = "main"

# [[app.sites]]
# name = "dev"

# [[app.sites]]
# name = "staging"

[site]
public = "dist"

# access = []
`)}

	s, err := generateContent(testFiles)
	assert.Empty(t, err)
	assert.Equal(t, `FROM ghcr.io/oursky/pageship:v1.2.3
EXPOSE 8000
COPY ./pageship.toml /var/pageship
COPY ./dist /var/pageship/dist

# INSTRUCTIONS:
# 1. install docker (if it is not installed yet)
# 2. open a terminal and navigate to folder containing pageship.toml
# 3. run in terminal:
#      pageship generate dockerfile
# 4. build the image:
#      docker build -t IMAGETAG .
# 5. run the container:
#      docker run -d --name CONTAINERNAME -p PORT:8000 IMAGETAG
# 6. visit in browser (URL):
#      localhost:PORT`, s)
}
