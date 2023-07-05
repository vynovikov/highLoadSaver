// Assembly point
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/vynovikov/highLoadSaver/internal/adapters/application"
	"github.com/vynovikov/highLoadSaver/internal/adapters/driven/saver"

	"github.com/vynovikov/highLoadSaver/internal/adapters/driver/rpc"
	"github.com/vynovikov/highLoadSaver/internal/logger"
)

// Tested in highLoadSaver_test.go
func main() {
	saver, err := saver.NewSaver("results")
	if err != nil {
		logger.L.Errorf("in main.main cannot create saver: %v\n", err)
	}
	app, done := application.NewApp(saver)
	receiver := rpc.NewReceiver("data", app)
	go receiver.Run()
	go SignalListen(app)
	<-done
	logger.L.Errorln("highLoadSaver is interrupted")
}

// SignalListen listens for Interrupt signal, when receiving one invokes stop function
func SignalListen(app application.Application) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	<-sigChan
	go app.Stop()
}
