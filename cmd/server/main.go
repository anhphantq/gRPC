package main

import (
	"flag"
	"fmt"
	"gRPC/pb"
	"gRPC/service"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main(){
	port := flag.Int("port", 0, "the server port")
	flag.Parse()

	log.Printf("start the server on port %d", *port)

	laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())
	grpcServer := grpc.NewServer()

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)

	listener, err := net.Listen("tcp", address)
	if err != nil{
		log.Fatal("cannot start server")
	}

	err = grpcServer.Serve(listener)
	if err != nil{
		log.Fatal(err)
	}
}