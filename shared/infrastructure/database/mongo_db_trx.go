package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"infrastructure/shared/infrastructure/logger"
)

type MongoWithTransaction struct {
	MongoClient *mongo.Client
	log         logger.Logger
}

func NewMongoDBWithTransaction(c *mongo.Client, log logger.Logger) *MongoWithTransaction {
	return &MongoWithTransaction{
		MongoClient: c,
		log:         log,
	}
}

func (r *MongoWithTransaction) BeginTransaction(ctx context.Context) (context.Context, error) {

	r.log.Info(ctx, "Begin trx")

	session, err := r.MongoClient.StartSession()
	if err != nil {
		return nil, err
	}

	sessionCtx := mongo.NewSessionContext(ctx, session)

	err = session.StartTransaction()
	if err != nil {
		panic(err)
	}

	return sessionCtx, nil
}

func (r *MongoWithTransaction) CommitTransaction(ctx context.Context) error {

	r.log.Info(ctx, "Commit trx")

	err := mongo.SessionFromContext(ctx).CommitTransaction(ctx)
	if err != nil {
		return err
	}

	mongo.SessionFromContext(ctx).EndSession(ctx)

	return nil
}

func (r *MongoWithTransaction) RollbackTransaction(ctx context.Context) error {

	r.log.Info(ctx, "Rollback trx")

	err := mongo.SessionFromContext(ctx).AbortTransaction(ctx)
	if err != nil {
		return err
	}

	mongo.SessionFromContext(ctx).EndSession(ctx)

	return nil
}
