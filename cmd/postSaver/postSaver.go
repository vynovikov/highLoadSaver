// Assembly point
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/vynovikov/postSaver/internal/adapters/application"
	"github.com/vynovikov/postSaver/internal/adapters/driven/saver"
	"github.com/vynovikov/postSaver/internal/adapters/driven/store"

	"github.com/vynovikov/postSaver/internal/adapters/driver/rpc"
	"github.com/vynovikov/postSaver/internal/logger"
)

// Tested in postSaver_test.go
func main() {
	store := store.NewStore()
	saver, err := saver.NewSaver("results")
	if err != nil {
		logger.L.Errorf("in main.main cannot create saver: %v\n", err)
	}
	app, done := application.NewApp(store, saver)
	receiver := rpc.NewReceiver(app)
	go receiver.Run()
	go SignalListen(app)
	<-done
	logger.L.Errorln("postSaver is interrupted")
}

// SignalListen listens for Interrupt signal, when receiving one invokes stop function
func SignalListen(app application.Application) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	<-sigChan
	go app.Stop()
}
