package models

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
)

var idEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

func newID(kind string) string {
	var data [8]byte
	_, err := rand.Read(data[:])
	if err != nil {
		panic(err)
	}
	r := strings.ToLower(idEncoding.EncodeToString(data[:]))
	return fmt.Sprintf("%s_%s", kind, r)
}
