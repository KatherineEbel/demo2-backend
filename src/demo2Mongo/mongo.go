package demo2Mongo

import (
	"context"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go_systems/src/demo2Config"
	"go_systems/src/demo2Users"
	"go_systems/src/demo2Utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type key string

const (
	HostKey     = key("hostKey")
	UsernameKey = key("usernameKey")
	PasswordKey = key("passwordKey")
	DatabaseKey = key("databaseKey")
)

var (
	ctx    context.Context
	client *mongo.Client
)

func configureMongoClient() {
	ctx = context.WithValue(ctx, HostKey, demo2Config.MongoHost)
	ctx = context.WithValue(ctx, UsernameKey, demo2Config.MongoUser)
	ctx = context.WithValue(ctx, PasswordKey, demo2Config.MongoPass)
	ctx = context.WithValue(ctx, DatabaseKey, demo2Config.MongoDB)
	uri := fmt.Sprintf("mongodb://%s:%s@%s/%s",
		ctx.Value(UsernameKey).(string),
		ctx.Value(PasswordKey).(string),
		ctx.Value(HostKey).(string),
		ctx.Value(DatabaseKey).(string))
	clientOptions := options.Client().ApplyURI(uri)
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("error connecting to mongo: ", err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		fmt.Println("error pinging to mongo ", err)
	}
	fmt.Println("Mongo connected...")
}

func setContext() {
	ctx = context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = context.WithValue(ctx, HostKey, demo2Config.MongoHost)
	ctx = context.WithValue(ctx, UsernameKey, demo2Config.MongoUser)
	ctx = context.WithValue(ctx, PasswordKey, demo2Config.MongoPass)
	ctx = context.WithValue(ctx, DatabaseKey, demo2Config.MongoDB)
}

func init() {
	setContext()
	configureMongoClient()
}

func AuthenticateUser(e []byte, p []byte) (*demo2Users.AuthUser, error) {
	var doc demo2Users.AuthUser
	collection := client.Database("api").Collection("users")
	filter := bson.D{{"$or", bson.A{bson.D{{"email", string(e)}}, bson.D{{"alias", string(e)}}}}}
	if err := collection.FindOne(ctx, filter).Decode(&doc); err != nil {
		return nil, err
	}
	ok, err := demo2Utils.IsValid(p, []byte(doc.Password))
	if err != nil && ok {
		return nil, err
	}
	return &doc, nil
}

func CheckDocumentExists(db string, col string, key string, value string) bool {
	var doc interface{}
	filter := bson.D{{key, value}}
	collection := client.Database(db).Collection(col)
	err := collection.FindOne(ctx, filter).Decode(&doc)
	return err != nil && doc == nil
}

func InsertDocument(db string, col string, doc []byte) (string, error) {
	collection := client.Database(db).Collection(col)
	var iDoc map[string]interface{}
	err := json.Unmarshal(doc, &iDoc)
	if err != nil {
		return "noop", err
	}
	if _, ok := iDoc["_id"]; ok {
		fmt.Println("Doc isn't new")
		return "noop", nil
	}
	result, err := collection.InsertOne(ctx, &iDoc)
	if err != nil {
		return "noop", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}
