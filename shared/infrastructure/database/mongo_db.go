package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// NewMongoDefault uri := "mongodb://localhost:27017/?replicaSet=rs0&readPreference=primary&ssl=false"
func NewMongoDefault(uri string) *mongo.Client {

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

func (r *MongoWithTransaction) PrepareCollection(databaseName string, collectionNames []string) {
	db := r.MongoClient.Database(databaseName)

	existingCollectionNames, err := db.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		panic(err)
	}

	mapCollName := map[string]int{}
	for _, name := range existingCollectionNames {
		mapCollName[name] = 1
	}

	for _, name := range collectionNames {
		if _, exist := mapCollName[name]; !exist {
			r.createCollection(db.Collection(name), db)
		}
	}
}

func (r *MongoWithTransaction) createCollection(coll *mongo.Collection, db *mongo.Database) {
	createCmd := bson.D{{"create", coll.Name()}}
	res := db.RunCommand(context.Background(), createCmd)
	err := res.Err()
	if err != nil {
		panic(err)
	}
}

func (r *MongoWithTransaction) SaveOrUpdate(ctx context.Context, databaseName, collectionName string, id primitive.ObjectID, data interface{}) (interface{}, error) {

	coll := r.MongoClient.Database(databaseName).Collection(collectionName)

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", data}}
	opts := options.Update().SetUpsert(true)

	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("%v %v %v\n", result.UpsertedCount, result.ModifiedCount, result.UpsertedID), nil
}
