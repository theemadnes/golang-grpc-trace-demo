package main

import (
	"flag"

	pb "github.com/theemadnes/golang-grpc-trace-demo"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedPingPongServer
}
