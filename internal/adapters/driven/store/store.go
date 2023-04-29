// Store adapter.
// Manages data being saved
package store

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/vynovikov/postSaver/internal/repo"
)

type Store interface {
	CheckTS(string) bool
	ToTable(repo.Request) error
	GetTable(string) map[string]repo.NameNumber
	BufferAdd(repo.Request)
	ReleaseBuffer() ([]repo.Request, error)
	RequestMatched(repo.Request) bool
	CleanTableMap(string)
}

type StoreStruct struct {
	T   map[string]map[string]repo.NameNumber
	B   map[string][]repo.Request
	rwl sync.RWMutex
}

func NewStore() *StoreStruct {
	t := make(map[string]map[string]repo.NameNumber)
	b := make(map[string][]repo.Request)

	return &StoreStruct{
		T: t,
		B: b,
	}
}
func (s *StoreStruct) CheckTS(ts string) bool {
	s.rwl.RLock()
	defer s.rwl.RUnlock()
	if _, ok := s.T[ts]; ok {
		return true
	}
	return false
}

// ToTable invokes for every request and updates store.T.
// Tested in store_test.go
func (s *StoreStruct) ToTable(r repo.Request) error {
	l2 := make(map[string]repo.NameNumber)
	//defer logger.L.Infof("in store.ToTable s.T became %v\n", s.T)

	switch {
	case len(r.FileName()) == 0:
		l2[r.Name()] = repo.NewNameNumber(string(r.GetBody()), 0)
	case len(r.FileName()) > 0:
		l2[r.Name()] = repo.NewNameNumber(r.FileName(), 0)
	}
	s.rwl.Lock()
	defer s.rwl.Unlock()
	if m1, ok1 := s.T[r.TS()]; ok1 {
		rec, ok2 := m1[r.Name()]
		switch {
		case ok2 && len(r.GetBody()) > 0 && rec.Number == r.Number():
			rec.Number++
			s.T[r.TS()][r.Name()] = rec
		case ok2 && len(r.GetBody()) > 0:
			return fmt.Errorf("in store.ToTable request number %d is not matched with store number %d", r.Number(), rec.Number)
		case len(r.FileName()) == 0:
			s.T[r.TS()][r.Name()] = repo.NewNameNumber(string(r.GetBody()), 0)
		case len(r.FileName()) > 0:
			s.T[r.TS()][r.Name()] = repo.NewNameNumber(r.FileName(), 0)
		}
		return nil
	}

	s.T[r.TS()] = l2
	return nil

}

func (s *StoreStruct) GetTable(ts string) map[string]repo.NameNumber {
	if t, ok := s.T[ts]; ok {
		return t
	}
	return nil
}

func (s *StoreStruct) GetTableIn(tmr *time.Timer, ts string) {
	<-tmr.C
}

// BufferAdd adds request to buffer.
// Keeps buffer being sorter after each addition.
// Tested in store_test.go
func (s *StoreStruct) BufferAdd(r repo.Request) {
	s.rwl.Lock()
	defer s.rwl.Unlock()
	rts := r.TS()
	if s.B == nil {
		s.B = map[string][]repo.Request{}
	}
	switch lenb := len(s.B[rts]); {

	case lenb == 0:

		s.B[rts] = make([]repo.Request, 0)
		s.B[rts] = append(s.B[rts], r)

	case lenb == 1:

		e := s.B[rts][0]

		if Equal(r, e) { // avoiding dublicates
			return
		}

		if r.Number() >= e.Number() {
			s.B[rts] = append(s.B[rts], r)
		}
		if r.Number() < e.Number() {
			s.Prepend(s.B[rts], r)
		}

	default:

		for _, v := range s.B[rts] { // avoiding dublicates
			if Equal(r, v) {
				return
			}
		}

		f := s.B[rts][0]
		l := s.B[rts][len(s.B[rts])-1]

		if f.Number()+1 != l.Number() { //if first and last elements are not neighbours

			if r.Number() > l.Number() { //new element's Number is more than last element's
				s.B[rts] = append(s.B[rts], r)
			}
			if r.Number() < f.Number() { //new element's Number is less than first element's
				s.Prepend(s.B[rts], r)
			}
			if r.Number() == l.Number()-1 {

				lastIndex := len(s.B[rts]) - 1
				s.B[rts] = append(s.B[rts], r)

				Swap(s.B[rts], lastIndex, lastIndex+1)

				return

			}
			if r.Number() == f.Number()+1 {

				s.Prepend(s.B[rts], r)

				Swap(s.B[rts], 0, 1)

				return

			}
			if r.Number() < l.Number()-1 && //new element's Number is less than last element's minus 1
				r.Number() > f.Number()+1 { //new element's Number is more than first element's plus 1

				s.B[rts] = append(s.B[rts], r)
				b := s.B[rts]
				sort.SliceStable(s.B[rts], func(i int, j int) bool { return b[i].Number() < b[j].Number() })
				return
			}

		}
		if f.Number()+1 == l.Number() { //if first and last element are neighbours
			if r.Number() > l.Number() { //new element's Number is more than last element's
				s.B[rts] = append(s.B[rts], r)
			}
			if r.Number() < f.Number() { //new element's Number is less than first element's
				s.Prepend(s.B[rts], r)
			}
		}
	}
}

func Equal(a, b repo.Request) bool {
	if a.TS() == b.TS() &&
		a.Name() == b.Name() &&
		a.Number() == b.Number() &&
		bytes.Contains(a.GetBody(), b.GetBody()) {
		return true
	}
	return false
}

func (s *StoreStruct) Prepend(tsb []repo.Request, r repo.Request) {

	rts := r.TS()

	if b, ok := s.B[rts]; ok {
		s.B[rts] = append([]repo.Request{}, r)
		s.B[rts] = append(s.B[rts], b...)
		return
	}
	s.B[rts] = append([]repo.Request{}, r) // something went wrong
}

func Swap(s []repo.Request, i, j int) {
	e := s[i]
	s[i] = s[j]
	s[j] = e
}

// ReleaseBuffer retuns requests which can be saved.
// Tested in store_test.go
func (s *StoreStruct) ReleaseBuffer() ([]repo.Request, error) {

	s.rwl.Lock()
	defer s.rwl.Unlock()

	reqs, ids := make([]repo.Request, 0), repo.IDsToRemove{}

	if len(s.B) == 0 {
		return nil, fmt.Errorf("in store.ReleaseBuffer buffer has no elements")
	}
	found := false
	for _, v := range s.B {
		if len(v) > 0 {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("in store.ReleaseBuffer buffer has no elements")
	}

	for i, v := range s.B {

		if m1, ok := s.T[i]; ok {

			for j, w := range v {

				if record, ok := m1[w.Name()]; ok {

					if record.Number == w.Number() {

						reqs = append(reqs, w)
						ids.Add(repo.NewIDToRemove(i, j))

					}
				}
			}
		}
	}
	s.CleanBuffer(ids)
	return reqs, nil
}

// CleanBuffer deletes requests from buffer.
// Buffer remains sorted after deletion.
// Tested in store_test.go
func (s *StoreStruct) CleanBuffer(ids repo.IDsToRemove) {

	if len(ids.I) == 0 {
		return
	}
	ts, marked := ids.TS, 0
	// setting Requests with ids == ids.I to empty structs
	for _, v := range ids.I {
		s.B[ts][v] = &repo.ReqStream{}
		marked++
	}
	if marked == len(s.B[ts]) {
		delete(s.B, ts)
	}
	// sorting B putting empty ones to beginning
	sort.SliceStable(s.B[ts], func(a, b int) bool {
		return s.B[ts][a].TS() < s.B[ts][b].TS()
	})
	//cutting off the beginning
	for j := range s.B[ts] {
		if s.B[ts][j].TS() != "" {
			s.B[ts] = s.B[ts][j:]
			break
		}
		if j == len(s.B[ts])-1 {
			s.B[ts] = []repo.Request{}
		}
	}

}

// RequestMatched returns true if request is in store.T
func (s *StoreStruct) RequestMatched(r repo.Request) bool {

	if r.IsStreamInfo() || r.IsUnary() {
		return true
	}

	if r.IsStreamData() {
		if _, ok := s.T[r.TS()]; !ok {
			return false
		}
	}

	if rec, ok := s.T[r.TS()][r.Name()]; ok {
		if r.Number() == rec.Number {
			return true
		}
	}
	return false
}

func (s *StoreStruct) CleanTableMap(ts string) {
	delete(s.T, ts)
	if len(s.T) == 0 {
		s.T = make(map[string]map[string]repo.NameNumber)
	}
}
