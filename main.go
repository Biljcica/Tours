package main

import (
	"context"
	handlers "database-example/handler"
	tourspb "database-example/proto/tours"
	"database-example/repo"
	"database-example/service"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// kreiranje repozitorijuma, servisa i handlera
	tourRepo := &repo.TourRepository{Collection: collection}
	tourService := &service.TourService{TourRepo: tourRepo}
	tourHandler := handlers.NewToursHandler(tourService)

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
	tourspb.RegisterToursServiceServer(grpcServer, tourHandler)
	reflection.Register(grpcServer)

	// start gRPC servera
	go func() {
		logger.Println("Starting gRPC server on", addr)
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
