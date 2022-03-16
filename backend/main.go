package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/theemadnes/golang-grpc-trace-demo/pingpong"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedPingPongServer
}

func (s *server) GetPong(ctx context.Context, in *pb.Ping) (*pb.Pong, error) {
	log.Printf("Received ping: %v", in.Ping)
	currentTime := time.Now()
	return &pb.Pong{Pong: "Pong @ " + currentTime.String()}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPingPongServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
