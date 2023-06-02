package config

type AccessLevel string

const (
	AccessLevelAdmin    AccessLevel = "admin"
	AccessLevelDeployer AccessLevel = "deployer"
	AccessLevelReader   AccessLevel = "reader"
)
