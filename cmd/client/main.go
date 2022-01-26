package main

import (
	"bufio"
	"context"
	"flag"
	"gRPC/pb"
	"gRPC/sample"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop){
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

func UploadImage(laptopClient pb.LaptopServiceClient, laptopID string, imagePath string){
	file, err := os.Open(imagePath)
	if err != nil{
		log.Fatal("cannot open image file")
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.UploadImage(ctx)
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId: laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if (err != nil){
		log.Fatal("cannot send header")
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for{
		n, err := reader.Read(buffer)
		if err == io.EOF{
			break
		}
		if err != nil{
			log.Fatal("connot read chunk to buffer")
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil{
			log.Fatal("send chunk to server ", err, stream.RecvMsg(nil))
		}

	}

	res, err := stream.CloseAndRecv()
		if err != nil{
			log.Fatal("cannot received res")
		}

		log.Print("image uploaded:", res.GetId(), res.GetSize())
}

func testSearchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter){
	log.Print("search filter: ", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}

	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cannot search laptop", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF{
			return
		}

		if err != nil{
			log.Fatal("cannot receive response")
		}

		laptop := res.GetLaptop()
		log.Println(" - found : ", laptop.GetId())
	}
}

func testUploadImage(laptopClient pb.LaptopServiceClient){
	laptop := sample.NewLaptop()
	CreateLaptop(laptopClient, laptop)
	UploadImage(laptopClient, laptop.GetId(), "tmp/laptop.jpg")
}

func main(){
	serverAddress := flag.String("address","", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddress)
	
	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil{
		log.Fatal("can not dial server")
	}

	laptopClient := pb.NewLaptopServiceClient(conn)

	for i:= 0; i < 10; i++ {
		CreateLaptop(laptopClient, sample.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz: 2.5,
		MinRam: &pb.Memory{Value: 0, Unit: pb.Memory_GIGABYTE},
	}

	testSearchLaptop(laptopClient, filter)
	testUploadImage(laptopClient)
}