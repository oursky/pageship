package config

import (
	"fmt"
	"strings"
)

type HostIDScheme string

const (
	HostIDSchemeSubdomain HostIDScheme = "subdomain"
	HostIDSchemeSuffix    HostIDScheme = "suffix"

	HostIDSchemeDefault = HostIDSchemeSubdomain
)

func (s HostIDScheme) IsValid() bool {
	switch s {
	case HostIDSchemeSubdomain, HostIDSchemeSuffix:
		return true
	default:
		return false
	}
}

func (s HostIDScheme) Split(hostID string) (main string, sub string) {
	switch s {
	case HostIDSchemeSubdomain:
		var found bool
		sub, main, found = strings.Cut(hostID, ".")
		if !found {
			main = sub
			sub = ""
		}

	case HostIDSchemeSuffix:
		i := strings.LastIndex(hostID, "--")
		if i == -1 {
			main = hostID
		} else {
			main = hostID[:i]
			sub = strings.ReplaceAll(hostID[i+2:], "-x", "-")
		}

	default:
		panic("hostid: invalid scheme :" + s)
	}
	return
}

func (s HostIDScheme) Make(main string, sub string) string {
	switch s {
	case HostIDSchemeSubdomain:
		if sub == "" {
			return main
		}
		return fmt.Sprintf("%s.%s", sub, main)

	case HostIDSchemeSuffix:
		if sub == "" {
			return main
		}
		return fmt.Sprintf("%s--%s", main, strings.ReplaceAll(sub, "-", "-x"))

	default:
		panic("hostid: invalid scheme :" + s)
	}
}
