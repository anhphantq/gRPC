package service_test

import (
	"context"
	"gRPC/pb"
	"gRPC/sample"
	"gRPC/serializer"
	"gRPC/service"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestClientCreateLaptop(t *testing.T){
	t.Parallel()

	laptopStore := service.NewInMemoryLaptopStore()

	serverAddress := startTestLaptopServer(t, laptopStore, nil)
	laptopClinet := newTestLaptopClient(t, serverAddress)

	laptop := sample.NewLaptop()
	expectedID := laptop.Id
	req := &pb.CreateLatopRequest{
		Latop: laptop,
	}

	res, err := laptopClinet.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedID, res.Id)

	// 
	other, err := laptopStore.Find(res.Id)
	require.NoError(t, err)
	require.NotNil(t, other)

	requireSameLaptop(t,laptop, other)
}

func startTestLaptopServer(t *testing.T, laptopStore service.LaptopStore, imageStore service.ImageStore) string{
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, nil)

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go grpcServer.Serve(listener)

	return listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient{
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, a *pb.Laptop, b *pb.Laptop){
	json1, err := serializer.ProtobufToJSON(a)
	require.NoError(t, err)

	json2, err := serializer.ProtobufToJSON(b)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}

