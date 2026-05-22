package integrate

import (
	"context"
	"github.com/cloudwego/hertz/pkg/route"
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
