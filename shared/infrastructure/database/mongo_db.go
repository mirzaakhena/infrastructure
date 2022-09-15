package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"infrastructure/shared/util"

	"reflect"
	"strconv"
	"strings"
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
	//DatabaseName string
	Database *mongo.Database
}

func NewMongoWithTransaction(c *mongo.Client, databaseName string) *MongoWithTransaction {
	return &MongoWithTransaction{
		MongoClient: c,
		//DatabaseName: databaseName,
		Database: c.Database(databaseName),
	}
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

func (r *MongoWithTransaction) PrepareCollection(collectionObjs ...any) *MongoWithTransaction {

	existingCollectionNames, err := r.Database.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		panic(err)
	}

	mapCollName := map[string]int{}
	for _, name := range existingCollectionNames {
		mapCollName[name] = 1
	}

	for _, obj := range collectionObjs {

		nameInDB := getCollectionNameFormat(obj)

		coll := r.Database.Collection(nameInDB)

		if _, exist := mapCollName[nameInDB]; exist {
			continue
		}

		r.createCollection(coll, r.Database)
		r.collectIndex(coll, obj)

	}

	return r
}

func (r *MongoWithTransaction) PrepareCollectionIndex(collectionObjs []any) *MongoWithTransaction {

	for _, obj := range collectionObjs {

		//theType := reflect.TypeOf(obj)
		//
		//name := theType.Name()

		nameInDB := getCollectionNameFormat(reflect.TypeOf(obj).Name())

		coll := r.Database.Collection(nameInDB)

		r.collectIndex(coll, obj)

	}

	return r
}

// SaveOrUpdate Insert new collection or update the existing collection if the id is exist
//
//	_, err := r.SaveOrUpdate(ctx, string(obj.ID), obj)
//
//	if err != nil {
//		r.log.Error(ctx, err.Error())
//		return err
//	}
func (r *MongoWithTransaction) SaveOrUpdate(ctx context.Context, id string, data any) (any, error) {

	name := getCollectionNameFormat(data)
	coll := r.Database.Collection(name)

	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", data}}
	opts := options.Update().SetUpsert(true)

	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("%v %v %v", result.UpsertedCount, result.ModifiedCount, result.UpsertedID), nil
}

// SaveBulk can use SaveBulk(ctx, util.ToSliceAny(yourSliceObjects))
func (r *MongoWithTransaction) SaveBulk(ctx context.Context, datas []any) (any, error) {

	if len(datas) == 0 {
		return nil, fmt.Errorf("data must > 0")
	}

	name := getCollectionNameFormat(datas[0])

	coll := r.Database.Collection(name)

	info, err := coll.InsertMany(ctx, datas)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// GetAll return multiple data from collection
//
//	filter := bson.M{}
//	results := make([]*entity.YourEntity, 0)
//	count, err := r.GetAll(ctx, page, size, filter, &results)
//	if err != nil {
//		return nil, 0, err
//	}
func (r *MongoWithTransaction) GetAll(ctx context.Context, page, size int64, filter bson.M, results any) (int64, error) {

	name := r.getSliceElementName(results)

	coll := r.Database.Collection(name)

	skip := size * (page - 1)
	limit := size
	sort := bson.M{}

	findOpts := options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  sort,
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	cursor, err := coll.Find(ctx, filter, &findOpts)
	if err != nil {
		return 0, err
	}

	err = cursor.All(ctx, results)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetOne get only one result from collection
//
//	var result entity.YourEntity
//	err := r.GetOne(ctx, yourEntityID, &result)
//	if err != nil {
//		r.log.Error(ctx, err.Error())
//		return nil, err
//	}
func (r *MongoWithTransaction) GetOne(ctx context.Context, id string, result any) error {

	coll := r.GetCollection(result)

	filter := bson.M{"_id": id}

	singleResult := coll.FindOne(ctx, filter)

	err := singleResult.Decode(&result)
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoWithTransaction) GetCollection(obj any) *mongo.Collection {
	return r.Database.Collection(getCollectionNameFormat(obj))
}

func (r *MongoWithTransaction) getSliceElementName(results any) string {
	name := ""

	if reflect.TypeOf(results).Kind() == reflect.Ptr {

		if reflect.TypeOf(results).Elem().Kind() == reflect.Slice {

			if reflect.TypeOf(results).Elem().Elem().Kind() == reflect.Struct {

				name = reflect.TypeOf(results).Elem().Elem().Name()

			} else if reflect.TypeOf(results).Elem().Elem().Kind() == reflect.Ptr {

				name = reflect.TypeOf(results).Elem().Elem().Elem().Name()

			}

		}

	}
	return util.SnakeCase(name)
}

func getCollectionNameFormat(obj any) string {

	name := ""
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		name = reflect.TypeOf(obj).Elem().Name()
	} else if reflect.TypeOf(obj).Kind() == reflect.Struct {
		name = reflect.TypeOf(obj).Name()
	}

	return util.SnakeCase(name)
}

func getCollectionFieldNameFormat(x string) string {
	return util.SnakeCase(x)
}

func (r *MongoWithTransaction) collectIndex(coll *mongo.Collection, obj any) {

	theType := reflect.TypeOf(obj)

	docs := bson.D{}
	for i := 0; i < theType.NumField(); i++ {
		theField := theType.Field(i)
		tagValue, exist := theField.Tag.Lookup("index")
		if !exist {
			continue
		}

		atoi, err := strconv.Atoi(tagValue)
		if err != nil {
			panic(err.Error())
		}

		docs = append(docs, bson.E{Key: strings.ToLower(getCollectionFieldNameFormat(theField.Name)), Value: atoi})
	}

	if len(docs) > 0 {
		_, err := coll.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
			Keys: docs,
			// Options: options.Index().SetUnique(true).SetExpireAfterSeconds(1),
		})
		if err != nil {
			panic(err)
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
