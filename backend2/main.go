package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	pb2 "github.com/theemadnes/golang-grpc-trace-demo/bingbong"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50052, "The server port")
)

type server struct {
	pb2.UnimplementedBingBongServer
}

func (s *server) GetPong(ctx context.Context, in *pb2.Bing) (*pb2.Bong, error) {
	log.Printf("Received ping: %v", in.Bing)
	currentTime := time.Now()
	return &pb2.Bong{Bong: "Bong @ " + currentTime.String()}, nil
}

func main() {
	flag.Parse()
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		log.Fatalf("texporter.NewExporter: %v", err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	defer tp.ForceFlush(ctx) // flushes any pending spans
	otel.SetTracerProvider(tp)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	)

	pb2.RegisterBingBongServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
