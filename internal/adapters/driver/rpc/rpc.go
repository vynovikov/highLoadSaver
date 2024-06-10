// Receiver adapter
package rpc

import (
	"context"
	"net"
	"os"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vynovikov/highLoadSaver/internal/adapters/application"
	"github.com/vynovikov/highLoadSaver/internal/adapters/driver/rpc/pb"
	"github.com/vynovikov/highLoadSaver/internal/logger"
	"google.golang.org/protobuf/proto"
)

type ReceiverStruct struct {
	A application.Application
	R *kafka.Reader
	l sync.Mutex
}
type Receiver interface {
	Run()
}

func NewReceiver(a application.Application) *ReceiverStruct {

	var (
		conn       *kafka.Conn
		err        error
		partitions []kafka.Partition
	)

	kafkaAddr := os.Getenv("KAFKA_ADDR")
	kafkaPort := os.Getenv("KAFKA_PORT")

	dialURI := net.JoinHostPort(kafkaAddr, kafkaPort)

	kafkaTopic := os.Getenv("KAFKA_TOPIC")

	dialURI = net.JoinHostPort(kafkaAddr, kafkaPort)

	for i := 0; i < 5; i++ {

		conn, err = kafka.Dial("tcp", dialURI)
		if err != nil {

			logger.L.Errorf("in rpc.NewReceiver cannot dial: %v. Trying again\n", err)

			time.Sleep(time.Second * 10)

			continue
		}

		partitions, err = conn.ReadPartitions()

		if err != nil {

			time.Sleep(time.Second * 10)

			continue

		}

		for _, p := range partitions {

			if p.Topic == kafkaTopic {

				conn.Close()

				rs := &ReceiverStruct{

					A: a,
					R: kafka.NewReader(kafka.ReaderConfig{
						Brokers: []string{dialURI},
						Topic:   kafkaTopic,
						GroupID: "0",
					}),
				}

				logger.L.Infof("in rpc.NewReceiver rs: %v\n", rs.R)

				return rs
			}
		}
	}

	return &ReceiverStruct{}
}

func (r *ReceiverStruct) Run() {

	logger.L.Infoln("waiting for kafka messages...")

	for {
		m, err := r.R.ReadMessage(context.Background())

		if err != nil {

			logger.L.Errorf("in rpc.Run cannot read from kafka: %v receiver: %v\n", err, r.R)
		}

		logger.L.Infof("in rpc.Run from message have read topic: %s, partition = %d, key: %q, value: %q\n", m.Topic, m.Partition, m.Key, m.Value)

		header, body := &pb.MessageHeader{}, &pb.MessageBody{}

		if err = proto.Unmarshal(m.Key, header); err != nil {
			logger.L.Errorf("in rpc.Run failed to unmarshal: %v\n", err)
		}
		logger.L.Infof("in rpc.Run unmarshalled header: %v\n", header)

		if err = proto.Unmarshal(m.Value, body); err != nil {
			logger.L.Errorf("in rpc.Run failed to unmarshal: %v\n", err)
		}
		logger.L.Infof("in rpc.Run unmarshalled body: %v\n", body)

	}
	/*
			if err = r.A.HandleKafkaMessage(m); err != nil {
				logger.L.Errorf("in rpc.Run cannot handle message: %v\n", err)
			}
		}
	*/
	/*
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
	*/
}
