package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	Version = "1.2.3"
	s, err := generateContent()
	assert.Empty(t, err)
	assert.Equal(t, `FROM ghcr.io/oursky/pageship:v1.2.3
EXPOSE 8000
COPY . /var/pageship

# INSTRUCTIONS:
# 1. install docker (if it is not installed yet)
# 2. open a terminal and navigate to your static page folder
# 3. run in terminal:
#      pageship generate dockerfile
# 4. build the image:
#      docker build -t IMAGETAG .
# 5. run the container:
#      docker run -d --name CONTAINERNAME -p PORT:8000 IMAGETAG
# 6. visit in browser (URL):
#      localhost:PORT`, s)
}
