package main

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB(collections *Collections) {
	clientOptions := options.Client().ApplyURI(os.Getenv("DB_CONNECTION_URI"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database(os.Getenv("DB_NAME"))
	collections.Releases = db.Collection(os.Getenv("RELEASES_COLLECTION"))
	collections.Updates = db.Collection(os.Getenv("UPDATES_COLLECTION"))
}
