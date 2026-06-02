package integrate

import (
	"context"

	"github.com/cloudwego/hertz/pkg/route"
	"github.com/xiehqing/hiauth/internal/authentication"
	"github.com/xiehqing/hiauth/internal/routes"
	"gorm.io/gorm"
)

type Integrator struct {
	router *routes.Router
}

func NewIntegrator(db *gorm.DB) *Integrator {
	return &Integrator{
		router: routes.New(db),
	}
}

func (i *Integrator) RefreshTokenManager(ctx context.Context) {
	i.router.RefreshTokenManager(ctx)
}

func (i *Integrator) RegisterRoutes(api *route.RouterGroup) {
	i.router.RegisterRoutes(api)
}

func (i *Integrator) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	response, err := i.router.Authentication().Login(ctx, authentication.LoginRequest{
		Username: req.Username,
		Password: req.Password,
		Device:   req.Device,
	})
	if err != nil {
		return nil, err
	}
	return convertLoginResponse(response), nil
}

func (i *Integrator) Logout(ctx context.Context, token string) error {
	return i.router.Authentication().Logout(ctx, token)
}

func (i *Integrator) CurrentUser(ctx context.Context, token string) (*LoginResponse, error) {
	response, err := i.router.Authentication().CurrentUser(ctx, token)
	if err != nil {
		return nil, err
	}
	return convertLoginResponse(response), nil
}

func (i *Integrator) GetUserByID(ctx context.Context, id int64) (*User, error) {
	user, err := i.router.Authorization().GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertUserPtr(user), nil
}

func (i *Integrator) GetUserByIDs(ctx context.Context, ids []int64) ([]User, error) {
	users, err := i.router.Authorization().GetUsers(ctx, ids)
	if err != nil {
		return nil, err
	}
	return convertUsers(users), nil
}
