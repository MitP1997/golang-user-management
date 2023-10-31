package models

import (
	"context"
	"fmt"
	"os"
	reflect "reflect"

	validator10 "github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbClient *mongo.Database
var userCollection *mongo.Collection

var collectionObjectMap = make(map[*mongo.Collection]interface{})

var validator *validator10.Validate

var host, port, db, uri string

func InitMongoConnection() (err error) {
	host = os.Getenv("DATABASE_HOST")
	port = os.Getenv("DATABASE_PORT")
	db = os.Getenv("DATABASE_NAME")
	uri = fmt.Sprintf("mongodb://%s:%s/%s?retryWrites=true&w=majority", host, port, db)

	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.Background(), opts)

	if err != nil {
		return
	}

	// Send a ping to confirm a successful connection
	if err = client.Ping(context.Background(), nil); err != nil {
		return
	}

	initServerVarsPostMongoConnection(client)
	createIndicesForAllCollections()
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return
}

func CloseMongoConnection() (err error) {
	err = dbClient.Client().Disconnect(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Disconnected from MongoDB!")
	return
}

func createIndicesForAllCollections() {
	for coll, obj := range collectionObjectMap {
		indices := getIndices(obj)
		err := createIndices(coll, indices)
		if err != nil {
			panic(err)
		}
	}
}

func initServerVarsPostMongoConnection(client *mongo.Client) {
	dbClient = client.Database(db)

	userCollectionOpts := options.Collection().SetRegistry(objectIdDecodingRegistry())
	userCollection = dbClient.Collection("users", userCollectionOpts)
	collectionObjectMap[userCollection] = &User{}
	validator = validator10.New()
}

func getIndices(obj interface{}) (indices []mongo.IndexModel) {
	value := reflect.ValueOf(obj)
	numFields := reflect.Indirect(value).NumField()
	structType := reflect.Indirect(value).Type()

	for i := 0; i < numFields; i++ {
		field := structType.Field(i)

		index := field.Tag.Get("index")
		if index != "" {
			indexModel := mongo.IndexModel{
				Keys: bson.D{{Key: field.Tag.Get("bson"), Value: 1}},
			}
			if index == "unique" {
				indexModel.Options = options.Index().SetUnique(true)
			}
			indices = append(indices, indexModel)
		}
	}
	return
}

func objectIdDecodingRegistry() *bsoncodec.Registry {
	// we need to generate a new object id and get its type to be used in the decoder
	// as primitive.ObjectID is not allowed to be used directly because of the error "not an expression"
	objectIdType := reflect.TypeOf(primitive.NewObjectID())
	registry := bson.NewRegistry()
	registry.RegisterTypeDecoder(
		objectIdType,
		bsoncodec.ValueDecoderFunc(func(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
			// this is the function when we read the datetime format
			oid, err := vr.ReadObjectID()
			if err != nil {
				return err
			}
			// Convert ObjectID to string and set it in the struct.
			val.SetString(oid.Hex())
			return nil
		}),
	)
	// read the datetime type and convert to integer
	return registry
}

func createIndices(coll *mongo.Collection, indices []mongo.IndexModel) (err error) {
	_, err = coll.Indexes().CreateMany(context.Background(), indices)
	return
}
