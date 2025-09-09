package db

import (
	"context"
	"strconv"
	"sync"

	"github.com/EutropicAI/FinalRip/module/config"
	"github.com/EutropicAI/FinalRip/module/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB   *mongo.Database
	once sync.Once
)

func Init() {
	once.Do(func() {
		initialize()
	})
}

func initialize() {
	credential := options.Credential{
		Username: config.DBConfig.Username,
		Password: config.DBConfig.Password,
	}
	applyURI := "mongodb://" + config.DBConfig.Host + ":" + strconv.Itoa(config.DBConfig.Port)
	log.Logger.Info("Connecting to MongoDB: " + applyURI)
	clientOpts := options.Client().ApplyURI(applyURI).SetAuth(credential)

	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Logger.Error("Failed to connect to MongoDB: " + err.Error())
		panic(err)
	}

	// Send a ping to confirm a successful connection
	var result bson.M
	if err = client.Database(config.DBConfig.Database).RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil { //nolint
		log.Logger.Error("Failed to connect to MongoDB: " + err.Error())
		panic(err)
	}

	DB = client.Database(config.DBConfig.Database)
}
