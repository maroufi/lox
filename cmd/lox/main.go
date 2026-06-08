package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/maroufi/lox/gen/go/proto/lox/v1"
	"github.com/maroufi/lox/internal/api"
	"github.com/maroufi/lox/internal/lock"
	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", ":50051", "gRPC listen address")
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen on %s: %v", *addr, err)
	}

	server := grpc.NewServer()
	loxv1.RegisterLockServiceServer(server, api.NewServer(lock.NewManager(nil)))

	fmt.Printf("lox gRPC server listening on %s\n", *addr)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve gRPC: %v", err)
	}
}
