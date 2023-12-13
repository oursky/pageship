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
var ErrDeploymentExpired = errors.New("deployment expired")

var ErrUndefinedDomain = errors.New("undefined domain")
var ErrDomainNotFound = errors.New("domain not found")
var ErrDomainUsedName = errors.New("used domain name")

var ErrUserNotFound = errors.New("user not found")
var ErrAccessDenied = errors.New("access denied")
var ErrInvalidCredentials = errors.New("invalid credentials")

var ErrCertificateDataNotFound = errors.New("cert data not found")
var ErrCertificateDataLocked = errors.New("cert locked")
