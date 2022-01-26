package main

import (
	"context"
	"flag"
	"gRPC/pb"
	"gRPC/sample"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main(){
	serverAddress := flag.String("address","", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddress)
	
	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil{
		log.Fatal("can not dial server")
	}

	laptopClient := pb.NewLaptopServiceClient(conn)

	laptop := sample.NewLaptop()
	req := &pb.CreateLatopRequest{
		Latop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(),req)

	if err != nil{
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists{
			log.Print("laptop already exists")
		} else {
			log.Fatal("can not create laptop", err)
		}
		return 
	}

	log.Printf("laptop created with id %s", res.Id)
}