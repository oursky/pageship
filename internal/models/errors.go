package models

import "errors"

var ErrAppUsedID = errors.New("used app ID")
var ErrAppNotFound = errors.New("app not found")

var ErrUndefinedSite = errors.New("undefined site")
var ErrSiteNotFound = errors.New("site not found")

var ErrDeploymentNotFound = errors.New("deployment not found")
var ErrDeploymentUsedName = errors.New("used deployment name")
var ErrDeploymentNotUploaded = errors.New("deployment is not uploaded")
var ErrDeploymentAlreadyUploaded = errors.New("deployment is already uploaded")

var ErrUserNotFound = errors.New("user not found")
var ErrDeleteCurrentUser = errors.New("cannot delete current user")
var ErrInvalidCredentials = errors.New("invalid credentials")
