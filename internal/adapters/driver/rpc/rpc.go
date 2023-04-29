// Receiver adapter
package rpc

import (
	"context"
	"io"
	"net"
	"os"
	"sync"

	"github.com/vynovikov/postSaver/internal/adapters/application"
	"github.com/vynovikov/postSaver/internal/adapters/driver/rpc/pb"
	"github.com/vynovikov/postSaver/internal/logger"
	"github.com/vynovikov/postSaver/internal/repo"

	"google.golang.org/grpc"
)

type ReceiverStruct struct {
	A application.Application
	pb.SaverServer
	Listener net.Listener
	Server   *grpc.Server
	l        sync.Mutex
}
type Receiver interface {
	Run()
	SinglePart(context.Context, *pb.TextFieldReq) (*pb.TextFieldRes, error)
	MultiPart(pb.Saver_MultiPartServer) error
}

func NewReceiver(a application.Application) *ReceiverStruct {

	lis, err := net.Listen("tcp", ":3100")
	if err != nil {
		logger.L.Errorf("in rpc.NewReceiver failed to listen on 3100: %v\n", err)
	}

	baseServer := grpc.NewServer()

	r := &ReceiverStruct{
		Listener: lis,
		A:        a,
	}

	pb.RegisterSaverServer(baseServer, r)
	r.Server = baseServer

	return r
}

func (r *ReceiverStruct) Run() {
	localHost := os.Getenv("HOSTNAME")
	if len(localHost) > 0 {
		logger.L.Infof("listening %s:3100", localHost)
	} else {
		logger.L.Infof("listening localhost:3100")
	}
	r.Server.Serve(r.Listener)
}

// SinglrPart makes *ReceiverStruct to implement pb.SaverServer interface.
// Initial point where gRPC unary requests are handled in saver
func (r *ReceiverStruct) SinglePart(ctx context.Context, in *pb.TextFieldReq) (*pb.TextFieldRes, error) {
	reqIn := repo.NewReqUnary(in)
	r.A.HandleUnary(reqIn)
	res := &pb.TextFieldRes{Result: true}
	return res, nil
}

// MultiPart makes *ReceiverStruct to implement pb.SaverServer interface.
// Initial point where gRPC stream requests are handled in saver
func (r *ReceiverStruct) MultiPart(stream pb.Saver_MultiPartServer) error {
	var (
		reqData  *pb.FileUploadReq
		errStore error
		err      error
	)
	n := 0
	reqInfo, err := stream.Recv()
	if err != nil {
		return err
	}

	rInfo := repo.NewReqStream(reqInfo)
	r.A.HandleStream(rInfo)

	for {
		reqData, err = stream.Recv()
		if reqData != nil {
			//error if goes to buffer
			errStore = r.A.HandleStream(repo.NewReqStream(reqData))
		}

		if err == io.EOF && errStore == nil {
			break
		}

		if err != nil {
			logger.L.Error(err)
			return err
		}

		n++

	}

	r.A.FileClose(rInfo)

	res := &pb.FileUploadRes{
		FileName: "filename",
		FileSize: uint32(100),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return err
	}

	return nil
}
