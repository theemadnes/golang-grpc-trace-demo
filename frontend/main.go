package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	pb "github.com/theemadnes/golang-grpc-trace-demo/pingpong"
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
		conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
