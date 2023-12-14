package mongoDB

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jeffyfung/flight-info-agg/config"
	"github.com/jeffyfung/flight-info-agg/pkg/collection"
	"github.com/jeffyfung/flight-info-agg/pkg/customContext"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

func InitDB() error {

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(config.Cfg.Database.MongodbUri).SetServerAPIOptions(serverAPI)
	opts.SetConnectTimeout(2 * time.Second)
	opts.SetBSONOptions(&options.BSONOptions{UseJSONStructTags: true})

	client, err := mongo.Connect(customContext.EmptyCtx, opts)
	if err != nil {
		return errors.New("cannot connect to MongoDB")
	}

	if err := client.Database("admin").RunCommand(customContext.EmptyCtx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return errors.New("Cannot ping MongoDB: " + err.Error())
	}
	fmt.Println("Connected to MongoDB")

	db = client.Database("flights")

	return nil
}

func Disconnect() error {
	if err := db.Client().Disconnect(customContext.EmptyCtx); err != nil {
		return errors.New("Error disconnecting MongDB" + err.Error())
	}
	return nil
}

func GetCollection(coll string) *mongo.Collection {
	return db.Collection(coll)
}

func GetById[T any](coll string, id string, opts ...*options.FindOneOptions) (T, error) {
	var result T
	filter := bson.D{{Key: "_id", Value: id}}
	err := GetCollection(coll).FindOne(customContext.EmptyCtx, filter, opts...).Decode(&result)
	return result, err
}

func InsertToCollection[T any](coll string, doc T, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return GetCollection(coll).InsertOne(customContext.EmptyCtx, doc, opts...)
}

func InsertBulkToCollection[T any](coll string, docs []T, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	// to pass type check
	_docs := collection.Map[T, any](docs, func(doc T) any {
		return doc
	})
	return GetCollection(coll).InsertMany(customContext.EmptyCtx, _docs, opts...)
}

func UpdateById(coll string, id string, update any, options ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	filter := bson.D{{Key: "_id", Value: id}}
	return GetCollection(coll).UpdateOne(customContext.EmptyCtx, filter, update, options...)
}

func Find[T any](coll string, filter any) (results []T, err error) {
	cursor, err := GetCollection(coll).Find(customContext.EmptyCtx, filter)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if err = cursor.All(customContext.EmptyCtx, &results); err != nil {
		log.Println(err)
		return nil, err
	}
	return results, err
}
