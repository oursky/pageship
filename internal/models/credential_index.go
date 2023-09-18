package models

import (
	"encoding/hex"
	"net/netip"
	"strings"

	"github.com/oursky/pageship/internal/config"
)

type CredentialIndexKey string

func MakeCredentialIDIndexKeys(id CredentialID) []CredentialIndexKey {
	kind, data, found := strings.Cut(string(id), ":")
	if !found {
		kind = ""
	}

	switch CredentialIDKind(kind) {
	case CredentialIDKindUserID,
		CredentialIDKindGitHubUser:
		return []CredentialIndexKey{CredentialIndexKey(id)}

	case CredentialIDGitHubRepositoryActions:
		owner, repo, ok := strings.Cut(data, "/")
		if !ok {
			return nil
		}
		prefix := string(CredentialIDGitHubRepositoryActions) + ":"
		return []CredentialIndexKey{
			CredentialIndexKey(prefix + "*"),
			CredentialIndexKey(prefix + owner),
			CredentialIndexKey(prefix + owner + "/" + repo),
		}

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

func MakeCredentialRuleIndexKeys(r *config.ACLSubjectRule) []CredentialIndexKey {
	switch {
	case r.PageshipUser != "":
		return MakeCredentialIDIndexKeys(CredentialUserID(r.PageshipUser))
	case r.GitHubUser != "":
		return MakeCredentialIDIndexKeys(CredentialGitHubUser(r.GitHubUser))
	case r.GitHubRepositoryActions != "":
		prefix := string(CredentialIDGitHubRepositoryActions) + ":"
		if r.GitHubRepositoryActions == "*" {
			return []CredentialIndexKey{CredentialIndexKey(prefix + "*")}
		}

		owner, repo, ok := strings.Cut(r.GitHubRepositoryActions, "/")
		if !ok {
			return nil
		}
		if repo == "*" {
			return []CredentialIndexKey{CredentialIndexKey(prefix + owner)}
		} else {
			return []CredentialIndexKey{CredentialIndexKey(prefix + owner + "/" + repo)}
		}
	case r.IpRange != "":
		cidr, err := netip.ParsePrefix(r.IpRange)
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
	for b := 0; b < len(bytes); b += 2 {
		s.WriteString(hex.EncodeToString(bytes[b : b+2]))
		keys = append(keys, CredentialIndexKey("ip:"+s.String()))
	}
	return keys
}
