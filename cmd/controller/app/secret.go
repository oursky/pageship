package app

import (
	"crypto/rand"
	"encoding/base64"
)

func generateSecret() string {
	data := make([]byte, 8)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
