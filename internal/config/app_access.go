package config

import (
	"fmt"
)

type AccessLevel string

const (
	AccessLevelAdmin    AccessLevel = "admin"
	AccessLevelDeployer AccessLevel = "deployer"
	AccessLevelReader   AccessLevel = "reader"

	AccessLevelDefault = AccessLevelReader
)

func (l AccessLevel) IsValid() bool {
	switch l {
	case AccessLevelAdmin, AccessLevelDeployer, AccessLevelReader:
		return true
	default:
		return false
	}
}

func (l AccessLevel) CanAccess(a AccessLevel) bool {
	switch l {
	case AccessLevelAdmin:
		return true
	case AccessLevelDeployer:
		return a == AccessLevelDeployer || a == AccessLevelReader
	case AccessLevelReader:
		return a == AccessLevelReader
	default:
		return false
	}
}

type AccessRule struct {
	CredentialMatcher `mapstructure:",squash"`
	Access            AccessLevel `json:"access" pageship:"omitempty,accessLevel"`
}

func (r *AccessRule) SetDefaults() {
	if r.Access == "" {
		r.Access = AccessLevelDefault
	}
}

type CredentialMatcher struct {
	PageshipUser            string `json:"pageshipUser,omitempty" pageship:"max=100"`
	GitHubUser              string `json:"githubUser,omitempty" pageship:"max=100"`
	GitHubRepositoryActions string `json:"gitHubRepositoryActions,omitempty" pageship:"max=100"`
	IpRange                 string `json:"ipRange,omitempty" pageship:"omitempty,max=100,cidr"`
}

func (c *CredentialMatcher) String() string {
	switch {
	case c.PageshipUser != "":
		return fmt.Sprintf("pageshipUser:%s", c.PageshipUser)
	case c.GitHubUser != "":
		return fmt.Sprintf("githubUser:%s", c.GitHubUser)
	case c.GitHubRepositoryActions != "":
		return fmt.Sprintf("gitHubRepositoryActions:%s", c.GitHubRepositoryActions)
	case c.IpRange != "":
		return fmt.Sprintf("ipRange:%s", c.IpRange)
	}
	return "<unknown>"
}
