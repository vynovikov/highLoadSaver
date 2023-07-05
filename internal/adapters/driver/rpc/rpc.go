// Receiver adapter
package rpc

import (
	"context"
	"sync"

	"github.com/segmentio/kafka-go"
	"github.com/vynovikov/highLoadSaver/internal/adapters/application"
	"github.com/vynovikov/highLoadSaver/internal/logger"
)

type ReceiverStruct struct {
	A  application.Application
	kr *kafka.Reader
	l  sync.Mutex
}
type Receiver interface {
	Run()
}

func NewReceiver(t string, a application.Application) *ReceiverStruct {

	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     t,
		GroupID:   "1",
		Partition: 0,
		MaxBytes:  10e6, // 10MB
	})

	r := &ReceiverStruct{
		A:  a,
		kr: kr,
	}

	return r
}

func (r *ReceiverStruct) Run() {
	for {
		m, err := r.kr.FetchMessage(context.Background())
		if err != nil {
			break
		}
		//logger.L.Infof("message at offset %d: %s = %q", m.Offset, string(m.Key), string(m.Value))
		err = r.kr.CommitMessages(context.Background(), m)
		if err != nil {
			logger.L.Errorf("in rpc.Run cannot commit message: %v\n", err)
		}
		err = r.A.HandleKafkaMessage(m)
		if err != nil {
			logger.L.Errorf("in rpc.Run cannot handle message: %v\n", err)
		}
	}
}
