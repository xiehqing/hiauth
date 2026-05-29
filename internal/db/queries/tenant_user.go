package queries

import (
	"context"

	"github.com/xiehqing/hiauth/internal/db/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (q *Queries) AddUsersToTenant(ctx context.Context, tenantID int64, userIDs []int64) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := insertUserTenants(tx, tenantID, userIDs); err != nil {
			return err
		}
		return q.auditAuthorize(ctx, tx, "user_tenants", "user_tenants", tenantID, nil, map[string]any{
			"tenantId": tenantID,
			"userIds":  normalizeAuditIDs(userIDs),
		})
	})
}

func (q *Queries) AddUserRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := insertUserRoles(tx, userID, roleIDs); err != nil {
			return err
		}
		return q.auditAuthorize(ctx, tx, "user_roles", "user_roles", userID, nil, map[string]any{
			"userId":  userID,
			"roleIds": normalizeAuditIDs(roleIDs),
		})
	})
}

func (q *Queries) ListTenantUsers(ctx context.Context, tenantID int64) ([]entity.User, error) {
	var users []entity.User
	err := q.db.WithContext(ctx).
		Model(&entity.User{}).
		Joins("JOIN user_tenants ON user_tenants.user_id = user.id AND user_tenants.tenant_id = ?", tenantID).
		Preload("Tenants").
		Preload("Roles").
		Preload("Department").
		Order("id asc").
		Find(&users).
		Error
	return users, err
}

func (q *Queries) GetUsersByIDs(ctx context.Context, ids []int64) ([]entity.User, error) {
	ids = normalizeAuditIDs(ids)
	if len(ids) == 0 {
		return []entity.User{}, nil
	}
	var users []entity.User
	err := q.db.WithContext(ctx).
		Preload("Tenants").
		Preload("Roles").
		Preload("Department").
		Where("id IN ?", ids).
		Order("id asc").
		Find(&users).
		Error
	return users, err
}

func insertUserTenants(tx *gorm.DB, tenantID int64, userIDs []int64) error {
	userIDs = normalizeAuditIDs(userIDs)
	if tenantID <= 0 || len(userIDs) == 0 {
		return nil
	}
	rows := make([]map[string]any, 0, len(userIDs))
	for _, userID := range userIDs {
		rows = append(rows, map[string]any{
			"user_id":   userID,
			"tenant_id": tenantID,
		})
	}
	return tx.Table("user_tenants").Clauses(clause.OnConflict{DoNothing: true}).Create(rows).Error
}

func insertUserRoles(tx *gorm.DB, userID int64, roleIDs []int64) error {
	roleIDs = normalizeAuditIDs(roleIDs)
	if userID <= 0 || len(roleIDs) == 0 {
		return nil
	}
	rows := make([]map[string]any, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		rows = append(rows, map[string]any{
			"user_id": userID,
			"role_id": roleID,
		})
	}
	return tx.Table("user_roles").Clauses(clause.OnConflict{DoNothing: true}).Create(rows).Error
}
