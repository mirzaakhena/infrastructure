package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func NewMongoDefault() *mongo.Client {

	// TODO fill this URI later
	uri := "mongodb://localhost:27017/?replicaSet=rs0&readPreference=primary&ssl=false"

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))

	err = client.Connect(context.Background())
	if err != nil {
		panic(err)
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		panic(err)
	}

	return client
}

type MongoWithoutTransaction struct {
	MongoClient *mongo.Client
}

func NewMongoWithoutTransaction(c *mongo.Client) *MongoWithoutTransaction {
	return &MongoWithoutTransaction{MongoClient: c}
}

func (r *MongoWithoutTransaction) GetDatabase(ctx context.Context) (context.Context, error) {
	session, err := r.MongoClient.StartSession()
	if err != nil {
		return nil, err
	}

	sessionCtx := mongo.NewSessionContext(ctx, session)

	return sessionCtx, nil
}

func (r *MongoWithoutTransaction) Close(ctx context.Context) error {
	mongo.SessionFromContext(ctx).EndSession(ctx)
	return nil
}

//----------------------------------------------------------------------------------------

type MongoWithTransaction struct {
	MongoClient *mongo.Client
}

func NewMongoWithTransaction(c *mongo.Client) *MongoWithTransaction {
	return &MongoWithTransaction{MongoClient: c}
}

func (r *MongoWithTransaction) BeginTransaction(ctx context.Context) (context.Context, error) {

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

	err := mongo.SessionFromContext(ctx).CommitTransaction(ctx)
	if err != nil {
		return err
	}

	mongo.SessionFromContext(ctx).EndSession(ctx)

	return nil
}

func (r *MongoWithTransaction) RollbackTransaction(ctx context.Context) error {

	err := mongo.SessionFromContext(ctx).AbortTransaction(ctx)
	if err != nil {
		return err
	}

	mongo.SessionFromContext(ctx).EndSession(ctx)

	return nil
}
