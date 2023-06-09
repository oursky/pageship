package models

import "github.com/oursky/pageship/internal/config"

type AppAuthzResult struct {
	CredentialID CredentialID
	Matcher      *config.CredentialMatcher // nil => is owner
}

func (i *AppAuthzResult) MatchedRule() string {
	if i.Matcher == nil {
		return "<owner>"
	}
	return i.Matcher.String()
}

func (a *App) CheckAuthz(level config.AccessLevel, userID string, credentials []CredentialID) (*AppAuthzResult, error) {
	if userID != "" && a.OwnerUserID == userID {
		return &AppAuthzResult{
			CredentialID: CredentialUserID(a.OwnerUserID),
			Matcher:      nil,
		}, nil
	}

	for _, r := range a.Config.Team {
		for _, id := range credentials {
			if id.Matches(&r.CredentialMatcher) && r.Access.CanAccess(level) {
				return &AppAuthzResult{
					CredentialID: id,
					Matcher:      &r.CredentialMatcher,
				}, nil
			}
		}
	}

	return nil, ErrAccessDenied
}

func CheckDeploymentAuthz(access []config.CredentialMatcher, credentials []CredentialID) (*AppAuthzResult, error) {
	for _, m := range access {
		for _, id := range credentials {
			if id.Matches(&m) {
				return &AppAuthzResult{
					CredentialID: id,
					Matcher:      &m,
				}, nil
			}
		}
	}

	return nil, ErrAccessDenied
}
