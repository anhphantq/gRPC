package service

import (
	"bytes"
	"context"
	"errors"
	"gRPC/pb"
	"io"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxImageSize = 1 << 20

type LaptopServer struct {
	laptopStore LaptopStore
	imageStore ImageStore
}

func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore) *LaptopServer {
	return &LaptopServer{
		laptopStore: laptopStore,
		imageStore: imageStore,
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

	err := server.laptopStore.Save(laptop)
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

func (server *LaptopServer) SearchLaptop(in *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error{
	filter := in.GetFilter()
	log.Printf("Filter received %v", filter)

	err := server.laptopStore.Search(filter, func(laptop *pb.Laptop) error {
		res := &pb.SearchLaptopResponse{
			Laptop: laptop,
		}

		err := stream.Send(res)
		if err != nil {
			return err
		}

		log.Printf("Founed with id: %s", laptop.Id)
		return nil
	})

	if (err != nil){
		return err
	}

	return nil
}

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error{
	req, err := stream.Recv()
	if err != nil{
		log.Print("cannot receive info", err)
		return status.Error(codes.Unknown, "cannot receive info")
	}
	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("received %s %s",laptopID, imageType)

	laptop, err := server.laptopStore.Find(laptopID)
	if err != nil {
		log.Print("not founded laptop")
		return status.Error(codes.Unknown, "laptop not founded")
	}

	if laptop == nil{
		log.Print("laptop null")
		return status.Error(codes.Unknown, "laptop null")
	}

	imageData := bytes.Buffer{}
	imageSize := 0
	
	for {
		log.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF{
			log.Print("no more data")
			break
		}
		if err != nil{
			log.Print("loop receiving failed")
			return status.Error(codes.Unknown, "loop receiving failed")
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		imageSize += size
		if imageSize > maxImageSize{
			return status.Error(codes.InvalidArgument, "too big")
		}

		_, err = imageData.Write(chunk)
		if (err != nil){
			return status.Error(codes.InvalidArgument, "cannot write chunk to buffer")
		}
	}

	imageID, err := server.imageStore.Save(laptopID, imageType, imageData)
	if err != nil{
		log.Print("cannot save image")
		return status.Error(codes.InvalidArgument, "save failed")
	}

	res := &pb.UploadImageResponse{
		Id: imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil{
		log.Print("cannot send res")
		return status.Error(codes.InvalidArgument, "can not send res")
	}

	log.Printf("saved image")
	return nil
}