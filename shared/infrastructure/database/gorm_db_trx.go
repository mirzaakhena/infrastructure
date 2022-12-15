package database

import (
	"context"
	"gorm.io/gorm"
)

type contextDBType string

var ContextDBValue contextDBType = "gormDB"

type gormWrapper struct {
	db *gorm.DB
}

// ExtractDB is used by other repo to extract the databasex from context
func (r *gormWrapper) ExtractDB(ctx context.Context) *gorm.DB {

	db, ok := ctx.Value(ContextDBValue).(*gorm.DB)
	if !ok {
		return r.db
	}

	return db
}

type GormWithTransaction struct {
	*gormWrapper
}

func NewGormWithTransaction(db *gorm.DB) *GormWithTransaction {
	return &GormWithTransaction{
		gormWrapper: &gormWrapper{db: db},
	}
}

func (r *GormWithTransaction) BeginTransaction(ctx context.Context) (context.Context, error) {
	dbTrx := r.db.Begin()
	trxCtx := context.WithValue(ctx, ContextDBValue, dbTrx)
	return trxCtx, nil
}

func (r *GormWithTransaction) CommitTransaction(ctx context.Context) error {
	return r.ExtractDB(ctx).Commit().Error
}

func (r *GormWithTransaction) RollbackTransaction(ctx context.Context) error {
	return r.ExtractDB(ctx).Rollback().Error
}
