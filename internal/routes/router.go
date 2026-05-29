package routes

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/xiehqing/hiauth/internal/authentication"
	"github.com/xiehqing/hiauth/internal/authorization"
	"github.com/xiehqing/hiauth/internal/configx"
	"github.com/xiehqing/hiauth/internal/db/entity"
	"github.com/xiehqing/hiauth/internal/db/queries"
	"github.com/xiehqing/hiauth/internal/hitokenx"
	"github.com/xiehqing/hitoken/htputil"
	"github.com/xiehqing/infra/pkg/hertzx"
	"github.com/xiehqing/infra/pkg/ormx"
	"gorm.io/gorm"
)

type tenantContextKey struct{}

type Router struct {
	q              *queries.Queries
	service        *authorization.Service
	authentication *authentication.Service
}

func New(db *gorm.DB) *Router {
	q := queries.New(db)
	q.AutoMigrate(context.Background())
	return &Router{
		q:              q,
		service:        authorization.New(q),
		authentication: authentication.New(q),
	}
}

func (r *Router) RefreshTokenManager(ctx context.Context) {
	hitokenx.RefreshManager(ctx, r.q)
}

func (r *Router) Init(server *server.Hertz) {
	api := server.Group("/api/v1")
	api.Use(r.auditContext())
	api.GET("/health", health)

	r.registerAuthenticationRoutes(api)
	r.registerTenantRoutes(api)
	r.registerUserRoutes(api)
	r.registerRoleRoutes(api)
	r.registerDepartmentRoutes(api)
	r.registerMenuRoutes(api)
	r.registerSystemConfigRoutes(api)
	r.registerAuditRoutes(api)
}

func (r *Router) RegisterRoutes(api *route.RouterGroup) {
	api.Use(r.auditContext())
	api.GET("/health", health)
	r.registerAuthenticationRoutes(api)
	r.registerTenantRoutes(api)
	r.registerUserRoutes(api)
	r.registerRoleRoutes(api)
	r.registerDepartmentRoutes(api)
	r.registerMenuRoutes(api)
	r.registerSystemConfigRoutes(api)
	r.registerAuditRoutes(api)
}

func bindJSON(c *app.RequestContext, req any) bool {
	if err := c.BindJSON(req); err != nil {
		hertzx.Badf(c, "请求体解析失败: %v", err)
		return false
	}
	return true
}

func pathID(c *app.RequestContext) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		hertzx.Badf(c, "参数 id 不合法")
		return 0, false
	}
	return id, true
}

func handleData(c *app.RequestContext, data any, err error) {
	if err != nil {
		handleError(c, err)
		return
	}
	hertzx.Data(c, formatResponseData(data))
}

func handleMsg(c *app.RequestContext, message string, err error) {
	if err != nil {
		handleError(c, err)
		return
	}
	hertzx.Msg(c, message)
}

func handleError(c *app.RequestContext, err error) {
	switch {
	case errors.Is(err, authorization.ErrInvalidID),
		errors.Is(err, authorization.ErrInvalidArgument),
		errors.Is(err, authorization.ErrNotFound),
		errors.Is(err, authorization.ErrBuiltInRole),
		errors.Is(err, authentication.ErrAuthConfig),
		errors.Is(err, authentication.ErrInvalidLogin),
		errors.Is(err, authentication.ErrInvalidRSAKey),
		errors.Is(err, authentication.ErrUserDisabled),
		errors.Is(err, authentication.ErrUserLocked),
		errors.Is(err, authentication.ErrAdminRequired),
		errors.Is(err, authentication.ErrNoPermission),
		errors.Is(err, authentication.ErrTokenRequired),
		errors.Is(err, authentication.ErrInvalidToken),
		errors.Is(err, authentication.ErrInvalidArgument):
		hertzx.Badf(c, "%v", err)
	default:
		hertzx.Errorf(c, "系统异常，请联系管理员")
	}
}

func pagination(c *app.RequestContext) ormx.Pagination {
	return ormx.Pagination{
		Keyword:   c.DefaultQuery("keyword", ""),
		PageNo:    queryInt(c, "pageNo"),
		PageSize:  queryInt(c, "pageSize"),
		SortField: c.DefaultQuery("sortField", ""),
		SortOrder: c.DefaultQuery("sortOrder", ""),
	}
}

func queryInt(c *app.RequestContext, name string) int {
	result, err := hertzx.QueryInt(c, name)
	if err != nil {
		return 0
	}
	return result
}

func queryIntPtr(c *app.RequestContext, name string) *int {
	result, err := hertzx.QueryIntPtr(c, name)
	if err != nil {
		return nil
	}
	return result
}

func queryInt64(c *app.RequestContext, name string) int64 {
	result, err := hertzx.QueryInt64(c, name)
	if err != nil {
		return 0
	}
	return result
}

func queryInt64Ptr(c *app.RequestContext, name string) *int64 {
	result, err := hertzx.QueryInt64Ptr(c, name)
	if err != nil {
		return nil
	}
	return result
}

func (r *Router) tenantContext() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		tenantID, ok := r.resolveTenantID(ctx, c)
		if !ok {
			return
		}
		c.Next(context.WithValue(ctx, tenantContextKey{}, tenantID))
	}
}

func (r *Router) pathTenantContext() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		tenantID, ok := pathID(c)
		if !ok {
			return
		}
		tenant, err := r.q.GetTenant(ctx, tenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		if tenant == nil {
			handleError(c, fmt.Errorf("%w: 租户不存在", authorization.ErrNotFound))
			return
		}
		if !r.hasTenantPermission(ctx, c, tenantID) {
			handleError(c, authentication.ErrNoPermission)
			return
		}
		c.Next(context.WithValue(ctx, tenantContextKey{}, tenantID))
	}
}

func tenantIDFromContext(ctx context.Context) int64 {
	tenantID, _ := ctx.Value(tenantContextKey{}).(int64)
	return tenantID
}

func (r *Router) platformAdminContext() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		user, ok := r.currentLoginUser(ctx, c)
		if !ok {
			handleError(c, authentication.ErrNoPermission)
			return
		}
		if !isPlatformAdmin(user) {
			handleError(c, authentication.ErrNoPermission)
			return
		}
		c.Next(ctx)
	}
}

func (r *Router) tenantCreateContext() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		user, ok := r.currentLoginUser(ctx, c)
		if !ok {
			handleError(c, authentication.ErrNoPermission)
			return
		}
		if isPlatformAdmin(user) || configx.New(r.q).Bool(ctx, entity.PlatformTenantCreateEnabled, false) {
			c.Next(ctx)
			return
		}
		handleError(c, authentication.ErrNoPermission)
	}
}

func (r *Router) pathTenantManagerContext() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		tenantID, ok := pathID(c)
		if !ok {
			return
		}
		tenant, err := r.q.GetTenant(ctx, tenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		if tenant == nil {
			handleError(c, fmt.Errorf("%w: 租户不存在", authorization.ErrNotFound))
			return
		}
		if !r.hasTenantManagePermission(ctx, c, tenant) {
			handleError(c, authentication.ErrNoPermission)
			return
		}
		c.Next(context.WithValue(ctx, tenantContextKey{}, tenantID))
	}
}

func (r *Router) resolveTenantID(ctx context.Context, c *app.RequestContext) (int64, bool) {
	tenantID := int64(0)
	if code := strings.TrimSpace(string(c.GetHeader("X-TENANT-CODE"))); code != "" {
		tenant, err := r.q.GetTenantByCode(ctx, code)
		if err != nil {
			handleError(c, err)
			return 0, false
		}
		if tenant == nil {
			handleError(c, fmt.Errorf("%w: 租户不存在", authorization.ErrNotFound))
			return 0, false
		}
		tenantID = tenant.ID
	} else {
		headerName := "X-TENAT-ID"
		value := strings.TrimSpace(string(c.GetHeader(headerName)))
		if value == "" {
			headerName = "X-TENANT-ID"
			value = strings.TrimSpace(string(c.GetHeader(headerName)))
		}
		if value != "" {
			id, err := strconv.ParseInt(value, 10, 64)
			if err != nil || id < 0 {
				hertzx.Badf(c, "请求头 %s 不合法", headerName)
				return 0, false
			}
			tenantID = id
		} else {
			tenantID = queryInt64(c, "tenantId")
		}
	}

	if tenantID <= 0 {
		return 0, true
	}
	tenant, err := r.q.GetTenant(ctx, tenantID)
	if err != nil {
		handleError(c, err)
		return 0, false
	}
	if tenant == nil {
		handleError(c, fmt.Errorf("%w: 租户不存在", authorization.ErrNotFound))
		return 0, false
	}
	if !r.hasTenantPermission(ctx, c, tenantID) {
		handleError(c, authentication.ErrNoPermission)
		return 0, false
	}
	return tenantID, true
}

func (r *Router) hasTenantPermission(ctx context.Context, c *app.RequestContext, tenantID int64) bool {
	user, ok := r.currentLoginUser(ctx, c)
	if !ok {
		return false
	}
	if isPlatformAdmin(user) {
		return true
	}
	tenant, err := r.q.GetTenant(ctx, tenantID)
	if err == nil && tenant != nil && tenant.OwnerUserID == user.ID {
		return true
	}
	for _, role := range user.Roles {
		if strings.EqualFold(strings.TrimSpace(role.Name), entity.RoleOfSystemManager) && role.TenantID == tenantID {
			return true
		}
	}
	for _, tenant := range user.Tenants {
		if tenant.ID == tenantID {
			return true
		}
	}
	return false
}

func (r *Router) hasTenantManagePermission(ctx context.Context, c *app.RequestContext, tenant *entity.Tenant) bool {
	user, ok := r.currentLoginUser(ctx, c)
	if !ok || tenant == nil {
		return false
	}
	if isPlatformAdmin(user) {
		return true
	}
	if tenant.OwnerUserID > 0 && tenant.OwnerUserID == user.ID {
		return true
	}
	for _, role := range user.Roles {
		if strings.EqualFold(strings.TrimSpace(role.Name), entity.RoleOfSystemManager) && role.TenantID == tenant.ID {
			return true
		}
	}
	return false
}

func (r *Router) currentLoginUser(ctx context.Context, c *app.RequestContext) (*entity.User, bool) {
	token := normalizeAuthorizationToken(authorizationToken(c))
	loginID, err := htputil.GetLoginID(token)
	if err != nil {
		return nil, false
	}
	userID, err := strconv.ParseInt(loginID, 10, 64)
	if err != nil {
		return nil, false
	}
	user, err := r.q.GetUserForAuth(ctx, userID)
	if err != nil || user == nil {
		return nil, false
	}
	return user, true
}

func isPlatformAdmin(user *entity.User) bool {
	if user == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(user.Username), "admin") {
		return true
	}
	for _, role := range user.Roles {
		if strings.EqualFold(strings.TrimSpace(role.Name), entity.RoleOfPlatformManager) {
			return true
		}
	}
	return false
}

func health(ctx context.Context, c *app.RequestContext) {
	hertzx.Data(c, map[string]string{"status": "ok"})
}
