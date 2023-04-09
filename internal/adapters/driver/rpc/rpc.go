package rpc

import (
	"context"
	"io"
	"net"
	"os"
	"postSaver/internal/adapters/application"
	"postSaver/internal/adapters/driver/rpc/pb"
	"postSaver/internal/logger"
	"postSaver/internal/repo"
	"sync"

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
func (r *ReceiverStruct) SinglePart(ctx context.Context, in *pb.TextFieldReq) (*pb.TextFieldRes, error) {
	//	r.l.Lock()
	//	defer r.l.Unlock()
	//logger.L.Infof("in rpc.SinglePart incoming request %v\n", in)
	reqIn := repo.NewReqUnary(in)
	r.A.HandleUnary(reqIn)
	if reqIn.IsLast() {
		r.A.TableSave(reqIn.TS())
	}
	res := &pb.TextFieldRes{Result: true}
	return res, nil
}

func (r *ReceiverStruct) MultiPart(stream pb.Saver_MultiPartServer) error {
	n := 0
	reqInfo, err := stream.Recv()
	//	r.l.Lock()
	if err != nil {
		return err
	}
	//logger.L.Infof("in rpc.MultiPart initial request %v\n", reqInfo)
	//logger.L.Infof("in rpc.MultiPart fileName: %s\n", reqInfo.GetFileInfo().FileName)

	rInfo := repo.NewReqStream(reqInfo)
	r.A.HandleStream(rInfo)

	for {
		reqData, err := stream.Recv()
		//	if n > 0 {
		//		r.l.Lock()
		//	}
		//logger.L.Infof("in rpc.MultiPart looped request %v, err: %v\n", reqData, err)
		if reqData != nil {
			r.A.HandleStream(repo.NewReqStream(reqData))
		}

		if err == io.EOF {
			//r.l.Unlock()
			break
		}

		if err != nil {
			logger.L.Error(err)
			return err
		}

		//logger.L.Infof("in main.MultiPart request number %v\n", reqData.GetFileData().Number)
		//logger.L.Infof("in main.MultiPart request GetFileInfo: %v, is nil? %t; GetFileData: %v, is nil? %t\n", reqData.GetFileInfo(), reqData.GetFileInfo() == nil, reqData.GetFileData(), reqData.GetFileData() == nil)
		n++
		//	r.l.Unlock()
	}

	r.A.FileClose(rInfo)

	err = r.A.TableSave(rInfo.TS())
	if err != nil {
		return err
	}

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
