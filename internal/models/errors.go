package models

import "errors"

var ErrUsedAppID = errors.New("used app ID")
var ErrAppNotFound = errors.New("app not found")

var ErrUndefinedSite = errors.New("undefined site")

var ErrDeploymentNotFound = errors.New("deployment not found")
var ErrDeploymentInvalidStatus = errors.New("deployment has invalid status")
