package routes

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/xiehqing/hiauth/internal/authentication"
	"github.com/xiehqing/hiauth/internal/authorization"
	"github.com/xiehqing/hiauth/internal/db/queries"
)

func (r *Router) registerTenantRoutes(api *route.RouterGroup) {
	tenants := api.Group("/tenants", checkLogin())
	tenants.GET("", r.listTenants)
	tenants.POST("", r.tenantCreateContext(), r.createTenant)
	tenants.GET("/:id/users", r.pathTenantManagerContext(), r.listTenantUsers)
	tenants.POST("/:id/users", r.pathTenantManagerContext(), r.addTenantUsers)
	tenants.POST("/:id/invitations", r.pathTenantManagerContext(), r.inviteTenantUser)
	tenants.GET("/:id", r.pathTenantManagerContext(), r.getTenant)
	tenants.PUT("/:id", r.pathTenantManagerContext(), r.updateTenant)
	tenants.DELETE("/:id", r.pathTenantManagerContext(), r.deleteTenant)
}

func (r *Router) createTenant(ctx context.Context, c *app.RequestContext) {
	var req authorization.CreateTenantRequest
	if !bindJSON(c, &req) {
		return
	}
	if user, ok := r.currentLoginUser(ctx, c); ok {
		if req.OwnerUserID <= 0 || !isPlatformAdmin(user) {
			req.OwnerUserID = user.ID
		}
		if req.Operator == "" {
			req.Operator = user.Username
		}
	}
	data, err := r.service.CreateTenant(ctx, req)
	handleData(c, data, err)
}

func (r *Router) updateTenant(ctx context.Context, c *app.RequestContext) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req authorization.UpdateTenantRequest
	if !bindJSON(c, &req) {
		return
	}
	req.ID = id
	if user, ok := r.currentLoginUser(ctx, c); ok {
		if !isPlatformAdmin(user) {
			req.OwnerUserID = 0
		}
		if req.Operator == "" {
			req.Operator = user.Username
		}
	}
	data, err := r.service.UpdateTenant(ctx, req)
	handleData(c, data, err)
}

func (r *Router) getTenant(ctx context.Context, c *app.RequestContext) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	data, err := r.service.GetTenant(ctx, id)
	handleData(c, data, err)
}

func (r *Router) listTenants(ctx context.Context, c *app.RequestContext) {
	user, ok := r.currentLoginUser(ctx, c)
	if !ok {
		handleError(c, authentication.ErrNoPermission)
		return
	}
	req := authorization.ListTenantsRequest{
		TenantListFilter: queries.TenantListFilter{
			Pagination: pagination(c),
			Status:     queryIntPtr(c, "status"),
		},
	}
	if !isPlatformAdmin(user) {
		req.UserID = user.ID
	}
	data, err := r.service.ListTenants(ctx, req)
	handleData(c, data, err)
}

func (r *Router) listTenantUsers(ctx context.Context, c *app.RequestContext) {
	data, err := r.service.ListTenantUsers(ctx, tenantIDFromContext(ctx))
	handleData(c, data, err)
}

func (r *Router) addTenantUsers(ctx context.Context, c *app.RequestContext) {
	var req authorization.AddTenantUsersRequest
	if !bindJSON(c, &req) {
		return
	}
	req.TenantID = tenantIDFromContext(ctx)
	data, err := r.service.AddTenantUsers(ctx, req)
	handleData(c, data, err)
}

func (r *Router) inviteTenantUser(ctx context.Context, c *app.RequestContext) {
	var req authorization.InviteTenantUserRequest
	if !bindJSON(c, &req) {
		return
	}
	req.TenantID = tenantIDFromContext(ctx)
	data, err := r.service.InviteTenantUser(ctx, req)
	handleData(c, data, err)
}

func (r *Router) deleteTenant(ctx context.Context, c *app.RequestContext) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	handleMsg(c, "租户删除成功", r.service.DeleteTenant(ctx, id))
}
