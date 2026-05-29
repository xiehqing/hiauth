package entity

import "github.com/xiehqing/infra/pkg/ormx"

const (
	RoleCustom            = 0
	RoleBuiltIn           = 1
	RoleOfPlatformManager = "role_platform_admin"
	RoleOfSystemManager   = "role_admin"
)

type Role struct {
	ormx.BaseModel
	TenantID    int64  `json:"tenantId" gorm:"type:bigint;not null;default:0;uniqueIndex:idx_role_tenant_name;index;comment:'租户ID'"`
	DisplayName string `json:"displayName" gorm:"type:varchar(64);not null;comment:'角色名称'"`
	Name        string `json:"name" gorm:"type:varchar(64);not null;uniqueIndex:idx_role_tenant_name;comment:'角色标识'"`
	Description string `json:"description" gorm:"type:varchar(255);comment:'角色描述'"`
	BuiltIn     int    `json:"builtIn" gorm:"type:int(11);not null;default:0;index;comment:'built-in flag: 0 custom, 1 built-in'"`
	Menus       []Menu `json:"menus,omitempty" gorm:"many2many:role_menus;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (r *Role) TableName() string {
	return "role"
}
