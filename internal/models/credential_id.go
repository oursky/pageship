package models

import (
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

func (c CredentialID) Matches(r *config.ACLSubjectRule) bool {
	kind, data, found := strings.Cut(string(c), ":")
	if !found {
		data = kind
		kind = ""
	}

	switch CredentialIDKind(kind) {
	case CredentialIDKindUserID:
		return r.PageshipUser != "" && r.PageshipUser == data
	case CredentialIDKindGitHubUser:
		return r.GitHubUser != "" && r.GitHubUser == data
	case CredentialIDGitHubRepositoryActions:
		if r.GitHubRepositoryActions == "*" || r.GitHubRepositoryActions == data {
			return true
		}

		repoOwner, _, ok := strings.Cut(data, "/")
		return ok && r.GitHubRepositoryActions == repoOwner+"/*"
	case CredentialIDIP:
		if r.IpRange == "" {
			return false
		}

		cidr, err := netip.ParsePrefix(r.IpRange)
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
