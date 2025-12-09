package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ride4Low/contracts/env"
	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/contracts/pkg/otel"
	amqpClient "github.com/ride4Low/contracts/pkg/rabbitmq"
	"github.com/ride4Low/trip-service/internal/adapter/mongo"
	"github.com/ride4Low/trip-service/internal/adapter/osrm"
	"github.com/ride4Low/trip-service/internal/events/rabbitmq"
	"github.com/ride4Low/trip-service/internal/repository"
	"github.com/ride4Low/trip-service/internal/service"
	"google.golang.org/grpc"

	grpcHandler "github.com/ride4Low/trip-service/internal/handler/grpc"
)

var (
	grpcAddr    = ":9093"
	osrmURL     = env.GetString("OSRM_URL", "http://router.project-osrm.org/")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672")
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

	otelCfg := otel.DefaultConfig("trip-service")
	otelCfg.JaegerEndpoint = env.GetString("JAEGER_ENDPOINT", "localhost:4317")

	otelProvider, err := otel.Setup(ctx, otelCfg)
	if err != nil {
		log.Fatalf("failed to setup otel: %v", err)
	}
	defer func() {
		if err := otelProvider.Shutdown(context.Background()); err != nil {
			log.Printf("failed to shutdown otel: %v", err)
		}
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

	rmq, err := amqpClient.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rmq.Close()

	publisher := amqpClient.NewPublisher(rmq)
	tripPublisher := rabbitmq.NewTripEventPublisher(publisher)

	driverEventHandler := rabbitmq.NewDriverEventHandler(publisher, svc)
	driverConsumer := amqpClient.NewConsumer(rmq, driverEventHandler)
	if err := driverConsumer.Consume(ctx, events.DriverTripResponseQueue); err != nil {
		log.Fatal(err)
	}

	paymentEventHandler := rabbitmq.NewPaymentEventHandler(svc)
	paymentConsumer := amqpClient.NewConsumer(rmq, paymentEventHandler)
	if err := paymentConsumer.Consume(ctx, events.NotifyPaymentSuccessQueue); err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(otel.ServerOptions()...)
	grpcHandler.NewHandler(grpcServer, svc, tripPublisher)

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
