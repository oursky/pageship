package config

import (
	"bufio"
	"io"
	"strings"
)

type Allowlist[T ~string] map[T]struct{}

func LoadAllowlist[T ~string](r io.Reader) (Allowlist[T], error) {
	scn := bufio.NewScanner(r)
	list := make(map[T]struct{})
	for scn.Scan() {
		line := scn.Text()

		if i := strings.Index(line, "#"); i != -1 {
			// Trim comments.
			line = line[:i]
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		list[T(line)] = struct{}{}
	}
	return list, scn.Err()
}

func (l Allowlist[T]) IsAllowed(values ...T) bool {
	for _, v := range values {
		if _, ok := l[v]; ok {
			return true
		}
	}
	return false
}
