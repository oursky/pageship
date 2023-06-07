package config

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
	AccessSubject `mapstructure:",squash"`
	Access        AccessLevel `json:"access" pageship:"omitempty,accessLevel"`
}

func (r *AccessRule) SetDefaults() {
	if r.Access == "" {
		r.Access = AccessLevelDefault
	}
}

type AccessSubject struct {
	PageshipUser            string `json:"pageshipUser,omitempty" pageship:"max=100"`
	GitHubUser              string `json:"githubUser,omitempty" pageship:"max=100"`
	GitHubRepositoryActions string `json:"gitHubRepositoryActions,omitempty" pageship:"max=100"`
}
