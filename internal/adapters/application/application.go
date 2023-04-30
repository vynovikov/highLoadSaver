// Application layer.
// All driven adapters are accessible from here
package application

import (
	"fmt"
	"strings"

	"github.com/vynovikov/postSaver/internal/adapters/driven/saver"
	"github.com/vynovikov/postSaver/internal/adapters/driven/store"

	"sync"
	"time"

	"github.com/vynovikov/postSaver/internal/logger"
	"github.com/vynovikov/postSaver/internal/repo"
)

type ApplicationStruct struct {
	St       store.Store
	Sv       saver.Saver
	stopping bool
	timer    *time.Timer
	done     chan struct{}
	l        sync.Mutex
}

func NewAppStoreOnly(st store.Store) *ApplicationStruct {
	done := make(chan struct{})
	return &ApplicationStruct{
		St:   st,
		done: done,
	}
}

func NewApp(st store.Store, sv saver.Saver) (*ApplicationStruct, chan struct{}) {
	done := make(chan struct{})
	return &ApplicationStruct{
		St:   st,
		Sv:   sv,
		done: done,
	}, done
}

type Application interface {
	HandleUnary(repo.Request)
	HandleStream(repo.Request) error
	FileClose(repo.Request) error
	TableSave(string) error
	ClearStore(string)
	LastAction(string)
	Stop()
}

// HandleUnary saves request data to .json table.
// Tested in application_test.go
func (a *ApplicationStruct) HandleUnary(r repo.Request) {
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
	if r.IsLast() {
		go a.LastAction(r.TS())
		if a.stopping {
			close(a.done)
		}
	}

}

// HandleStrean saves requests in .json table.
// Saves files on disk
func (a *ApplicationStruct) HandleStream(r repo.Request) error {
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
			return fmt.Errorf("in application.HandleStream request ts %s, name %q, number %d is out of order and was sent to buffer")
		}
		a.l.Unlock()

		a.St.ToTable(r)
		err := a.Sv.FileWrite(r)
		if err != nil {
			logger.L.Errorf("in application.HandleStream unable to write to file: %v\n", err)
		}
		if r.IsLast() {
			go a.LastAction(r.TS())
			if a.stopping {
				close(a.done)
			}
			return nil
		}
		reqs, err := a.St.ReleaseBuffer()
		if err != nil && !strings.Contains(err.Error(), "no elements") {
			logger.L.Errorf("in application.HandleStream errors during release from buffer: %v\n", err)
		}
		for _, v := range reqs {
			a.St.ToTable(v)
			err := a.Sv.FileWrite(v)
			if err != nil {
				logger.L.Errorf("in application.HandleStream unable to write to file: %v\n", err)
			}
			if v.IsLast() {
				go a.LastAction(v.TS())
				if a.stopping {
					close(a.done)
				}
				return nil
			}
		}

	}
	return nil
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

func (a *ApplicationStruct) LastAction(ts string) {
	time.Sleep(time.Millisecond * 50)
	a.l.Lock()
	defer a.l.Unlock()
	m := a.St.GetTable(ts)
	a.Sv.TableSave(m, ts)
	a.St.CleanTableMap(ts)
}

func (a *ApplicationStruct) ClearStore(ts string) {
	a.l.Lock()
	defer a.l.Unlock()
	a.St.CleanTableMap(ts)
}

func (a *ApplicationStruct) Stop() {
	a.stopping = true
	if !a.St.Busy() {
		close(a.done)
	}
}
