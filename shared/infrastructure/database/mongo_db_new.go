package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"reflect"
	"regexp"
	"strings"
)

type GetAllParam struct {
	Page   int64
	Size   int64
	Sort   map[string]any
	Filter map[string]any
}

func (g GetAllParam) SetPage(page int64) GetAllParam {
	g.Page = page
	return g
}

func (g GetAllParam) SetSize(size int64) GetAllParam {
	g.Size = size
	return g
}

func (g GetAllParam) SetSort(field string, sort any) GetAllParam {
	g.Sort[field] = sort
	return g
}

func (g GetAllParam) SetFilter(key string, value any) GetAllParam {

	if reflect.ValueOf(value).String() == "" {
		return g
	}

	g.Filter[key] = value
	return g
}

func NewDefaultParam() GetAllParam {
	return GetAllParam{
		Page:   1,
		Size:   2000,
		Sort:   map[string]any{},
		Filter: map[string]any{},
	}
}

// =======================================

type InsertOrUpdateRepo[T any] interface {
	InsertOrUpdate(obj *T) error
}

type InsertManyRepo[T any] interface {
	InsertMany(objs ...*T) error
}

type GetOneRepo[T any] interface {
	GetOne(filter map[string]any, result *T) error
}

type GetAllRepo[T any] interface {
	GetAll(param GetAllParam, results *[]*T) (int64, error)
}

type GetAllEachItemRepo[T any] interface {
	GetAllEachItem(param GetAllParam, resultEachItem func(result T)) (int64, error)
}

type DeleteRepo[T any] interface {
	Delete(filter map[string]any) error
}

type Repository[T any] interface {
	InsertOrUpdateRepo[T]
	InsertManyRepo[T]
	GetOneRepo[T]
	GetAllRepo[T]
	GetAllEachItemRepo[T]
	DeleteRepo[T]
	GetTypeName() string
}

var matchFirstCapSnakeCase = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCapSnakeCase = regexp.MustCompile("([a-z\\d])([A-Z])")

// SnakeCase is
func snakeCase(str string) string {
	snake := matchFirstCapSnakeCase.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCapSnakeCase.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func toSliceAny[T any](objs []T) []any {
	var results []any
	for _, obj := range objs {
		results = append(results, obj)
	}
	return results
}

// =======================================

type MongoGateway[T any] struct {
	Database *mongo.Database
}

func NewMongoGateway[T any](db *mongo.Database) *MongoGateway[T] {
	return &MongoGateway[T]{
		Database: db,
	}
}

func NewDatabase(databaseName string) *mongo.Database {

	uri := "mongodb://localhost:27017/?readPreference=primary&ssl=false"

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))

	err = client.Connect(context.Background())
	if err != nil {
		panic(err)
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		panic(err)
	}

	return client.Database(databaseName)

}

func (g *MongoGateway[T]) GetTypeName() string {
	var x T
	return snakeCase(reflect.TypeOf(x).Name())
}

//func (g *MongoGateway[T]) GetCollection() *mongo.Collection {
//	var x T
//	name := snakeCase(reflect.TypeOf(x).Name())
//	return g.Database.Collection(name)
//}

func (g *MongoGateway[T]) InsertOrUpdate(obj *T) error {

	sf, exist := reflect.TypeOf(obj).Elem().FieldByName("ID")
	if !exist {
		return fmt.Errorf("field ID as primary key is not found in %s", reflect.TypeOf(obj).Name())
	}

	tagValue, exist := sf.Tag.Lookup("bson")
	if !exist || tagValue != "_id" {
		return fmt.Errorf("field ID must have tag `bson:\"_id\"`")
	}

	filter := bson.D{{"_id", reflect.ValueOf(obj).Elem().FieldByName("ID").Interface()}}
	update := bson.D{{"$set", obj}}
	opts := options.Update().SetUpsert(true)

	coll := g.Database.Collection(g.GetTypeName())
	_, err := coll.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

func (g *MongoGateway[T]) InsertMany(objs ...*T) error {

	if len(objs) == 0 {
		return fmt.Errorf("objs must > 0")
	}

	opts := options.InsertMany().SetOrdered(false)

	coll := g.Database.Collection(g.GetTypeName())
	_, err := coll.InsertMany(context.TODO(), toSliceAny(objs), opts)
	if err != nil {
		return err
	}

	return nil
}

func (g *MongoGateway[T]) GetOne(filter map[string]any, result *T) error {

	coll := g.Database.Collection(g.GetTypeName())

	singleResult := coll.FindOne(context.TODO(), filter)

	err := singleResult.Decode(result)
	if err != nil {
		return err
	}

	return nil
}

func (g *MongoGateway[T]) GetAll(param GetAllParam, results *[]*T) (int64, error) {

	coll := g.Database.Collection(g.GetTypeName())

	skip := param.Size * (param.Page - 1)
	limit := param.Size

	findOpts := options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  param.Sort,
	}

	ctx := context.TODO()

	count, err := coll.CountDocuments(ctx, param.Filter)
	if err != nil {
		return 0, err
	}

	cursor, err := coll.Find(ctx, param.Filter, &findOpts)
	if err != nil {
		return 0, err
	}

	err = cursor.All(ctx, results)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (g *MongoGateway[T]) GetAllEachItem(param GetAllParam, resultEachItem func(result T)) (int64, error) {

	coll := g.Database.Collection(g.GetTypeName())

	skip := param.Size * (param.Page - 1)
	limit := param.Size

	findOpts := options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  param.Sort,
	}

	ctx := context.TODO()

	count, err := coll.CountDocuments(ctx, param.Filter)
	if err != nil {
		return 0, err
	}

	cursor, err := coll.Find(ctx, param.Filter, &findOpts)
	if err != nil {
		return 0, err
	}

	for cursor.Next(ctx) {

		var result T
		err := cursor.Decode(&result)
		if err != nil {
			return 0, err
		}

		resultEachItem(result)

	}

	err = cursor.Err()
	if err != nil {
		return 0, err
	}

	return count, nil

}

func (g *MongoGateway[T]) Delete(filter map[string]any) error {

	coll := g.Database.Collection(g.GetTypeName())

	_, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}

//func (r *MongoWithTransaction) PrepareCollection(collectionObjs ...any) *MongoWithTransaction {
//
//	existingCollectionNames, err := r.Database.ListCollectionNames(context.Background(), bson.D{})
//	if err != nil {
//		panic(err)
//	}
//
//	mapCollName := map[string]int{}
//	for _, name := range existingCollectionNames {
//		mapCollName[name] = 1
//	}
//
//	for _, obj := range collectionObjs {
//
//		nameInDB := getCollectionNameFormat(obj)
//
//		coll := r.Database.Collection(nameInDB)
//
//		if _, exist := mapCollName[nameInDB]; exist {
//			continue
//		}
//
//		r.createCollection(coll, r.Database)
//		r.collectIndex(coll, obj)
//
//	}
//
//	return r
//}
//
//func (r *MongoWithTransaction) PrepareCollectionIndex(collectionObjs []any) *MongoWithTransaction {
//
//	for _, obj := range collectionObjs {
//
//		//theType := reflect.TypeOf(obj)
//		//
//		//name := theType.Name()
//
//		nameInDB := getCollectionNameFormat(reflect.TypeOf(obj).Name())
//
//		coll := r.Database.Collection(nameInDB)
//
//		r.collectIndex(coll, obj)
//
//	}
//
//	return r
//}

// SaveOrUpdate Insert new collection or update the existing collection if the id is exist
//
//	_, err := r.SaveOrUpdate(ctx, string(obj.ID), obj)
//
//	if err != nil {
//		r.log.Error(ctx, err.Error())
//		return err
//	}
//func (r *MongoWithTransaction) SaveOrUpdate(ctx context.Context, id string, data any) (any, error) {
//
//	name := getCollectionNameFormat(data)
//	coll := r.Database.Collection(name)
//
//	filter := bson.D{{"_id", id}}
//	update := bson.D{{"$set", data}}
//	opts := options.Update().SetUpsert(true)
//
//	result, err := coll.UpdateOne(ctx, filter, update, opts)
//	if err != nil {
//		return nil, err
//	}
//
//	return fmt.Sprintf("%v %v %v", result.UpsertedCount, result.ModifiedCount, result.UpsertedID), nil
//}

//// SaveBulk can use SaveBulk(ctx, util.ToSliceAny(yourSliceObjects))
//func (r *MongoWithTransaction) SaveBulk(ctx context.Context, datas []any) (any, error) {
//
//	if len(datas) == 0 {
//		return nil, fmt.Errorf("data must > 0")
//	}
//
//	name := getCollectionNameFormat(datas[0])
//
//	coll := r.Database.Collection(name)
//
//	info, err := coll.InsertMany(ctx, datas)
//	if err != nil {
//		return nil, err
//	}
//
//	return info, nil
//}

// GetAll return multiple data from collection
//
//	filter := bson.M{}
//	results := make([]*entity.YourEntity, 0)
//	count, err := r.GetAll(ctx, page, size, filter, &results)
//	if err != nil {
//		return nil, 0, err
//	}
//func (r *MongoWithTransaction) GetAll(ctx context.Context, page, size int64, filter bson.M, results any) (int64, error) {
//
//	name := r.getSliceElementName(results)
//
//	coll := r.Database.Collection(name)
//
//	skip := size * (page - 1)
//	limit := size
//	sort := bson.M{}
//
//	findOpts := options.FindOptions{
//		Limit: &limit,
//		Skip:  &skip,
//		Sort:  sort,
//	}
//
//	count, err := coll.CountDocuments(ctx, filter)
//	if err != nil {
//		return 0, err
//	}
//
//	cursor, err := coll.Find(ctx, filter, &findOpts)
//	if err != nil {
//		return 0, err
//	}
//
//	err = cursor.All(ctx, results)
//	if err != nil {
//		return 0, err
//	}
//
//	return count, nil
//}

// GetOne get only one result from collection
//
//	var result entity.YourEntity
//	err := r.GetOne(ctx, yourEntityID, &result)
//	if err != nil {
//		r.log.Error(ctx, err.Error())
//		return nil, err
//	}
//func (r *MongoWithTransaction) GetOne(ctx context.Context, id string, result any) error {
//
//	coll := r.GetCollection(result)
//
//	filter := bson.M{"_id": id}
//
//	singleResult := coll.FindOne(ctx, filter)
//
//	err := singleResult.Decode(&result)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

//func (r *MongoWithTransaction) GetCollection(obj any) *mongo.Collection {
//	return r.Database.Collection(getCollectionNameFormat(obj))
//}

//func (r *MongoWithTransaction) getSliceElementName(results any) string {
//	name := ""
//
//	if reflect.TypeOf(results).Kind() == reflect.Ptr {
//
//		if reflect.TypeOf(results).Elem().Kind() == reflect.Slice {
//
//			if reflect.TypeOf(results).Elem().Elem().Kind() == reflect.Struct {
//
//				name = reflect.TypeOf(results).Elem().Elem().Name()
//
//			} else if reflect.TypeOf(results).Elem().Elem().Kind() == reflect.Ptr {
//
//				name = reflect.TypeOf(results).Elem().Elem().Elem().Name()
//
//			}
//
//		}
//
//	}
//	return util.SnakeCase(name)
//}
//
//func getCollectionNameFormat(obj any) string {
//
//	name := ""
//	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
//		name = reflect.TypeOf(obj).Elem().Name()
//	} else if reflect.TypeOf(obj).Kind() == reflect.Struct {
//		name = reflect.TypeOf(obj).Name()
//	}
//
//	return util.SnakeCase(name)
//}
//
//func getCollectionFieldNameFormat(x string) string {
//	return util.SnakeCase(x)
//}
//
//func (r *MongoWithTransaction) collectIndex(coll *mongo.Collection, obj any) {
//
//	theType := reflect.TypeOf(obj)
//
//	docs := bson.D{}
//	for i := 0; i < theType.NumField(); i++ {
//		theField := theType.Field(i)
//		tagValue, exist := theField.Tag.Lookup("index")
//		if !exist {
//			continue
//		}
//
//		atoi, err := strconv.Atoi(tagValue)
//		if err != nil {
//			panic(err.Error())
//		}
//
//		docs = append(docs, bson.E{Key: strings.ToLower(getCollectionFieldNameFormat(theField.Name)), Value: atoi})
//	}
//
//	if len(docs) > 0 {
//		_, err := coll.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
//			Keys: docs,
//			// Options: options.Index().SetUnique(true).SetExpireAfterSeconds(1),
//		})
//		if err != nil {
//			panic(err)
//		}
//	}
//
//}
//
//func (r *MongoWithTransaction) createCollection(coll *mongo.Collection, db *mongo.Database) {
//	createCmd := bson.D{{"create", coll.Name()}}
//	res := db.RunCommand(context.Background(), createCmd)
//	err := res.Err()
//	if err != nil {
//		panic(err)
//	}
//}

// NewMongoDefault uri := "mongodb://localhost:27017/?replicaSet=rs0&readPreference=primary&ssl=false"
//func NewMongoDefault(uri string) *mongo.Client {
//
//	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
//
//	err = client.Connect(context.Background())
//	if err != nil {
//		panic(err)
//	}
//
//	err = client.Ping(context.TODO(), readpref.Primary())
//	if err != nil {
//		panic(err)
//	}
//
//	return client
//}
//
//type MongoWithoutTransaction struct {
//	MongoClient *mongo.Client
//}
//
//func NewMongoWithoutTransaction(c *mongo.Client) *MongoWithoutTransaction {
//	return &MongoWithoutTransaction{MongoClient: c}
//}
//
//func (r *MongoWithoutTransaction) GetDatabase(ctx context.Context) (context.Context, error) {
//	session, err := r.MongoClient.StartSession()
//	if err != nil {
//		return nil, err
//	}
//
//	sessionCtx := mongo.NewSessionContext(ctx, session)
//
//	return sessionCtx, nil
//}
//
//func (r *MongoWithoutTransaction) Close(ctx context.Context) error {
//	mongo.SessionFromContext(ctx).EndSession(ctx)
//	return nil
//}
