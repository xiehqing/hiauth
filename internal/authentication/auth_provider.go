package authentication

import (
	"context"
	"strings"

	"github.com/xiehqing/hiauth/internal/configx"
	"github.com/xiehqing/hiauth/internal/db/entity"
)

const (
	authProviderLocal = "local"
	authProviderLDAP  = "ldap"
)

type AuthProvider interface {
	Name() string
	Authenticate(ctx context.Context, user *entity.User, password string) error
}

type localAuthProvider struct{}

func (p localAuthProvider) Name() string {
	return authProviderLocal
}

func (p localAuthProvider) Authenticate(ctx context.Context, user *entity.User, password string) error {
	if user == nil || user.Password != md5Password(password) {
		return ErrInvalidLogin
	}
	return nil
}

func (s *Service) authProvider(ctx context.Context) AuthProvider {
	config := configx.New(s.queries)
	provider := strings.ToLower(strings.TrimSpace(config.String(ctx, entity.SecurityAuthProvider, "")))
	if provider == "" && config.Bool(ctx, entity.SecurityAuthLDAPEnabled, false) {
		provider = authProviderLDAP
	}

	switch provider {
	case authProviderLDAP:
		return newLDAPAuthProvider(config)
	default:
		return localAuthProvider{}
	}
}
