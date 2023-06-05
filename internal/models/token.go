package models

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	Name        string         `json:"name,omitempty"`
	Credentials []CredentialID `json:"credentials,omitempty"`
	jwt.RegisteredClaims
}

func NewTokenClaims(sub TokenSubject, name string) *TokenClaims {
	return &TokenClaims{
		Name: name,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: string(sub),
		},
	}
}

type TokenSubjectKind string

const (
	TokenSubjectKindUser          TokenSubjectKind = ""
	TokenSubjectKindGitHubActions TokenSubjectKind = "github-actions"
)

func (k TokenSubjectKind) IsValid() bool {
	switch k {
	case TokenSubjectKindUser, TokenSubjectKindGitHubActions:
		return true
	}
	return false
}

type TokenSubject string

func TokenSubjectUser(id string) TokenSubject {
	return TokenSubject(id)
}

func TokenSubjectGitHubActions(jti string) TokenSubject {
	return TokenSubject(fmt.Sprintf("%s:%s", TokenSubjectKindGitHubActions, jti))
}

func (s TokenSubject) Parse() (TokenSubjectKind, string, bool) {
	k, data, ok := strings.Cut(string(s), ":")
	if !ok {
		data = k
		k = ""
	}

	kind := TokenSubjectKind(k)
	if !kind.IsValid() {
		return "", "", false
	}

	return kind, data, true
}
