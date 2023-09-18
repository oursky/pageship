package models_test

import (
	"net/netip"
	"testing"

	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func matchRule(rule *config.ACLSubjectRule, id models.CredentialID) bool {
	index := make(map[string]struct{})
	for _, key := range models.MakeCredentialRuleIndexKeys(rule) {
		index[string(key)] = struct{}{}
	}

	for _, key := range models.MakeCredentialIDIndexKeys(id) {
		if _, ok := index[string(key)]; ok {
			return true
		}
	}
	return false
}

type GitHubActionsCredentialsTestSuite struct {
	suite.Suite
}

func TestGitHubActionsCredentials(t *testing.T) {
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/pageship"},
		models.CredentialGitHubRepositoryActions("oursky/pageship"),
	))
	assert.False(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/pageship"},
		models.CredentialGitHubRepositoryActions("oursky/other"),
	))
	assert.False(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/pageship"},
		models.CredentialGitHubRepositoryActions("other/pageship"),
	))

	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/*"},
		models.CredentialGitHubRepositoryActions("oursky/pageship"),
	))
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/*"},
		models.CredentialGitHubRepositoryActions("oursky/other"),
	))
	assert.False(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "oursky/*"},
		models.CredentialGitHubRepositoryActions("other/oursky"),
	))

	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "*"},
		models.CredentialGitHubRepositoryActions("oursky/pageship"),
	))
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "*"},
		models.CredentialGitHubRepositoryActions("oursky/other"),
	))
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubRepositoryActions: "*"},
		models.CredentialGitHubRepositoryActions("other/oursky"),
	))
}

func TestGitHubUserCredentials(t *testing.T) {
	assert.True(t, matchRule(
		&config.ACLSubjectRule{GitHubUser: "oursky"},
		models.CredentialGitHubUser("oursky"),
	))
	assert.False(t, matchRule(
		&config.ACLSubjectRule{GitHubUser: "oursky"},
		models.CredentialGitHubUser("other"),
	))
}

func TestIPCredentials(t *testing.T) {
	assert.True(t, matchRule(
		&config.ACLSubjectRule{IpRange: "127.0.0.1/32"},
		models.CredentialIP("127.0.0.1"),
	))
	assert.False(t, matchRule(
		&config.ACLSubjectRule{IpRange: "127.0.0.1/32"},
		models.CredentialIP("127.0.0.2"),
	))

	assert.True(t, matchRule(
		&config.ACLSubjectRule{IpRange: "127.0.0.1/16"},
		models.CredentialIP("127.0.0.1"),
	))
	assert.True(t, matchRule(
		&config.ACLSubjectRule{IpRange: "127.0.0.1/16"},
		models.CredentialIP("127.0.100.2"),
	))

	assert.True(t, matchRule(
		&config.ACLSubjectRule{IpRange: "2001:db8::/48"},
		models.CredentialIP("2001:db8::1"),
	))
	assert.False(t, matchRule(
		&config.ACLSubjectRule{IpRange: "2001:db8::/48"},
		models.CredentialIP("::1"),
	))

	assert.True(t, matchRule(
		&config.ACLSubjectRule{IpRange: "::/0"},
		models.CredentialIP("192.168.0.1"),
	))
	assert.True(t, matchRule(
		&config.ACLSubjectRule{IpRange: "::ffff:192.168.0.1/120"},
		models.CredentialIP("192.168.0.255"),
	))
}

func FuzzIPCredentials(f *testing.F) {
	add := func(cidr string, ip string) {
		prefix := netip.MustParsePrefix(cidr)
		f.Add(prefix.Addr().AsSlice(), prefix.Bits(), netip.MustParseAddr(ip).AsSlice())
	}
	add("0.0.0.0/0", "10.0.0.1")
	add("192.168.0.1/24", "8.8.8.8")
	add("127.0.0.1/32", "127.0.0.2")
	add("::0/0", "192.168.0.1")
	add("2001:db8::/48", "2001:db8::1")
	add("::1/128", "::1")

	f.Fuzz(func(t *testing.T, ipPrefix []byte, bits int, ip []byte) {
		ruleAddr, ok := netip.AddrFromSlice(ipPrefix)
		if !ok {
			return
		}
		ruleCIDR, err := ruleAddr.Prefix(bits)
		if err != nil {
			return
		}
		credIP, ok := netip.AddrFromSlice(ip)
		if !ok {
			return
		}

		if !ruleCIDR.Contains(credIP) {
			return
		}

		rule := &config.ACLSubjectRule{IpRange: ruleCIDR.String()}
		cred := models.CredentialIP(credIP.String())

		assert.True(t, matchRule(
			&config.ACLSubjectRule{IpRange: ruleCIDR.String()},
			models.CredentialIP(credIP.String()),
		), "range=%s;ip=%s;range_keys=%+v;cred_keys=%+v",
			ruleCIDR.String(), credIP.String(),
			models.MakeCredentialRuleIndexKeys(rule), models.MakeCredentialIDIndexKeys(cred))
	})
}
