package models

import (
	"encoding/hex"
	"net/netip"
	"strings"

	"github.com/oursky/pageship/internal/config"
)

type CredentialIDKind string

const (
	CredentialIDKindUserID              CredentialIDKind = ""
	CredentialIDKindGitHubUser          CredentialIDKind = "github"
	CredentialIDGitHubRepositoryActions CredentialIDKind = "github-repo-actions"
	CredentialIDIP                      CredentialIDKind = "ip"
)

type CredentialID string

func CredentialUserID(id string) CredentialID {
	return CredentialID(id)
}

func CredentialGitHubUser(username string) CredentialID {
	return CredentialID(string(CredentialIDKindGitHubUser) + ":" + username)
}

func CredentialGitHubRepositoryActions(repo string) CredentialID {
	return CredentialID(string(CredentialIDGitHubRepositoryActions) + ":" + repo)
}
func CredentialIP(ip string) CredentialID {
	return CredentialID(string(CredentialIDIP) + ":" + ip)
}

func (c CredentialID) Matches(m *config.CredentialMatcher) bool {
	kind, data, found := strings.Cut(string(c), ":")
	if !found {
		data = kind
		kind = ""
	}

	switch CredentialIDKind(kind) {
	case CredentialIDKindUserID:
		return m.PageshipUser != "" && m.PageshipUser == data
	case CredentialIDKindGitHubUser:
		return m.GitHubUser != "" && m.GitHubUser == data
	case CredentialIDGitHubRepositoryActions:
		return m.GitHubRepositoryActions != "" && m.GitHubRepositoryActions == data
	case CredentialIDIP:
		if m.IpRange == "" {
			return false
		}

		cidr, err := netip.ParsePrefix(m.IpRange)
		if err != nil {
			return false
		}

		addr, err := netip.ParseAddr(data)
		if err != nil {
			return false
		}

		addr = addr.WithZone("")
		if cidr.Addr().Is6() {
			addr = netip.AddrFrom16(addr.As16())
		}

		return cidr.Contains(addr)

	default:
		return false
	}
}

type CredentialIndexKey string

func MakeCredentialIDIndexKeys(id CredentialID) []CredentialIndexKey {
	kind, data, found := strings.Cut(string(id), ":")
	if !found {
		kind = ""
	}

	switch CredentialIDKind(kind) {
	case CredentialIDKindUserID,
		CredentialIDKindGitHubUser,
		CredentialIDGitHubRepositoryActions:
		return []CredentialIndexKey{CredentialIndexKey(id)}

	case CredentialIDIP:
		addr, err := netip.ParseAddr(data)
		if err != nil {
			return nil
		}

		addr = netip.AddrFrom16(addr.As16())
		return makeIPKeys(addr, addr.BitLen())

	default:
		return nil
	}
}

func CollectCredentialIDIndexKeys(ids []CredentialID) []CredentialIndexKey {
	var keys []CredentialIndexKey
	for _, id := range ids {
		keys = append(keys, MakeCredentialIDIndexKeys(id)...)
	}
	return keys
}

func MakeCredentialMatcherIndexKeys(c *config.CredentialMatcher) []CredentialIndexKey {
	switch {
	case c.PageshipUser != "":
		return MakeCredentialIDIndexKeys(CredentialUserID(c.PageshipUser))
	case c.GitHubUser != "":
		return MakeCredentialIDIndexKeys(CredentialGitHubUser(c.GitHubUser))
	case c.GitHubRepositoryActions != "":
		return MakeCredentialIDIndexKeys(CredentialGitHubRepositoryActions(c.GitHubRepositoryActions))
	case c.IpRange != "":
		cidr, err := netip.ParsePrefix(c.IpRange)
		if err != nil {
			return nil
		}

		bits := cidr.Bits()
		addr := netip.AddrFrom16(cidr.Addr().As16())
		if cidr.Addr().Is4() {
			// Map ipv4 CIDR to ipv6.
			bits += 96
		}

		keys := makeIPKeys(addr, bits)
		return []CredentialIndexKey{keys[len(keys)-1]} // Use longest key (i.e. last key)
	}
	return nil
}

func makeIPKeys(addr netip.Addr, bits int) []CredentialIndexKey {
	addrBytes := addr.As16()
	bytes := addrBytes[:(bits/8)&(^1)]

	keys := []CredentialIndexKey{"ip:"}
	var s strings.Builder
	zeroes := 0
	for b := 0; b < len(bytes); b += 2 {
		if bytes[b] == 0 && bytes[b+1] == 0 {
			zeroes++
			continue
		}

		if zeroes != 0 {
			s.WriteByte(':')
			zeroes = 0
		}
		if b != 0 {
			s.WriteByte(':')
		}
		s.WriteString(hex.EncodeToString(bytes[b : b+2]))
		keys = append(keys, CredentialIndexKey("ip:"+s.String()))
	}
	if zeroes > 0 {
		keys = append(keys, CredentialIndexKey("ip:"+s.String()+":"))
	}
	return keys
}