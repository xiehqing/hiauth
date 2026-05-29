package routes

import (
	"context"
	"github.com/xiehqing/hiauth/internal/authorization"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/xiehqing/hiauth/internal/db/queries"
)

func (r *Router) registerRoleRoutes(api *route.RouterGroup) {
	roles := api.Group("/roles", checkLogin(), r.tenantContext())
	roles.GET("", r.listRoles)
	roles.GET("/options", r.listRoleOptions)
	roles.POST("", r.createRole)
	roles.GET("/:id", r.getRole)
	roles.PUT("/:id", r.updateRole)
	roles.DELETE("/:id", r.deleteRole)
	roles.GET("/:id/menus", r.getRoleMenus)
	roles.PUT("/:id/menus", r.updateRoleMenus)
}

func (r *Router) createRole(ctx context.Context, c *app.RequestContext) {
	tenantID := tenantIDFromContext(ctx)
	var req authorization.CreateRoleRequest
	if !bindJSON(c, &req) {
		return
	}
	if tenantID > 0 {
		req.TenantID = tenantID
	}
	data, err := r.service.CreateRole(ctx, req)
	handleData(c, data, err)
}

func (r *Router) updateRole(ctx context.Context, c *app.RequestContext) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	tenantID := tenantIDFromContext(ctx)

	var req authorization.UpdateRoleRequest
	if !bindJSON(c, &req) {
		return
	}
	req.ID = id
	if tenantID > 0 {
		req.TenantID = tenantID
	}
	data, err := r.service.UpdateRole(ctx, req)
	handleData(c, data, err)
}

func (r *Router) getRole(ctx context.Context, c *app.RequestContext) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	data, err := r.service.GetRole(ctx, id)
	handleData(c, data, err)
}

func (r *Router) listRoles(ctx context.Context, c *app.RequestContext) {
	tenantID := tenantIDFromContext(ctx)
	req := authorization.ListRolesRequest{
		RoleListFilter: queries.RoleListFilter{
			Pagination: pagination(c),
			TenantID:   tenantID,
			BuiltIn:    queryIntPtr(c, "builtIn"),
		},
	}
	data, err := r.service.ListRoles(ctx, req)
	handleData(c, data, err)
}

func (r *Router) listRoleOptions(ctx context.Context, c *app.RequestContext) {
	tenantID := tenantIDFromContext(ctx)
	data, err := r.service.ListAllRoles(ctx, tenantID)
	handleData(c, data, err)
}

func (r *Router) updateRoleMenus(ctx context.Context, c *app.RequestContext) {
	id, ok := pathID(c)
	if !ok {
		return
	}

	var req authorization.UpdateRoleMenusRequest
	if !bindJSON(c, &req) {
		return
	}
	req.RoleID = id
	data, err := r.service.UpdateRoleMenus(ctx, req)
	handleData(c, data, err)
}

func (r *Router) getRoleMenus(ctx context.Context, c *app.RequestContext) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	data, err := r.service.GetRoleMenus(ctx, id)
	handleData(c, data, err)
}

func (r *Router) deleteRole(ctx context.Context, c *app.RequestContext) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	handleMsg(c, "角色删除成功", r.service.DeleteRole(ctx, id))
}
