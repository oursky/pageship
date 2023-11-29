package site

import (
	"context"
	"errors"
)

var ErrSiteNotFound = errors.New("site not found")

type Resolver interface {
	Kind() string
	// IsWildcard indicates whether this resolver would always resolve successfully.
	IsWildcard() bool
	Resolve(ctx context.Context, matchedID string) (*Descriptor, error)
}

func CheckDefaultSite(siteName *string, defaultSite string) bool {
	if *siteName == defaultSite {
		// Default site must be accessed through empty ID
		return false
	}

	if *siteName == "" {
		if defaultSite == "" || defaultSite == "-" { // Allow use `-` to disable default site
			// Default site is disabled; treat as not found
			return false
		}
		*siteName = defaultSite
	}

	return true
}
