package authorization

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/xiehqing/hiauth/internal/db/entity"
	"github.com/xiehqing/hiauth/internal/db/queries"
	"github.com/xiehqing/infra/pkg/ormx"
)

var tenantCodePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{1,31}$`)

type CreateTenantRequest struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	OwnerUserID  int64  `json:"ownerUserId"`
	ContactName  string `json:"contactName"`
	ContactPhone string `json:"contactPhone"`
	ExpireAt     string `json:"expireAt"`
	Description  string `json:"description"`
	Status       *int   `json:"status"`
	Operator     string `json:"operator"`
}

type UpdateTenantRequest struct {
	ID           int64  `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	OwnerUserID  int64  `json:"ownerUserId"`
	ContactName  string `json:"contactName"`
	ContactPhone string `json:"contactPhone"`
	ExpireAt     string `json:"expireAt"`
	Description  string `json:"description"`
	Status       int    `json:"status"`
	Operator     string `json:"operator"`
}

type ListTenantsRequest struct {
	queries.TenantListFilter
}

type AddTenantUsersRequest struct {
	TenantID int64   `json:"tenantId"`
	UserIDs  []int64 `json:"userIds"`
	Operator string  `json:"operator"`
}

type InviteTenantUserRequest struct {
	TenantID     int64   `json:"tenantId"`
	Username     string  `json:"username"`
	Password     string  `json:"password"`
	Nickname     string  `json:"nickname"`
	Phone        string  `json:"phone"`
	Email        string  `json:"email"`
	Status       *int    `json:"status"`
	RoleIDs      []int64 `json:"roleIds"`
	DepartmentID int64   `json:"departmentId"`
	Operator     string  `json:"operator"`
}

func (as *Service) CreateTenant(ctx context.Context, req CreateTenantRequest) (*entity.Tenant, error) {
	if err := validateTenant(req.Code, req.Name); err != nil {
		return nil, err
	}
	status := entity.TenantStatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	if !validTenantStatus(status) {
		return nil, fmt.Errorf("%w: 租户状态不合法", ErrInvalidArgument)
	}
	expireAt, err := parseOptionalTime(req.ExpireAt)
	if err != nil {
		return nil, err
	}
	if req.OwnerUserID > 0 {
		users, err := as.queries.GetUsersByIDs(ctx, []int64{req.OwnerUserID})
		if err != nil {
			return nil, err
		}
		if len(users) == 0 {
			return nil, fmt.Errorf("%w: 租户所有者不存在", ErrInvalidArgument)
		}
	}

	tenant := &entity.Tenant{
		StatusAbleModel: ormx.StatusAbleModel{
			BaseModel: ormx.BaseModel{
				CreatedBy: normalizeString(req.Operator),
				UpdatedBy: normalizeString(req.Operator),
			},
			Status: status,
		},
		Code:         normalizeString(req.Code),
		Name:         normalizeString(req.Name),
		OwnerUserID:  req.OwnerUserID,
		ContactName:  normalizeString(req.ContactName),
		ContactPhone: normalizeString(req.ContactPhone),
		ExpireAt:     expireAt,
		Description:  normalizeString(req.Description),
	}
	if err := as.queries.CreateTenant(ctx, tenant); err != nil {
		return nil, normalizeDBError(err)
	}
	if req.OwnerUserID > 0 {
		if err := as.queries.AddUsersToTenant(ctx, tenant.ID, []int64{req.OwnerUserID}); err != nil {
			return nil, normalizeDBError(err)
		}
	}
	return as.queries.GetTenant(ctx, tenant.ID)
}

func (as *Service) UpdateTenant(ctx context.Context, req UpdateTenantRequest) (*entity.Tenant, error) {
	if req.ID <= 0 {
		return nil, ErrInvalidID
	}
	if err := validateTenant(req.Code, req.Name); err != nil {
		return nil, err
	}
	if !validTenantStatus(req.Status) {
		return nil, fmt.Errorf("%w: 租户状态不合法", ErrInvalidArgument)
	}
	before, err := as.GetTenant(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	ownerUserID := req.OwnerUserID
	if ownerUserID <= 0 {
		ownerUserID = before.OwnerUserID
	}
	if ownerUserID > 0 && ownerUserID != before.OwnerUserID {
		users, err := as.queries.GetUsersByIDs(ctx, []int64{ownerUserID})
		if err != nil {
			return nil, err
		}
		if len(users) == 0 {
			return nil, fmt.Errorf("%w: 租户所有者不存在", ErrInvalidArgument)
		}
	}
	expireAt, err := parseOptionalTime(req.ExpireAt)
	if err != nil {
		return nil, err
	}

	tenant := &entity.Tenant{
		StatusAbleModel: ormx.StatusAbleModel{
			BaseModel: ormx.BaseModel{
				ID:        req.ID,
				UpdatedBy: normalizeString(req.Operator),
			},
			Status: req.Status,
		},
		Code:         normalizeString(req.Code),
		Name:         normalizeString(req.Name),
		OwnerUserID:  ownerUserID,
		ContactName:  normalizeString(req.ContactName),
		ContactPhone: normalizeString(req.ContactPhone),
		ExpireAt:     expireAt,
		Description:  normalizeString(req.Description),
	}
	if err := as.queries.UpdateTenant(ctx, tenant); err != nil {
		return nil, normalizeDBError(err)
	}
	if ownerUserID > 0 {
		if err := as.queries.AddUsersToTenant(ctx, req.ID, []int64{ownerUserID}); err != nil {
			return nil, normalizeDBError(err)
		}
	}
	return as.GetTenant(ctx, req.ID)
}

func (as *Service) GetTenant(ctx context.Context, id int64) (*entity.Tenant, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	tenant, err := as.queries.GetTenant(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, ErrNotFound
	}
	return tenant, nil
}

func (as *Service) ListTenants(ctx context.Context, req ListTenantsRequest) (ormx.PageResult[entity.Tenant], error) {
	return as.queries.ListTenants(ctx, req.TenantListFilter)
}

func (as *Service) DeleteTenant(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}
	return normalizeDBError(as.queries.DeleteTenant(ctx, id))
}

func (as *Service) ListTenantUsers(ctx context.Context, tenantID int64) ([]entity.User, error) {
	if err := as.validateTenantAvailable(ctx, tenantID); err != nil {
		return nil, err
	}
	return as.queries.ListTenantUsers(ctx, tenantID)
}

func (as *Service) AddTenantUsers(ctx context.Context, req AddTenantUsersRequest) ([]entity.User, error) {
	if err := as.validateTenantAvailable(ctx, req.TenantID); err != nil {
		return nil, err
	}
	userIDs := normalizeIDs(req.UserIDs)
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("%w: 用户不能为空", ErrInvalidArgument)
	}
	users, err := as.queries.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	if len(users) != len(userIDs) {
		return nil, fmt.Errorf("%w: 存在无效用户", ErrInvalidArgument)
	}
	if err := as.queries.AddUsersToTenant(ctx, req.TenantID, userIDs); err != nil {
		return nil, normalizeDBError(err)
	}
	return as.queries.ListTenantUsers(ctx, req.TenantID)
}

func (as *Service) InviteTenantUser(ctx context.Context, req InviteTenantUserRequest) (*entity.User, error) {
	if err := as.validateTenantAvailable(ctx, req.TenantID); err != nil {
		return nil, err
	}
	username := normalizeString(req.Username)
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	user, err := as.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user, err = as.createInvitedTenantUser(ctx, req, username)
		if err != nil {
			return nil, err
		}
	} else {
		if err := as.queries.AddUsersToTenant(ctx, req.TenantID, []int64{user.ID}); err != nil {
			return nil, normalizeDBError(err)
		}
		if roleIDs := normalizeIDs(req.RoleIDs); len(roleIDs) > 0 {
			if err := as.queries.AddUserRoles(ctx, user.ID, roleIDs); err != nil {
				return nil, normalizeDBError(err)
			}
		}
	}
	return as.queries.GetUser(ctx, user.ID)
}

func (as *Service) createInvitedTenantUser(ctx context.Context, req InviteTenantUserRequest, username string) (*entity.User, error) {
	if !required(req.Password) || !required(req.Nickname) {
		return nil, fmt.Errorf("%w: 新用户密码和昵称不能为空", ErrInvalidArgument)
	}
	valid, pwdLength := as.checkPwdLength(ctx, req.Password)
	if !valid {
		return nil, fmt.Errorf("密码长度最少不得低于%d位数", pwdLength)
	}
	status := UserStatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	if !validUserStatus(status) {
		return nil, fmt.Errorf("%w: 用户状态不合法", ErrInvalidArgument)
	}
	user := &entity.User{
		StatusAbleModel: ormx.StatusAbleModel{
			BaseModel: ormx.BaseModel{
				CreatedBy: normalizeString(req.Operator),
				UpdatedBy: normalizeString(req.Operator),
			},
			Status: status,
		},
		Username:     username,
		Password:     md5Password(req.Password),
		Nickname:     normalizeString(req.Nickname),
		Phone:        normalizeString(req.Phone),
		Email:        normalizeString(req.Email),
		DepartmentID: req.DepartmentID,
	}
	if err := as.queries.CreateUser(ctx, user, normalizeIDs(req.RoleIDs), []int64{req.TenantID}); err != nil {
		return nil, normalizeDBError(err)
	}
	return as.queries.GetUser(ctx, user.ID)
}

func (as *Service) validateTenantAvailable(ctx context.Context, tenantID int64) error {
	if tenantID <= 0 {
		return ErrInvalidID
	}
	tenant, err := as.queries.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}
	if tenant == nil {
		return ErrNotFound
	}
	if tenant.Status != entity.TenantStatusEnabled {
		return fmt.Errorf("%w: 租户已禁用", ErrInvalidArgument)
	}
	return nil
}

func validateTenant(code, name string) error {
	code = normalizeString(code)
	if code == "" || !tenantCodePattern.MatchString(code) {
		return fmt.Errorf("%w: 租户编码必须以字母开头，长度 2-32，仅支持字母、数字、下划线和短横线", ErrInvalidArgument)
	}
	if !required(name) {
		return fmt.Errorf("%w: 租户名称不能为空", ErrInvalidArgument)
	}
	return nil
}

func validTenantStatus(status int) bool {
	return status == entity.TenantStatusDisabled || status == entity.TenantStatusEnabled
}

func parseOptionalTime(value string) (*time.Time, error) {
	value = normalizeString(value)
	if value == "" {
		return nil, nil
	}
	result, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		return nil, fmt.Errorf("%w: 时间格式不合法，请使用 yyyy-MM-dd HH:mm:ss", ErrInvalidArgument)
	}
	return &result, nil
}
