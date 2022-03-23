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
	pb "github.com/theemadnes/golang-grpc-trace-demo/pingpong"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	port    = flag.Int("port", 50051, "The server port")
	addr    = flag.String("addr", "localhost:50052", "the address to connect to")
	service = "backend"
)

type server struct {
	pb.UnimplementedPingPongServer
}

func (s *server) GetPong(ctx context.Context, in *pb.Ping) (*pb.Pong, error) {
	log.Printf("Received ping: %v", in.Ping)
	// call backend2

	//ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		log.Fatalf("texporter.NewExporter: %v", err)
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	defer tp.ForceFlush(ctx) // flushes any pending spans
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb2.NewBingBongClient(conn)

	// Contact the server and print out its response.
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	r, err := c.GetBong(ctx, &pb2.Bing{Bing: "Bing @ " + time.Now().String()})
	if err != nil {
		log.Fatalf("could not bing: %v", err)
	}
	log.Printf("Got bong: %s", r.Bong)

	// back to the response for frontend
	return &pb.Pong{Pong: "Pong @ " + time.Now().String()}, nil
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

	pb.RegisterPingPongServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
