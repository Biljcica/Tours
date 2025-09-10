package main

import (
	"database-example/handler"
	"database-example/model"
	"database-example/repo"
	"database-example/service"
	"log"
	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"net/http"
    "github.com/gorilla/mux"
	"context"
	"fmt"
	"os"
	"os/signal"
    "syscall"
)

func initDB() *mongo.Database {
	mongoUri := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI("mongodb://" + mongoUri + "/?connect=direct")

	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	database := client.Database("tours")
	collection := database.Collection("tours")
	fmt.Println(collection.Name())

	tour := model.Tour{
		ID:          "aec7e123-233d-4a09-a289-75308ea5b7e6",
		AuthorID:    "author-123",
		Name:        "Planinarska tura",
		Description: "Tura kroz planine sa lepim vidicima",
		Difficulty:  "srednja",
		Tags:        []string{"planine", "aktivnost", "priroda"},
		Status:      "draft",
		Price:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	insertResult, err := collection.InsertOne(context.TODO(), tour)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	return database
}

func startServer(tourHandler *handler.TourHandler) {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/tours", tourHandler.CreateTour).Methods("POST")
	router.HandleFunc("/tours/{id}", tourHandler.GetTour).Methods("GET")
	router.HandleFunc("/authors/{authorID}/tours", tourHandler.GetToursByAuthor).Methods("GET")
	router.HandleFunc("/tours/{id}/publish", tourHandler.PublishTour).Methods("PUT")
	router.HandleFunc("/tours", tourHandler.GetAllTours).Methods("GET")

	println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}


func main() {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	database := initDB()
	if database == nil {
		print("FAILED TO CONNECT TO DB")
		return
	}

	collection := database.Collection("tours")
	repo := &repo.TourRepository{Collection: collection}
	service := &service.TourService{TourRepo: repo}
	handler := &handler.TourHandler{TourService: service}

	startServer(handler)
}
