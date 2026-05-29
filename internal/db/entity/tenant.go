package entity

import (
	"time"

	"github.com/xiehqing/infra/pkg/ormx"
)

const (
	TenantStatusDisabled = 0
	TenantStatusEnabled  = 1
)

type Tenant struct {
	ormx.StatusAbleModel
	Code         string     `json:"code" gorm:"type:varchar(64);not null;uniqueIndex;comment:'tenant code'"`
	Name         string     `json:"name" gorm:"type:varchar(128);not null;comment:'tenant name'"`
	OwnerUserID  int64      `json:"ownerUserId" gorm:"column:owner_user_id;type:bigint;not null;default:0;index;comment:'tenant owner user id'"`
	ContactName  string     `json:"contactName" gorm:"type:varchar(64);comment:'contact name'"`
	ContactPhone string     `json:"contactPhone" gorm:"type:varchar(32);comment:'contact phone'"`
	ExpireAt     *time.Time `json:"expireAt" gorm:"type:dateTime;comment:'expire time'"`
	Description  string     `json:"description" gorm:"type:varchar(255);comment:'description'"`
}

func (t *Tenant) TableName() string {
	return "tenant"
}
