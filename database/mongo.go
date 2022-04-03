package database

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/AbhayAbe/notzy_backend/models"
	"github.com/AbhayAbe/notzy_backend/statics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var _URI_ string
var Client *mongo.Client
var DB *mongo.Database

func ConfigureMongodb() {

	var err error
	_URI_ = os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(_URI_)
	Client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}
	if err := Client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected and pinged.")

	DB = Client.Database("notzy")

	res := <-initUniqueIndices(models.User{}, nil)
	res = <-initSortIndices(models.Note{}, -1, nil)
	fmt.Println(res.Result)
}

func DisconnectMongodb() {
	if err := Client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func initUniqueIndices(model interface{}, name interface{}) <-chan statics.Result {
	ch := make(chan statics.Result)

	go func() {
		var nameVal string = ""
		defer close(ch)
		switch v := name.(type) {
		case string:
			nameVal = v
		}

		mod := reflect.TypeOf(model)
		var schemaName string
		if len(nameVal) <= 0 {
			schemaName = strings.ToLower(mod.Name()) + "s"
		} else {
			schemaName = nameVal
		}

		for i := 0; i < mod.NumField(); i++ {
			field := mod.Field(i)
			tag := field.Tag.Get(Api.Constants.IsUnique)
			key := field.Tag.Get("json")
			if len(tag) > 0 && len(key) > 0 {
				_, err := DB.Collection(schemaName).Indexes().CreateOne(context.Background(),
					mongo.IndexModel{
						Keys:    bson.M{key: 1},
						Options: options.Index().SetUnique(true),
					})
				if err != nil {
					DisconnectMongodb()
					fmt.Println("Error: ", err)
					ch <- statics.Result{Error: err, Result: 0}
				}
				fmt.Println("Index for", schemaName, "created")
				ch <- statics.Result{Error: nil, Result: 1}
			}
		}
	}()
	return ch
}

func initSortIndices(model interface{}, sortOrder int64, name interface{}) <-chan statics.Result {
	ch := make(chan statics.Result)

	go func() {
		var nameVal string = ""
		defer close(ch)
		switch v := name.(type) {
		case string:
			nameVal = v
		}

		mod := reflect.TypeOf(model)
		var schemaName string
		if len(nameVal) <= 0 {
			schemaName = strings.ToLower(mod.Name()) + "s"
		} else {
			schemaName = nameVal
		}

		for i := 0; i < mod.NumField(); i++ {
			field := mod.Field(i)
			tag := field.Tag.Get(Api.Constants.Sort)
			key := field.Tag.Get("json")
			if len(tag) > 0 && len(key) > 0 {
				_, err := DB.Collection(schemaName).Indexes().CreateOne(context.Background(),
					mongo.IndexModel{
						Keys: bson.M{key: sortOrder},
					})
				if err != nil {
					DisconnectMongodb()
					fmt.Println("Error: ", err)
					ch <- statics.Result{Error: err, Result: 0}
				}
				fmt.Println("Index for", schemaName, "created")
				ch <- statics.Result{Error: nil, Result: 1}
			}
		}
	}()
	return ch
}
