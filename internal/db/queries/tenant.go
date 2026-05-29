package queries

import (
	"context"

	"github.com/xiehqing/hiauth/internal/db/entity"
	"github.com/xiehqing/infra/pkg/ormx"
	"gorm.io/gorm"
)

type TenantListFilter struct {
	ormx.Pagination
	Status *int  `json:"status" form:"status"`
	UserID int64 `json:"userId" form:"userId"`
}

func (q *Queries) CreateTenant(ctx context.Context, tenant *entity.Tenant) error {
	return q.db.WithContext(ctx).Create(tenant).Error
}

func (q *Queries) UpdateTenant(ctx context.Context, tenant *entity.Tenant) error {
	return q.db.WithContext(ctx).
		Model(&entity.Tenant{}).
		Where("id = ?", tenant.ID).
		Select("Code", "Name", "OwnerUserID", "ContactName", "ContactPhone", "ExpireAt", "Description", "Status", "UpdatedBy").
		Updates(tenant).
		Error
}

func (q *Queries) GetTenant(ctx context.Context, id int64) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := q.db.WithContext(ctx).First(&tenant, "id = ?", id).Error
	if err != nil {
		return nil, ormx.NotFoundAsNil(err)
	}
	return &tenant, nil
}

func (q *Queries) GetTenantByCode(ctx context.Context, code string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := q.db.WithContext(ctx).First(&tenant, "code = ?", code).Error
	if err != nil {
		return nil, ormx.NotFoundAsNil(err)
	}
	return &tenant, nil
}

func (q *Queries) DeleteTenant(ctx context.Context, id int64) error {
	tenant := entity.Tenant{StatusAbleModel: ormx.StatusAbleModel{BaseModel: ormx.BaseModel{ID: id}}}
	result := q.db.WithContext(ctx).Delete(&tenant)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (q *Queries) ListTenants(ctx context.Context, filter TenantListFilter) (ormx.PageResult[entity.Tenant], error) {
	db := q.db.WithContext(ctx).Model(&entity.Tenant{})
	if ormx.KeywordPresent(filter.Keyword) {
		keyword := ormx.LikeKeyword(filter.Keyword)
		db = db.Where("code LIKE ? OR name LIKE ? OR contact_name LIKE ? OR contact_phone LIKE ?", keyword, keyword, keyword, keyword)
	}
	if filter.Status != nil {
		db = db.Where("status = ?", *filter.Status)
	}
	if filter.UserID > 0 {
		db = db.Where("owner_user_id = ? OR id IN (SELECT tenant_id FROM user_tenants WHERE user_id = ?)", filter.UserID, filter.UserID)
	}

	return ormx.Paginate[entity.Tenant](db, filter.Pagination, map[string]string{
		"id":        "id",
		"code":      "code",
		"name":      "name",
		"status":    "status",
		"expireAt":  "expire_at",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
	})
}
