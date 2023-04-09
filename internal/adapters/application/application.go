package application

import (
	"postSaver/internal/adapters/driven/saver"
	"postSaver/internal/adapters/driven/store"
	"postSaver/internal/logger"
	"postSaver/internal/repo"
	"sync"
	"time"
)

type ApplicationStruct struct {
	St    store.Store
	Sv    saver.Saver
	timer *time.Timer
	l     sync.Mutex
}

func NewAppStoreOnly(st store.Store) *ApplicationStruct {
	return &ApplicationStruct{
		St: st,
	}
}

func NewApp(st store.Store, sv saver.Saver) *ApplicationStruct {
	return &ApplicationStruct{
		St: st,
		Sv: sv,
	}
}

type Application interface {
	HandleUnary(repo.Request)
	HandleStream(repo.Request)
	FileClose(repo.Request) error
	TableSave(string) error
}

func (a *ApplicationStruct) HandleUnary(r repo.Request) {
	// Update this considering gRPC request shuffling

	a.St.ToTable(r)
	if r.FileName() != "" {
		_, err := a.Sv.FileCreate(r)
		if err != nil {
			logger.L.Errorf("in application.HandleUnary unable to create file: %v\n", err)
		}
		err = a.Sv.FileWrite(r)
		if err != nil {
			logger.L.Errorf("in application.HandleUnary unable to write %q to file: %v\n", r.GetBody(), err)
		}
		err = a.Sv.FileClose(r)
		if err != nil {
			logger.L.Errorf("in application.HandleUnary unable to close file: %v\n", err)
		}
	}

}

func (a *ApplicationStruct) HandleStream(r repo.Request) {

	//n := r.Number()
	//logger.L.Infof("in main.HandleStream request %v, IsStreamInfo %t\n", r, r.IsStreamInfo())

	switch b := r.IsStreamInfo(); b {
	case true:
		a.St.ToTable(r)
		_, err := a.Sv.FileCreate(r)
		if err != nil {
			logger.L.Errorf("in application.HandleStream unable to create file: %v\n", err)
		}
	default:
		a.l.Lock()

		if !a.St.RequestMatched(r) {
			a.St.BufferAdd(r)
			a.l.Unlock()
			return
		}
		a.l.Unlock()

		a.St.ToTable(r)
		err := a.Sv.FileWrite(r)
		if err != nil {
			logger.L.Errorf("in application.HandleStream unable to write to file: %v\n", err)
		}
		reqs, errs := a.St.ReleaseBuffer()
		if errs != nil {
			logger.L.Errorf("in application.HandleStream errors during release from buffer: %v\n", errs)
		}
		for _, v := range reqs {
			a.St.ToTable(v)
			err := a.Sv.FileWrite(v)
			if err != nil {
				logger.L.Errorf("in application.HandleStream unable to write to file: %v\n", err)
			}
		}

	}
}
func (a *ApplicationStruct) FileClose(r repo.Request) error {
	return a.Sv.FileClose(r)
}

func (a *ApplicationStruct) TableSave(ts string) error {
	a.l.Lock()
	defer a.l.Unlock()
	m := a.St.GetTable(ts)
	return a.Sv.TableSave(m, ts)
}
