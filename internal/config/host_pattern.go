package config

import "strings"

const DefaultHostPattern = `http://*.localhost:8001`

type HostPattern struct {
	Prefix        string
	Suffix        string
	LeadingScheme string
	TrailingPort  string
}

func NewHostPattern(pattern string) *HostPattern {
	// Leading scheme
	leadingScheme := ""
	if scheme, remaining, found := strings.Cut(pattern, "://"); found {
		leadingScheme = scheme + "://"
		pattern = remaining
	} else {
		leadingScheme = "https://"
	}

	prefix, suffix, ok := strings.Cut(pattern, "*")
	if !ok { // Wildcard not found in pattern; assume suffix
		suffix = prefix
		prefix = ""
	}

	trailingPort := ""
	// Trailing port
	if host, port, found := strings.Cut(suffix, ":"); found {
		trailingPort = ":" + port
		suffix = host
	}

	// Trim leading dot in suffix to match for bare domain too
	suffix = strings.TrimPrefix(suffix, ".")

	return &HostPattern{
		Prefix:        prefix,
		Suffix:        suffix,
		LeadingScheme: leadingScheme,
		TrailingPort:  trailingPort,
	}
}

func (p *HostPattern) MatchString(s string) (string, bool) {
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

	return s, true
}
func (p *HostPattern) MakeURL(value string) string {
	url := p.LeadingScheme + p.Prefix + value

	if p.Suffix != "" {
		if value != "" {
			url += "."
		}
		url += p.Suffix
	}

	url += p.TrailingPort

	return url
}
