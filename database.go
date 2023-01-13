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
	timeout      = 10 * time.Second
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
	// a slice of all collections
	skills = []string{otherskills, education, software, language, proglanguage}
)

func getDatabase(ctx context.Context) *mongo.Database {
	mux.Lock()
	defer mux.Unlock()
	if client == nil {
		var err error
		client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:rootpassword@localhost:27017"))
		if err != nil {
			log.Fatal("could not connect to database: ", err)
		}
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			log.Fatal("could not ping database: ", err)
		}
	}
	return client.Database("mydb")
}

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

func readJSONFileToDatabase(ctx context.Context, filename string, collection string) {
	myCollection := getDatabase(ctx).Collection(collection)
	err := myCollection.Drop(ctx)
	if err != nil {
		log.Printf("could not drop collection %v", err)
	}
	content := readJSON(filename)
	_, err = myCollection.InsertMany(ctx, content)
	if err != nil {
		log.Fatalf("could not insert entries %v ", err)
	} else {
		log.Printf("inserted entries: %v ", len(content))
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

func getProjectFromDatabase(id string) (Page, int) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	database := getDatabase(ctx)

	myProjects := database.Collection("projects")

	result := myProjects.FindOne(ctx, bson.M{"id": id})
	var resultMap bson.M
	err := result.Decode(&resultMap)
	if result.Err() != nil || err != nil {
		log.Printf("could not decode result: %v", err)
		return Page{Title: "Project Not Found", Product: Product{Type: "project", Noproduct: true}}, http.StatusNotFound
	}

	tablemap := make(map[string][]bson.M)
	tablemap["Software"] = getTableContent(resultMap, software)
	tablemap["Skills"] = getTableContent(resultMap, "skills")
	return Page{
		Title: resultMap["name"].(string),
		CSS:   "productpage",
		Product: Product{
			Image:       resultMap["img"].(string),
			Description: resultMap["long"].(string),
			Table:       tablemap,
			Type:        "project",
		},
	}, http.StatusOK
}

func getToolFromDatabase(nameID string) (Page, int) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	database := getDatabase(ctx)

	myProjects := database.Collection("software")

	result := myProjects.FindOne(ctx, bson.M{"id": nameID})
	var resultMap bson.M
	err := result.Decode(&resultMap)
	if result.Err() != nil || err != nil {
		log.Printf("could not decode result: %v", err)
		return Page{Title: "Tool Not Found", Product: Product{Type: "tool", Noproduct: true}}, http.StatusNotFound
	}

	tablemap := make(map[string][]bson.M)

	// first entry of company
	tablemap["Company"] = []bson.M{{"name": resultMap["company"].(string)}}
	tablemap["Projects"] = getProjectsFromSoftware(nameID)

	return Page{
		Title: resultMap["name"].(string),
		CSS:   "productpage",
		Product: Product{
			Image:       resultMap["img"].(string),
			Description: resultMap["description"].(string),
			Table:       tablemap,
			External:    resultMap["externallink"].(string),
			Type:        "tool",
		},
	}, http.StatusOK
}

func getProjectsFromSoftware(id string) []bson.M {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	database := getDatabase(ctx)

	myProjects := database.Collection("projects")

	// find all projects where software name is id
	cursor, err := myProjects.Find(ctx, bson.M{"software.id": id})

	if err != nil {
		log.Printf("could not find projects: %v", err)
	}
	var results []bson.M
	for cursor.Next(ctx) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Printf("could not decode result: %v", err)
		}
		result["link"] = "project/" + result["id"].(string)
		results = append(results, result)
	}
	if err != nil {
		log.Printf("could not decode results: %v", err)
	}
	return results
}

func getTableContent(resultMap bson.M, category string) []bson.M {
	var tableContent []bson.M
	for _, skill := range resultMap[category].(bson.A) {
		skill := skill.(bson.M)
		if category == software {
			skill["link"] = "tool/" + skill["id"].(string)
		}
		tableContent = append(tableContent, skill)
	}
	return tableContent
}

func getAllFromCollection(collection string) (*mongo.Cursor, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	database := getDatabase(ctx)

	myProjects := database.Collection(collection)
	result, err := myProjects.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("could not find %s: %v", collection, err)
	}
	return result, ctx
}

func getAllProjectsInCategories() map[string][]bson.M {

	result, _ := getAllFromCollection(projects)
	categories := make(map[string][]bson.M)

	for result.Next(context.TODO()) {
		var project bson.M
		err := result.Decode(&project)
		if err != nil {
			log.Fatal(err)
		}
		if project["categories"] != nil {
			//change project date from Y-M-D to YYYY
			project["date"] = project["date"].(string)[:4]
			//check if image file exists
			project["img"] = checkImage(project["img"].(string))
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

func checkImage(imagePath string) string {
	imagepath := "./static/images/lores/" + imagePath
	if _, err := os.Stat(imagepath); os.IsNotExist(err) {
		imagepath = "./static/images/hires/" + imagePath
		if _, err := os.Stat(imagepath); os.IsNotExist(err) {
			imagepath = "./static/images/lores/coming-soon.png"
		}
	}
	return imagepath
}

func getSkillsFromDatabase() map[string][]bson.M {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	database := getDatabase(ctx)

	m := make(map[string][]bson.M)

	for _, collection := range skills {
		myCollection := database.Collection(collection)
		result, err := myCollection.Find(ctx, bson.M{})
		if err != nil {
			log.Fatalf("could not find projects: %v", err)
		}
		var results []bson.M
		err = result.All(ctx, &results)
		if err != nil {
			log.Fatalf("could not decode results: %v", err)
		}
		m[collection] = results
	}
	return m
}

func getEducationFromDatabase() []bson.M {
	return getSkillFromDatabase(education)
}

func getProgLangFromDatabase() []bson.M {
	return getSkillFromDatabase(proglanguage)
}

func getSoftwareFromDatabase() []bson.M {
	return getSkillFromDatabase(software)
}

func getOtherSkillsFromDatabase() []bson.M {
	return getSkillFromDatabase(otherskills)
}
func getLanguageFromDatabase() []bson.M {
	return getSkillFromDatabase(language)
}

func getSkillFromDatabase(col string) []bson.M {
	result, ctx := getAllFromCollection(col)
	var content []bson.M
	err := result.All(ctx, &content)
	if err != nil {
		log.Fatalf("could not decode results: %v", err)
	}
	return content
}
