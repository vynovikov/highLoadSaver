// Receiver adapter
package rpc

import (
	"context"
	"os"
	"strconv"
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
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	kafkaPartitionString := os.Getenv("KAFKA_PARTITION")
	logger.L.Infof("in rpc.NewReceiver partition string: %s\n", kafkaPartitionString)
	partition, err := strconv.Atoi(kafkaPartitionString)
	if err != nil {
		logger.L.Errorf("in main.main cannot convert %v\n", err)
	}
	logger.L.Infof("in rpc.NewReceiver addr = %s, topic = %s, partition = %d\n", kafkaAddr, kafkaTopic, partition)
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaAddr, kafkaTopic, partition)
	if err != nil {
		logger.L.Errorf("in rpc.NewReceiver cannot dial: %w\n", err)
		time.Sleep(time.Second * 20)
		if conn != nil {
			conn.Close()
		}
	}
	logger.L.Infoln("in rpc.NewReceiver dialed successfully")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaAddr},
		GroupID:   kafkaPartitionString,
		Topic:     kafkaTopic,
		Partition: partition,
		MaxBytes:  10e6, // 10MB
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
		m, err := r.R.FetchMessage(context.Background())

		if err != nil {
			logger.L.Errorf("in rpc.Run cannot read from kafka: %v\n", err)
			if strings.Contains(err.Error(), "connection refused") {

				time.Sleep(time.Second * 5)
				r = NewReceiver(r.A)
				continue
			}
		}
		logger.L.Infof("in rpc.Run from message have read topic: %s, partition = %d, key: %q, value: %q\n", m.Topic, m.Partition, m.Key, m.Value)

		if err := r.R.CommitMessages(context.Background(), m); err != nil {
			logger.L.Errorf("in rpc.Run failed to commit messages: %v\n", err)
		}
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
