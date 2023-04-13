package recorder

import (
	"context"
	"os"
	"postSaver/internal/adapters/driven/recorder/pb"
	"postSaver/internal/logger"
	"postSaver/internal/repo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RecorderStruct struct {
	loggerClient pb.LoggerClient
}

type Recorder interface {
	Record(string) error
}

func NewRecorder() *RecorderStruct {
	loggerHostName := os.Getenv("LOGGER_HOSTNAME")
	if len(loggerHostName) == 0 {
		loggerHostName = "localhost"
	}
	connStringToLogger := loggerHostName + ":" + "3200"
	connToLogger, err := grpc.Dial(connStringToLogger, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.L.Errorf("in tologger.newToLogger unable to connect to logger service: %v\n", err)
	}
	c := pb.NewLoggerClient(connToLogger)
	//logger.L.Infof("in recorder.NewRecorder c: %v\n", c)
	return &RecorderStruct{
		loggerClient: c,
	}
}

func (r *RecorderStruct) Record(s string) error {
	req := &pb.LogReq{}

	req.Ts = repo.NewTS()
	req.LogString = s

	_, err := r.loggerClient.Log(context.Background(), req)
	if err != nil {
		logger.L.Errorf("in tologger.Log unable to execute Log with logString \"%s\": %v\n", s, err)
		return err
	}
	return nil
}
