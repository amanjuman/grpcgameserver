package main

import (
	"net"

	pb "github.com/amanjuman/grpcgameserver/proto"

	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})

	log.SetOutput(os.Stdout)

	log.SetLevel(log.DebugLevel)
}

const (

	Address = ":8080"
)

type rpcService struct{}

var rpc *rpcService = &rpcService{}

func (r *rpcService) SyncPostion(ctx context.Context, in *pb.Pos) (*pb.PosReply, error) {
	log.Debug(ctx)
	log.Debug(in)
	return new(pb.PosReply), nil
}

func (r *rpcService) CallServer(ctx context.Context, in *pb.Callin) (*pb.Reply, error) {
	log.Debug(in)
	return new(pb.Reply), nil
}
func (r *rpcService) CallClient(in *pb.ClientStart, stream pb.Packet_CallClientServer) error {
	log.Debug(in)
	return nil
}

func main() {
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		fmt.Println("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterPacketServer(s, &rpcService{})

	fmt.Println("Listen on " + Address)

	s.Serve(listen)
}
