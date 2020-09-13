package demo2Mongo

import (
	"context"
	"fmt"

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
