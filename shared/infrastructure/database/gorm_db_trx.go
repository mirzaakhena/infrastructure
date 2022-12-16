package database

import (
	"context"
	"gorm.io/gorm"
	"infrastructure/shared/infrastructure/logger"
)

type contextDBType string

var ContextDBValue contextDBType = "gormDB"

type gormWrapper struct {
	db  *gorm.DB
	log logger.Logger
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

func NewGormWithTransaction(db *gorm.DB, log logger.Logger) *GormWithTransaction {
	return &GormWithTransaction{
		gormWrapper: &gormWrapper{db: db, log: log},
	}
}

func (r *GormWithTransaction) BeginTransaction(ctx context.Context) (context.Context, error) {
	r.log.Info(ctx, "Begin trx")
	dbTrx := r.db.Begin()
	trxCtx := context.WithValue(ctx, ContextDBValue, dbTrx)
	return trxCtx, nil
}

func (r *GormWithTransaction) CommitTransaction(ctx context.Context) error {
	r.log.Info(ctx, "Commit trx")
	return r.ExtractDB(ctx).Commit().Error
}

func (r *GormWithTransaction) RollbackTransaction(ctx context.Context) error {
	r.log.Info(ctx, "Rollback trx")
	return r.ExtractDB(ctx).Rollback().Error
}
