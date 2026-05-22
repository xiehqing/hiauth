package authorization

import (
	"context"
	"fmt"
	"github.com/xiehqing/infra/pkg/ormx"

	"hiauth/internal/db/entity"
	"hiauth/internal/db/queries"
)

type CreateRoleRequest struct {
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Operator    string `json:"operator"`
}

type UpdateRoleRequest struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Operator    string `json:"operator"`
}

type ListRolesRequest struct {
	queries.RoleListFilter
}

type UpdateRoleMenusRequest struct {
	RoleID   int64   `json:"roleId"`
	MenuIDs  []int64 `json:"menuIds"`
	Operator string  `json:"operator"`
}

func (as *Service) CreateRole(ctx context.Context, req CreateRoleRequest) (*entity.Role, error) {
	if !required(req.DisplayName) || !required(req.Name) {
		return nil, fmt.Errorf("%w: 角色名称和角色标识不能为空", ErrInvalidArgument)
	}

	role := &entity.Role{
		BaseModel: ormx.BaseModel{
			CreatedBy: normalizeString(req.Operator),
			UpdatedBy: normalizeString(req.Operator),
		},
		DisplayName: normalizeString(req.DisplayName),
		Name:        normalizeString(req.Name),
		Description: normalizeString(req.Description),
		BuiltIn:     entity.RoleCustom,
	}

	if err := as.queries.CreateRole(ctx, role); err != nil {
		return nil, normalizeDBError(err)
	}
	return as.queries.GetRole(ctx, role.ID)
}

func (as *Service) UpdateRole(ctx context.Context, req UpdateRoleRequest) (*entity.Role, error) {
	if req.ID <= 0 {
		return nil, ErrInvalidID
	}
	if !required(req.DisplayName) {
		return nil, fmt.Errorf("%w: 角色名称不能为空", ErrInvalidArgument)
	}

	role := &entity.Role{
		BaseModel: ormx.BaseModel{
			ID:        req.ID,
			UpdatedBy: normalizeString(req.Operator),
		},
		DisplayName: normalizeString(req.DisplayName),
		Description: normalizeString(req.Description),
	}

	if err := as.queries.UpdateRole(ctx, role); err != nil {
		return nil, normalizeDBError(err)
	}
	return as.GetRole(ctx, req.ID)
}

func (as *Service) GetRole(ctx context.Context, id int64) (*entity.Role, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	role, err := as.queries.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrNotFound
	}
	return role, nil
}

func (as *Service) ListRoles(ctx context.Context, req ListRolesRequest) (ormx.PageResult[entity.Role], error) {
	if req.BuiltIn != nil && *req.BuiltIn != entity.RoleCustom && *req.BuiltIn != entity.RoleBuiltIn {
		return ormx.PageResult[entity.Role]{}, fmt.Errorf("%w: 角色类型不合法", ErrInvalidArgument)
	}
	return as.queries.ListRoles(ctx, req.RoleListFilter)
}

func (as *Service) ListAllRoles(ctx context.Context) ([]entity.Role, error) {
	return as.queries.ListAllRoles(ctx)
}

func (as *Service) DeleteRole(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}
	role, err := as.GetRole(ctx, id)
	if err != nil {
		return err
	}
	if role.BuiltIn == entity.RoleBuiltIn {
		return ErrBuiltInRole
	}
	return normalizeDBError(as.queries.DeleteRole(ctx, id))
}

func (as *Service) UpdateRoleMenus(ctx context.Context, req UpdateRoleMenusRequest) ([]entity.Menu, error) {
	if req.RoleID <= 0 {
		return nil, ErrInvalidID
	}
	if err := as.queries.UpdateRoleMenus(ctx, req.RoleID, normalizeIDs(req.MenuIDs)); err != nil {
		return nil, normalizeDBError(err)
	}
	return as.GetRoleMenus(ctx, req.RoleID)
}

func (as *Service) GetRoleMenus(ctx context.Context, roleID int64) ([]entity.Menu, error) {
	if roleID <= 0 {
		return nil, ErrInvalidID
	}
	menus, err := as.queries.GetRoleMenus(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return menus, nil
}
