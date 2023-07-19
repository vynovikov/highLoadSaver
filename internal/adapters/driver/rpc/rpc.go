// Receiver adapter
package rpc

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vynovikov/highLoadSaver/internal/adapters/application"
	"github.com/vynovikov/highLoadSaver/internal/logger"
)

type ReceiverStruct struct {
	A application.Application
	c *kafka.Conn
	l sync.Mutex
}
type Receiver interface {
	Run()
}

func NewReceiver(a application.Application) *ReceiverStruct {
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	topic := os.Getenv("KAFKA_TOPIC")
	partition, err := strconv.Atoi(os.Getenv("KAFKA_PARTITION"))
	if err != nil {
		logger.L.Errorf("in main.main cannot convert %v\n", err)
	}
	logger.L.Infof("in rpc.NewReceiver addr = %s, topic = %s, partition = %d\n", kafkaAddr, topic, partition)
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaAddr, topic, partition)
	if err != nil {
		logger.L.Errorf("in rpc.NewReceiver cannot dial %f\n", err)
		time.Sleep(time.Second * 20)
		if conn != nil {
			conn.Close()
		}
	}
	//conn.Close()

	r := &ReceiverStruct{
		A: a,
		c: conn,
	}

	return r
}

func (r *ReceiverStruct) Run() {
	logger.L.Infoln("run invoked")
	batch := r.c.ReadBatch(10, 1e6)
	logger.L.Infof("in rpc.Run waiting for new messages from kafka %s topic ...\n", os.Getenv("KAFKA_TOPIC"))

	for {
		m, err := batch.ReadMessage()
		if err != nil {

		}
		err = r.A.HandleKafkaMessage(m)
		if err != nil {
			logger.L.Errorf("in rpc.Run cannot handle message: %v\n", err)
		}
	}
}
