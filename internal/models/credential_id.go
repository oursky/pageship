package models

import (
	"strings"

	"github.com/oursky/pageship/internal/config"
)

type CredentialIDKind string

const (
	CredentialIDKindUserID              CredentialIDKind = ""
	CredentialIDKindGitHubUser          CredentialIDKind = "github"
	CredentialIDGitHubRepositoryActions CredentialIDKind = "github-repo-actions"
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
	default:
		return false
	}
}

type CredentialndexKey string

func MakeCredentialIDIndexKeys(id CredentialID) []CredentialndexKey {
	kind, _, found := strings.Cut(string(id), ":")
	if !found {
		kind = ""
	}

	switch CredentialIDKind(kind) {
	case CredentialIDKindUserID,
		CredentialIDKindGitHubUser,
		CredentialIDGitHubRepositoryActions:
		return []CredentialndexKey{CredentialndexKey(id)}

	default:
		return nil
	}
}

func CollectCredentialIDIndexKeys(ids []CredentialID) []CredentialndexKey {
	var keys []CredentialndexKey
	for _, id := range ids {
		keys = append(keys, MakeCredentialIDIndexKeys(id)...)
	}
	return keys
}

func MakeCredentialMatcherIndexKeys(c *config.CredentialMatcher) []CredentialndexKey {
	switch {
	case c.PageshipUser != "":
		return MakeCredentialIDIndexKeys(CredentialUserID(c.PageshipUser))
	case c.GitHubUser != "":
		return MakeCredentialIDIndexKeys(CredentialGitHubUser(c.GitHubUser))
	case c.GitHubRepositoryActions != "":
		return MakeCredentialIDIndexKeys(CredentialGitHubRepositoryActions(c.GitHubRepositoryActions))
	}
	return nil
}
