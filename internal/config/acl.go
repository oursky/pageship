package config

import (
	"fmt"
	"io"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"
)

type ACL []ACLSubjectRule

func LoadACL(r io.Reader) (ACL, error) {
	var m map[string]any
	if err := toml.NewDecoder(r).Decode(&m); err != nil {
		return nil, err
	}

	var aclFile struct {
		Access ACL `json:"access"`
	}
	if err := mapstructure.Decode(m, &aclFile); err != nil {
		return nil, err
	}

	if err := validate.Struct(aclFile); err != nil {
		return nil, err
	}

	return aclFile.Access, nil
}

type ACLSubjectRule struct {
	PageshipUser            string `json:"pageshipUser,omitempty" pageship:"max=100"`
	GitHubUser              string `json:"githubUser,omitempty" pageship:"max=100"`
	GitHubRepositoryActions string `json:"gitHubRepositoryActions,omitempty" pageship:"max=100"`
	IpRange                 string `json:"ipRange,omitempty" pageship:"omitempty,max=100,cidr"`
}

func (c *ACLSubjectRule) String() string {
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
