package main

import (
	"postSaver/internal/adapters/application"
	"postSaver/internal/adapters/driven/saver"
	"postSaver/internal/adapters/driven/store"

	"postSaver/internal/adapters/driver/rpc"
	"postSaver/internal/logger"
)

func main() {
	store := store.NewStore()
	//recorder := recorder.NewRecorder()
	saver, err := saver.NewSaver("results")
	if err != nil {
		logger.L.Errorf("in main.main cannot create saver: %v\n", err)
	}
	app := application.NewApp(store, saver)
	receiver := rpc.NewReceiver(app)
	receiver.Run()
}
