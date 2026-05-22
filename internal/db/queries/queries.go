package queries

import (
	"context"
	"github.com/xiehqing/hiauth/internal/db/entity"
	"gorm.io/gorm"
)

type Queries struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Queries {
	return &Queries{
		db: db,
	}
}

func (q *Queries) DB() *gorm.DB {
	return q.db
}

func (q *Queries) AutoMigrate(ctx context.Context) error {
	return q.db.WithContext(ctx).AutoMigrate(
		&entity.User{},
		&entity.Role{},
		&entity.Department{},
		&entity.Menu{},
		&entity.SystemConfig{},
		&entity.AuditLog{},
		&entity.AuditChange{},
	)
}
