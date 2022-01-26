package service

import (
	"context"
	"errors"
	"gRPC/pb"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LaptopServer struct {
	Store LaptopStore
}

func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{
		Store: store,
	}
}

func (server *LaptopServer) CreateLaptop(ctx context.Context,req *pb.CreateLatopRequest) (*pb.CreateLatopResponse, error){
	laptop := req.GetLatop()
	log.Printf("receive a create-latop req with id: %s", laptop.Id)
	if (len(laptop.Id) > 0){
		_, err := uuid.Parse(laptop.Id)
		if err != nil{
			return nil, status.Errorf(codes.InvalidArgument, "latop ID is not a valid UUID")
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil{
			return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	err := server.Store.Save(laptop)
	if err != nil{
		code := codes.Internal
		if (errors.Is(err, ErrAlreadyExists)){
			code = codes.AlreadyExists
		}

		return nil, status.Errorf(code, "cannot save latop to the store")
	}

	log.Printf("saved laptop with id %s", laptop.Id)

	res := &pb.CreateLatopResponse{
		Id : laptop.Id,
	}

	return res, nil
}