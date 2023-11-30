package models_test

import (
	"testing"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
	"github.com/stretchr/testify/assert"
)

func matchRule(rule *config.ACLSubjectRule, id models.CredentialID) bool {
	return id.Matches(rule)
}

func TestGitHubUserCredentials(t *testing.T) {
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubUser: "example"},
		models.CredentialGitHubUser("example"),
	))
	assert.False(t, matchRule(
		&config.ACLSubjectRule{GitHubUser: "example"},
		models.CredentialGitHubUser("foobar"),
	))
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubUser: "example"},
		models.CredentialGitHubUser("Example"),
	))
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubUser: "Example"},
		models.CredentialGitHubUser("example"),
	))
}

func TestGitHubActionsCredentials(t *testing.T) {
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "*"},
		models.CredentialGitHubRepositoryActions("oursky/pageship"),
	))

	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/*"},
		models.CredentialGitHubRepositoryActions("oursky/pageship"),
	))
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/*"},
		models.CredentialGitHubRepositoryActions("Oursky/Pageship"),
	))
	assert.False(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/*"},
		models.CredentialGitHubRepositoryActions("github/example"),
	))

	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/pageship"},
		models.CredentialGitHubRepositoryActions("oursky/pageship"),
	))
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/pageship"},
		models.CredentialGitHubRepositoryActions("Oursky/Pageship"),
	))
	assert.False(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/pageship"},
		models.CredentialGitHubRepositoryActions("Oursky/Example"),
	))
}
