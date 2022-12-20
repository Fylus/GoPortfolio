package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"strconv"
	"time"
)

type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name,omitempty"`
	Description string             `bson:"description,omitempty"`
	Date        int                `bson:"date,omitempty"`
}

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

	entries := readCSV("resources/projects.csv")

	_, err = myProjects.InsertMany(ctx, entries)
	if err != nil {
		log.Printf("could not insert entries %v: %v", entries, err)
	}
}
func readCSV(filePath string) []interface{} {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Cannot open file: "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	entries := make([]interface{}, len(records)-1)
	for i, record := range records {
		if i == 0 {
			continue
		}
		date, _ := strconv.Atoi(record[2])
		project := Project{Name: record[0], Description: record[1], Date: date}
		marshal, err := bson.Marshal(project)
		if err != nil {
			return nil
		}
		fmt.Println(project)
		entries[i-1] = marshal
	}
	if err != nil {
		log.Fatal("Cannot read file: "+filePath, err)
	}
	return entries
}
