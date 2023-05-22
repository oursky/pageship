package config

import "strings"

const DefaultHostPattern = `*.localhost`

type HostPattern struct {
	Prefix string
	Suffix string
}

func NewHostPattern(pattern string) HostPattern {
	prefix, suffix, ok := strings.Cut(pattern, "*")
	if !ok { // Wildcard not found in pattern; assume suffix
		suffix = prefix
		prefix = ""
	}

	// Trim leading dot in suffix to match for bare domain too
	suffix = strings.TrimPrefix(suffix, ".")

	return HostPattern{Prefix: prefix, Suffix: suffix}
}

func (p HostPattern) MatchString(s string) (string, bool) {
	// Remove trailing port
	if host, _, found := strings.Cut(s, ":"); found {
		s = host
	}

	if !strings.HasPrefix(s, p.Prefix) {
		return "", false
	}
	s = strings.TrimPrefix(s, p.Prefix)

	if !strings.HasSuffix(s, p.Suffix) {
		return "", false
	}
	s = strings.TrimSuffix(s, p.Suffix)

	s = strings.TrimSuffix(s, ".")

	if len(s) == 0 {
		return "", false
	}
	return s, true
}
