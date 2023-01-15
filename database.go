/*
 This file contains all database functions for the webserver and webbuilder
 It uses the mongodb driver to connect to the database
*/
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
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	timeout      = 5 * time.Second
	projects     = "projects"
	otherskills  = "otherskills"
	education    = "education"
	software     = "software"
	language     = "language"
	proglanguage = "proglanguage"
)

var (
	client *mongo.Client
	mux    sync.Mutex
)

// getDatabase returns the database in a thread safe way in a singleton pattern
func getDatabase(ctx context.Context) *mongo.Database {
	mux.Lock()
	defer mux.Unlock()
	// singleton client
	if client == nil {
		var err error
		name := os.Getenv("DB_NAME")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASS")
		port := os.Getenv("DB_PORT")
		host := "mongodb://" + user + ":" + pass + "@" + name + ":" + port
		log.Println("connecting to database: ", host)
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(host))
		if err != nil {
			log.Fatalln("could not connect to database: ", err)
		}
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			log.Fatalln("could not ping database: ", err)
		}
	}
	return client.Database("mydb")
}

// buildDatabase  reads all json files for each category and inserts them into the database
// this is used in main.go to build the database every time the app starts
func buildDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	readJSONFileToDatabase(ctx, "json/projects.json", projects)
	readJSONFileToDatabase(ctx, "json/otherskills.json", otherskills)
	readJSONFileToDatabase(ctx, "json/education.json", education)
	readJSONFileToDatabase(ctx, "json/software.json", software)
	readJSONFileToDatabase(ctx, "json/language.json", language)
	readJSONFileToDatabase(ctx, "json/proglanguage.json", proglanguage)
}

// readJSONFileToDatabase reads a json file and inserts it into the database
func readJSONFileToDatabase(ctx context.Context, filename string, collection string) {
	myCollection := getDatabase(ctx).Collection(collection)
	err := myCollection.Drop(ctx)
	if err != nil {
		log.Println("could not drop collection ", err)
	}
	content := readJSON(filename)
	_, err = myCollection.InsertMany(ctx, content)
	if err != nil {
		log.Fatalln("could not insert entries: ", err)
	} else {
		log.Printf("inserted entries: %v to %v \n", len(content), collection)
	}
}

// readJSON reads a json file, converts returns it as a database suitable slice of bson.M
func readJSON(filePath string) bson.A {
	jsonFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalln("could not read file: ", err)
	}
	var bsonData bson.A
	err = json.Unmarshal(jsonFile, &bsonData)
	if err != nil {
		log.Fatalln("could not unmarshal json: ", err)
	}
	return bsonData
}

// getProjectFromDatabase returns one project from the database as a ProductPage with a http status code if the project was found
func getProjectFromDatabase(id string) (ProductPage, int) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	database := getDatabase(ctx)

	myProjects := database.Collection("projects")

	result := myProjects.FindOne(ctx, bson.M{"id": id})
	var resultMap bson.M
	err := result.Decode(&resultMap)
	if result.Err() != nil || err != nil {
		log.Println("could not decode result: ", err)
		return ProductPage{
			Page: Page{
				Title: "Project Not Found",
				HTML:  getHTML(),
			},
			Type:      "project",
			Noproduct: true,
		}, http.StatusNotFound
	}
	// TableContent is a map of all skills used in the project
	tablemap := make(map[string][]bson.M)
	tablemap["Software"] = getTableContent(resultMap, software)
	tablemap["Skills"] = getTableContent(resultMap, "skills")
	return ProductPage{
		Page: Page{
			Title: resultMap["name"].(string),
			CSS:   "productpage",
			HTML:  getHTML(),
		},
		Image:       resultMap["img"].(string),
		Description: resultMap["long"].(string),
		Table:       tablemap,
		Type:        "project",
	}, http.StatusOK
}

// getToolFromDatabase returns one tool from the database as a ProductPage with a http status code if the tool was found
func getToolFromDatabase(nameID string) (ProductPage, int) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	database := getDatabase(ctx)

	myProjects := database.Collection(software)

	result := myProjects.FindOne(ctx, bson.M{"id": nameID})
	var resultMap bson.M
	err := result.Decode(&resultMap)
	if result.Err() != nil || err != nil {
		log.Println("could not decode result: ", err)
		return ProductPage{
			Page: Page{
				Title: "Tool Not Found",
				HTML:  getHTML(),
			},
			Type:      "tool",
			Noproduct: true,
		}, http.StatusNotFound
	}

	// TableContent is a map of all information about the tool
	tablemap := make(map[string][]bson.M)
	tablemap["Company"] = []bson.M{{"name": resultMap["company"].(string)}}
	tablemap["Projects"] = getProjectsFromSoftware(nameID)
	return ProductPage{
		Page: Page{
			Title: resultMap["name"].(string),
			CSS:   "productpage",
			HTML:  getHTML(),
		},
		Image:       resultMap["img"].(string),
		Description: resultMap["description"].(string),
		Table:       tablemap,
		External:    resultMap["externallink"].(string),
		Type:        "tool",
	}, http.StatusOK
}

// getProjectsFromSoftware returns all projects that use a specific tool
func getProjectsFromSoftware(id string) []bson.M {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	database := getDatabase(ctx)
	myProjects := database.Collection(projects)
	cursor, err := myProjects.Find(ctx, bson.M{"software.id": id})
	if err != nil {
		log.Println("could not find projects: ", err)
	}
	var results []bson.M
	// iterate over all projects to change link to project
	for cursor.Next(ctx) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Println("could not decode result: ", err)
		}
		result["link"] = "project/" + result["id"].(string)
		results = append(results, result)
	}
	if err != nil {
		log.Println("could not decode results: ", err)
	}
	return results
}

// getTableContent returns all content that is shown in the table of a product page
func getTableContent(resultMap bson.M, category string) []bson.M {
	var tableContent []bson.M
	for _, content := range resultMap[category].(bson.A) {
		content := content.(bson.M)
		// if the category is software, the link is changed to the tool page
		if category == software {
			content["link"] = "tool/" + content["id"].(string)
		}
		tableContent = append(tableContent, content)
	}
	return tableContent
}

// getAllProjectsOfCollection returns all projects in the database from a specific collection as mongo cursor
func getAllProjectsOfCollection(collection string) (*mongo.Cursor, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	database := getDatabase(ctx)
	myProjects := database.Collection(collection)
	result, err := myProjects.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("could not find %s: %v \n", collection, err)
	}
	return result, ctx
}

// getAllProjects returns all projects in the database as a map of their categories in a bson.M object
func getAllProjectsInCategories() map[string][]bson.M {
	result, _ := getAllProjectsOfCollection(projects)
	categories := make(map[string][]bson.M)
	for result.Next(context.TODO()) {
		var project bson.M
		err := result.Decode(&project)
		if err != nil {
			log.Fatalln("could not decode project: ", err)
		}
		if project["categories"] != nil {
			//change project date from Y-M-D to YYYY
			project["date"] = project["date"].(string)[:4]
			//check if image file exists
			project["img"] = checkImage(project["img"].(string))
			// put project in the right category
			for _, category := range project["categories"].(bson.A) {
				name := category.(bson.M)["name"].(string)
				categories[name] = append(categories[name], project)
			}
		} else {
			categories["other"] = append(categories["other"], project)
		}
	}
	return categories
}

// getAllIDs returns all database nameIDs of a category projects as a string array
func getAllIDs(category string) []string {
	result, _ := getAllProjectsOfCollection(category)
	var ids []string
	for result.Next(context.TODO()) {
		var content bson.M
		err := result.Decode(&content)
		if err != nil {
			log.Fatal("could not decode content: ", err)
		}
		ids = append(ids, content["id"].(string))
	}
	return ids
}

// checkImage checks if an image file exists and returns the path to the image or a default image
func checkImage(imagePath string) string {
	// lores image is the image specially made for the home page
	imagepath := "./static/images/lores/" + imagePath
	if _, err := os.Stat(imagepath); os.IsNotExist(err) {
		// if the lores image does not exist, the normal image is used which is maybe too big for the home page
		imagepath = "./static/images/hires/" + imagePath
		if _, err := os.Stat(imagepath); os.IsNotExist(err) {
			// if the normal image does not exist, a default image is used
			imagepath = "./static/images/lores/coming-soon.png"
		}
	}
	return imagepath
}

// getEductionFromDatabase returns all education from the database as a slice of bson.M objects
func getEducationFromDatabase() []bson.M {
	return getSkillFromDatabase(education)
}

// getProgLangFromDatabase returns all programming languages from the database as a slice of bson.M objects
func getProgLangFromDatabase() []bson.M {
	return getSkillFromDatabase(proglanguage)
}

// getSoftwareFromDatabase returns all software from the database as a slice of bson.M objects
func getSoftwareFromDatabase() []bson.M {
	return getSkillFromDatabase(software)
}

// getOtherSkillsFromDatabase returns all other skills from the database as a slice of bson.M objects
func getOtherSkillsFromDatabase() []bson.M {
	return getSkillFromDatabase(otherskills)
}

// getLanguageFromDatabase returns all languages from the database as a slice of bson.M objects
func getLanguageFromDatabase() []bson.M {
	return getSkillFromDatabase(language)
}

// getSkillFromDatabase returns all skills from a specific category from the database as a slice of bson.M objects
func getSkillFromDatabase(col string) []bson.M {
	result, ctx := getAllProjectsOfCollection(col)
	var content []bson.M
	err := result.All(ctx, &content)
	if err != nil {
		log.Fatalf("could not decode results: %v", err)
	}
	return content
}
