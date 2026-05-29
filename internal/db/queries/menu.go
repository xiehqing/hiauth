package queries

import (
	"context"
	"github.com/xiehqing/infra/pkg/ormx"

	"github.com/xiehqing/hiauth/internal/db/entity"

	"gorm.io/gorm"
)

type MenuListFilter struct {
	TenantID int64  `json:"tenantId" form:"tenantId"`
	Keyword  string `json:"keyword" form:"keyword"`
	Type     *int   `json:"type" form:"type"`
	ParentID *int64 `json:"parentId" form:"parentId"`
	Show     *int   `json:"show" form:"show"`
}

func (q *Queries) CreateMenu(ctx context.Context, menu *entity.Menu) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		roleIDs, err := firstChildInheritRoleIDs(tx, menu.ParentID)
		if err != nil {
			return err
		}
		if err := tx.Create(menu).Error; err != nil {
			return err
		}
		if err := q.auditCreate(ctx, tx, "menu", menu.TableName(), menu.ID, menu); err != nil {
			return err
		}
		if len(roleIDs) == 0 {
			return nil
		}
		if err := inheritMenuRoles(tx, menu.ID, roleIDs); err != nil {
			return err
		}
		return q.auditAuthorize(ctx, tx, "role_menus", "role_menus", menu.ID, map[string]any{
			"menuId":  menu.ID,
			"roleIds": []int64{},
		}, map[string]any{
			"menuId":  menu.ID,
			"roleIds": roleIDs,
		})
	})
}

func (q *Queries) UpdateMenu(ctx context.Context, menu *entity.Menu) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var before entity.Menu
		beforeDB := tx
		if menu.TenantID > 0 {
			beforeDB = beforeDB.Where("tenant_id = ?", menu.TenantID)
		}
		if err := beforeDB.First(&before, "id = ?", menu.ID).Error; err != nil {
			return err
		}
		updateDB := tx.Model(&entity.Menu{}).Where("id = ?", menu.ID)
		if menu.TenantID > 0 {
			updateDB = updateDB.Where("tenant_id = ?", menu.TenantID)
		}
		if err := updateDB.Select("Type", "ParentID", "Name", "Route", "Sort", "Icon", "Show", "UpdatedBy").Updates(menu).Error; err != nil {
			return err
		}
		var after entity.Menu
		afterDB := tx
		if menu.TenantID > 0 {
			afterDB = afterDB.Where("tenant_id = ?", menu.TenantID)
		}
		if err := afterDB.First(&after, "id = ?", menu.ID).Error; err != nil {
			return err
		}
		return q.auditUpdate(ctx, tx, "menu", after.TableName(), menu.ID, before, after)
	})
}

func (q *Queries) GetMenu(ctx context.Context, id int64) (*entity.Menu, error) {
	var menu entity.Menu
	err := q.db.WithContext(ctx).First(&menu, "id = ?", id).Error
	if err != nil {
		return nil, ormx.NotFoundAsNil(err)
	}
	return &menu, nil
}

func (q *Queries) HasChildMenus(ctx context.Context, parentID int64) (bool, error) {
	var count int64
	err := q.db.WithContext(ctx).Model(&entity.Menu{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count > 0, err
}

func (q *Queries) DeleteMenu(ctx context.Context, id int64) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var before entity.Menu
		if err := tx.First(&before, "id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Exec("DELETE FROM role_menus WHERE menu_id = ?", id).Error; err != nil {
			return err
		}

		menu := entity.Menu{BaseModel: ormx.BaseModel{ID: id}}
		result := tx.Delete(&menu)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return q.auditDelete(ctx, tx, "menu", before.TableName(), id, before)
	})
}

func (q *Queries) ListMenus(ctx context.Context, filter MenuListFilter) ([]entity.Menu, error) {
	db := q.db.WithContext(ctx).Model(&entity.Menu{})
	if filter.TenantID > 0 {
		db = db.Where("tenant_id = ?", filter.TenantID)
	}
	if ormx.KeywordPresent(filter.Keyword) {
		keyword := ormx.LikeKeyword(filter.Keyword)
		db = db.Where("name LIKE ? OR route LIKE ?", keyword, keyword)
	}
	if filter.Type != nil {
		db = db.Where("type = ?", *filter.Type)
	}
	if filter.ParentID != nil {
		db = db.Where("parent_id = ?", *filter.ParentID)
	}
	if filter.Show != nil {
		db = db.Where("show = ?", *filter.Show)
	}

	var menus []entity.Menu
	err := db.Order("sort asc, id asc").Find(&menus).Error
	return menus, err
}

func (q *Queries) ListAllMenus(ctx context.Context, tenantID int64) ([]entity.Menu, error) {
	var menus []entity.Menu
	db := q.db.WithContext(ctx)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	err := db.Order("sort asc, id asc").Find(&menus).Error
	return menus, err
}

func firstChildInheritRoleIDs(tx *gorm.DB, parentID int64) ([]int64, error) {
	if parentID <= 0 {
		return nil, nil
	}

	var childCount int64
	if err := tx.Model(&entity.Menu{}).Where("parent_id = ?", parentID).Count(&childCount).Error; err != nil {
		return nil, err
	}
	if childCount > 0 {
		return nil, nil
	}

	var roleIDs []int64
	err := tx.Table("role_menus").
		Where("menu_id = ?", parentID).
		Order("role_id asc").
		Pluck("role_id", &roleIDs).
		Error
	return roleIDs, err
}

func inheritMenuRoles(tx *gorm.DB, menuID int64, roleIDs []int64) error {
	for _, roleID := range roleIDs {
		if roleID <= 0 {
			continue
		}
		if err := tx.Exec("INSERT IGNORE INTO role_menus (role_id, menu_id) VALUES (?, ?)", roleID, menuID).Error; err != nil {
			return err
		}
	}
	return nil
}
