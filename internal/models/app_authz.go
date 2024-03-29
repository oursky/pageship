package models

import "github.com/oursky/pageship/internal/config"

type AppAuthzResult struct {
	CredentialID CredentialID
	Rule         *config.ACLSubjectRule // nil => is owner
}

func (i *AppAuthzResult) MatchedRule() string {
	if i.Rule == nil {
		return "<owner>"
	}
	return i.Rule.String()
}

func (a *App) CheckAuthz(level config.AccessLevel, userID string, credentials []CredentialID) (*AppAuthzResult, error) {
	if userID != "" && a.OwnerUserID == userID {
		return &AppAuthzResult{
			CredentialID: CredentialUserID(a.OwnerUserID),
			Rule:         nil,
		}, nil
	}

	for _, r := range a.Config.Team {
		for _, id := range credentials {
			if id.Matches(&r.ACLSubjectRule) && r.Access.CanAccess(level) {
				return &AppAuthzResult{
					CredentialID: id,
					Rule:         &r.ACLSubjectRule,
				}, nil
			}
		}
	}

	return nil, ErrAccessDenied
}

func CheckACLAuthz(access config.ACL, credentials []CredentialID) (*AppAuthzResult, error) {
	for _, r := range access {
		for _, id := range credentials {
			if id.Matches(&r) {
				return &AppAuthzResult{
					CredentialID: id,
					Rule:         &r,
				}, nil
			}
		}
	}

	return nil, ErrAccessDenied
}
