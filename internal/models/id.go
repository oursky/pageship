package models

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
)

var idEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

func newID(kind string) string {
	return fmt.Sprintf("%s_%s", kind, RandomID(8))
}

func RandomID(n int) string {
	data := make([]byte, n)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return strings.ToLower(idEncoding.EncodeToString(data[:]))
}
