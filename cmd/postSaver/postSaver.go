// Assembly point
package main

import (
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
	app := application.NewApp(store, saver)
	receiver := rpc.NewReceiver(app)
	receiver.Run()
}
