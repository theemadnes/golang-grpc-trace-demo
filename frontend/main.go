package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	pb "github.com/theemadnes/golang-grpc-trace-demo/pingpong"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {

	flag.Parse()

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		//io.WriteString(w, "Hello, world!\n")
		ctx := context.Background()
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		exporter, err := texporter.New(texporter.WithProjectID(projectID))
		if err != nil {
			log.Fatalf("texporter.NewExporter: %v", err)
		}
		tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
		defer tp.ForceFlush(ctx) // flushes any pending spans
		otel.SetTracerProvider(tp)
		conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewPingPongClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		currentTime := time.Now()
		r, err := c.GetPong(ctx, &pb.Ping{Ping: "Ping @ " + currentTime.String()})
		if err != nil {
			log.Fatalf("could not ping: %v", err)
		}
		log.Printf("Got pong: %s", r.Pong)
		fmt.Fprintf(w, "Got pong: %s\n", r.Pong)
	}

	http.HandleFunc("/ping", helloHandler)
	log.Println("Listing for requests at http://localhost:8000/ping")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
