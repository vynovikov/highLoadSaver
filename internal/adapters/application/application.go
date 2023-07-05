// Application layer.
// All driven adapters are accessible from here
package application

import (
	"github.com/segmentio/kafka-go"
	"github.com/vynovikov/highLoadSaver/internal/adapters/driven/saver"
	"github.com/vynovikov/highLoadSaver/internal/adapters/driver/rpc/pb"
	"google.golang.org/protobuf/proto"

	"sync"
	"time"
)

type ApplicationStruct struct {
	S        saver.Saver
	stopping bool
	timer    *time.Timer
	done     chan struct{}
	l        sync.Mutex
}

func NewAppStoreOnly(s saver.Saver) *ApplicationStruct {
	done := make(chan struct{})
	return &ApplicationStruct{
		S:    s,
		done: done,
	}
}

func NewApp(s saver.Saver) (*ApplicationStruct, chan struct{}) {
	done := make(chan struct{})
	return &ApplicationStruct{
		S:    s,
		done: done,
	}, done
}

type Application interface {
	HandleKafkaMessage(kafka.Message) error
	/*
		FileClose(repo.Request) error
		TableSave(string) error
		ClearStore(string)
		LastAction(string)
	*/
	Stop()
}

func (a *ApplicationStruct) HandleKafkaMessage(m kafka.Message) error {
	unmarshalled := &pb.Message{}

	err := proto.Unmarshal(m.Value, unmarshalled)
	if err != nil {
		return err
	}
	//logger.L.Infof("in application.HandleKafkaMessage %v\n", unmarshalled)
	return a.S.Save(unmarshalled)
}

func (a *ApplicationStruct) FileClose() error {
	return nil
}

func (a *ApplicationStruct) TableSave(ts string) error {
	return nil
}

func (a *ApplicationStruct) LastAction(ts string) {
}

func (a *ApplicationStruct) ClearStore(ts string) {
}

func (a *ApplicationStruct) Stop() {
	/*
	   a.stopping = true

	   	if !a.St.Busy() {
	   		close(a.done)
	   	}
	*/
	close(a.done)
}
