// Receiver adapter
package rpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
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
		conn *kafka.Conn
		err  error
	)

	kafkaAddr := os.Getenv("KAFKA_ADDR")
	kafkaPort := os.Getenv("KAFKA_PORT")

	dialURI := net.JoinHostPort(kafkaAddr, kafkaPort)

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	kafkaConsumerGroupID := os.Getenv("KAFKA_CONSUMER_GROUP_ID")

	logger.L.Infof("in rpc.NewReceiver addr = %s, topic = %s\n", dialURI, kafkaTopic)

	for i := 0; i < 5; i++ {

		logger.L.Infof("in rpc.NewReceiver %d attempt to dial to %s\n", i, dialURI)

		conn, err = kafka.Dial("tcp", dialURI)
		if err != nil {

			logger.L.Errorf("in rpc.NewReceiver cannot dial: %v\n", err)
			time.Sleep(time.Second * 10)

			continue
		}
	}
	if conn == nil {

		logger.L.Errorln("in rpc.NewReceiver failed to establish conn")
		os.Exit(1)
	}

	partitions, err := conn.ReadPartitions()
	if err != nil {

		logger.L.Errorln(err)
		os.Exit(1)
	}

	for _, p := range partitions {

		if p.Topic == kafkaTopic {

			logger.L.Infof("in rpc.NewReceiver topic %s is found\n", kafkaTopic)
		}
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{fmt.Sprintf("%s:%s", kafkaAddr, kafkaPort)},
		GroupID:  kafkaConsumerGroupID,
		Topic:    kafkaTopic,
		MaxBytes: 10e6, // 10MB
	})

	rcvr := &ReceiverStruct{
		A: a,
		R: r,
	}

	return rcvr
}

func (r *ReceiverStruct) Run() {

	logger.L.Infoln("waiting for kafka messages...")

	for {
		m, err := r.R.ReadMessage(context.Background())

		if err != nil {

			// TODO fix ERROR[06.08.2024 18:17:20] in rpc.Run cannot read from kafka: failed to dial:
			//failed to open connection to [http://localhost:29092]:9092: dial tcp: lookup http://localhost:29092: no such host

			logger.L.Errorf("in rpc.Run cannot read from kafka: %v\n", err)
			if strings.Contains(err.Error(), "connection refused") {

				time.Sleep(time.Second * 5)
				r = NewReceiver(r.A)
				continue
			}
		}
		logger.L.Infof("in rpc.Run from message have read topic: %s, partition = %d, key: %q, value: %q\n", m.Topic, m.Partition, m.Key, m.Value)

		//	if err := r.R.CommitMessages(context.Background(), m); err != nil {
		//		logger.L.Errorf("in rpc.Run failed to commit messages: %v\n", err)
		//	}
		header, body := &pb.MessageHeader{}, &pb.MessageBody{}

		if err = proto.Unmarshal(m.Key, header); err != nil {
			logger.L.Errorf("in rpc.Run failed to unmarshal: %v\n", err)
		}
		logger.L.Infof("in rpc.Run unmarshalled body: %v\n", header)

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
