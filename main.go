package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	handlers "database-example/handler"
	imagepb "database-example/proto/image"
	positionpb "database-example/proto/position"
	reviewpb "database-example/proto/review" // DODAJTE OVO
	tourspb "database-example/proto/tours"
	"database-example/repo"
	"database-example/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func initDB() *mongo.Database {
	mongoUri := os.Getenv("MONGODB_URI")
	if mongoUri == "" {
		log.Fatal("MONGODB_URI environment variable not set")
	}

	clientOptions := options.Client().ApplyURI(mongoUri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatal("MongoDB ping error:", err)
	}

	fmt.Println("Connected to MongoDB!")
	return client.Database("toursdb")
}

func main() {
	logger := log.New(os.Stdout, "[tours] ", log.LstdFlags)

	// inicijalizacija baze
	db := initDB()
	collection := db.Collection("tours")
	positionCollection := db.Collection("positions")
	reviewCollection := db.Collection("reviews") // DODAJTE OVO

	// kreiranje repozitorijuma, servisa i handlera

	// za Tour
	tourRepo := &repo.TourRepository{Collection: collection}
	tourService := &service.TourService{TourRepo: tourRepo}

	// --- Inicijalizacija TourExecution ---
	tourExecutionCollection := db.Collection("tour_executions") // nova kolekcija u MongoDB
	tourExecutionRepo := repo.NewTourExecutionRepository(tourExecutionCollection)
	tourExecutionService := service.NewTourExecutionService(tourExecutionRepo)

	// prosleđivanje u ToursHandler
	tourHandler := handlers.NewToursHandler(tourService, tourExecutionService)

	// za Position
	positionRepo := repo.NewPositionRepository(positionCollection)
	positionService := service.NewPositionService(positionRepo)
	positionHandler := handlers.NewPositionHandler(positionService)

	// za Review - DODAJTE OVO
	reviewRepo := repo.NewReviewRepository(reviewCollection)
	reviewService := service.NewReviewService(reviewRepo)
	reviewHandler := handlers.NewReviewHandler(reviewService)

	// --- Folder za upload slika ---
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			logger.Fatal("Failed to create uploads directory:", err)
		}
	}

	// Kreiranje ImageHandlera
	tourImageHandler := handlers.NewImageHandler(uploadDir)

	// adresa gRPC servera
	addr := os.Getenv("TOURS_SERVICE_ADDRESS")
	if addr == "" {
		addr = ":8084"
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()

	// Registracija svih servisa
	tourspb.RegisterToursServiceServer(grpcServer, tourHandler)
	imagepb.RegisterImageServiceServer(grpcServer, tourImageHandler)
	positionpb.RegisterPositionServiceServer(grpcServer, positionHandler)
	reviewpb.RegisterReviewServiceServer(grpcServer, reviewHandler) // DODAJTE OVO

	reflection.Register(grpcServer)

	// start gRPC servera
	go func() {
		logger.Println("Starting gRPC server on", addr)
		logger.Println("Registered services: Tours, Image, Position, Review") // DODAJTE Review
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("gRPC server error:", err)
		}
	}()

	// čekanje na prekid
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh

	logger.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	time.Sleep(1 * time.Second) // opcionalno da se završe aktuelni pozivi
}
