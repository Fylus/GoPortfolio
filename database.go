package main

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io/ioutil"
	"log"
	"time"
)

func buildDatabase() {
	//connect with mongodb
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opt := options.Client().ApplyURI("mongodb://root:rootpassword@localhost:27017")
	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	myProjects := client.Database("mydb").Collection("projects")

	err = myProjects.Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}

	projects := readJSON("resources/projects.json")
	_, err = myProjects.InsertMany(ctx, projects)
	if err != nil {
		log.Fatalf("could not insert entries %v ", err)
	} else {
		log.Printf("inserted entries: %v ", len(projects))
	}
}

func readJSON(filePath string) bson.A {
	jsonFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("could not read file: %v", err)
	}

	var bsonData bson.A
	err = json.Unmarshal(jsonFile, &bsonData)
	if err != nil {
		log.Fatalf("could not unmarshal json: %v", err)
	}
	return bsonData
}
