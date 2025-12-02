package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ride4Low/contracts/env"
	"github.com/ride4Low/trip-service/internal/adapter/mongo"
	"github.com/ride4Low/trip-service/internal/adapter/osrm"
	"github.com/ride4Low/trip-service/internal/repository"
	"github.com/ride4Low/trip-service/internal/service"
	"google.golang.org/grpc"

	grpcHandler "github.com/ride4Low/trip-service/internal/handler/grpc"
)

var (
	grpcAddr = env.GetString("GRPC_ADDR", "0.0.0.0:9093")
	osrmURL  = env.GetString("OSRM_URL", "http://router.project-osrm.org/")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	dbCfg := mongo.NewMongoDefaultConfig()
	mongoClient, err := mongo.NewMongoClient(dbCfg)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	repo := repository.NewRepository(mongo.GetDatabase(mongoClient, dbCfg.Database))

	osrmClient := osrm.NewClient(osrmURL)
	svc := service.NewService(osrmClient, repo)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcHandler.NewHandler(grpcServer, svc)

	go func() {
		log.Printf("Server listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down trip service")
	grpcServer.GracefulStop()

}
