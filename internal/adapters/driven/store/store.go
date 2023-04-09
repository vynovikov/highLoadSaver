package store

import (
	"bytes"
	"fmt"
	"postSaver/internal/repo"
	"sort"
	"sync"
	"time"
)

type Store interface {
	CheckTS(string) bool
	ToTable(repo.Request) error
	GetTable(string) map[string]repo.NameNumber
	//GetTableIn(*time.Timer, string)
	BufferAdd(repo.Request)
	ReleaseBuffer() ([]repo.Request, []error)
	RequestMatched(repo.Request) bool
}

type StoreStruct struct {
	T   map[string]map[string]repo.NameNumber
	B   map[string][]repo.Request
	rwl sync.RWMutex
}

func NewStore() *StoreStruct {
	t := make(map[string]map[string]repo.NameNumber)

	return &StoreStruct{
		T: t,
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
func (s *StoreStruct) ToTable(r repo.Request) error {
	//logger.L.Infof("store.ToTable invoked by Request name %q, number %d, s.T: %v\n", r.Name(), r.Number(), s.T)
	l2 := make(map[string]repo.NameNumber)
	//logger.L.Infof("in store.ToTable r: %v, number: %d, r.FileName() %q, len(r.FileName()): %d\n", r, number, r.FileName(), len(r.FileName()))
	switch {
	case len(r.FileName()) == 0:
		l2[r.Name()] = repo.NewNameNumber(string(r.GetBody()), 0)
	case len(r.FileName()) > 0:
		l2[r.Name()] = repo.NewNameNumber(r.FileName(), 0)
	}
	//logger.L.Infof("in store.ToTable after all l2: %v, r.Filename = %s\n", l2, r.FileName())
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
		//logger.L.Infof("in store.ToTable after all s.T: %v\n", s.T)
		return nil
	}

	s.T[r.TS()] = l2
	//logger.L.Infof("in store.ToTable after all s.T: %v\n", s.T)
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
	//logger.L.Infof("in store.GetTableIn table is %v\n", s.GetTable(ts))
}

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
	//logger.L.Infof("in store.BufferAdd s.B: %v, s.T: %v\n", s.B, s.T)
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

func (s *StoreStruct) ReleaseBuffer() ([]repo.Request, []error) {

	s.rwl.Lock()
	defer s.rwl.Unlock()

	reqs, errs, ids := make([]repo.Request, 0), make([]error, 0), repo.IDsToRemove{}

	//logger.L.Infof("in store.ReleaseBuffer s.B = %v\n", s.B)

	if len(s.B) == 0 {
		errs = append(errs, fmt.Errorf("in store.ReleaseBuffer buffer has no elements"))
		return nil, errs
	}
	found := false
	for _, v := range s.B {
		if len(v) > 0 {
			found = true
		}
	}
	if !found {
		errs = append(errs, fmt.Errorf("in store.ReleaseBuffer buffer has no elements"))
		return nil, errs
	}

	for i, v := range s.B {

		if m1, ok := s.T[i]; ok {

			for j, w := range v {

				if record, ok := m1[w.Name()]; ok {

					if record.Number == w.Number() {
						/*
							record.Number++
							m1[w.Name()] = record
						*/
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
		//logger.L.Infof("store.CleanBuffer j = %d, s.B[ids.TS][j].TS(): %v\n", j, s.B[ids.TS][j].TS())
		if s.B[ts][j].TS() != "" {
			s.B[ts] = s.B[ts][j:]
			break
		}
		if j == len(s.B[ts])-1 {
			s.B[ts] = []repo.Request{}
		}
	}

}

func (s *StoreStruct) RequestMatched(r repo.Request) bool {
	//logger.L.Infof("store.RequestMatched invoked for r: %v, s.T: %v \n", r, s.T[r.TS()])

	if r.IsStreamInfo() || r.IsUnary() {
		return true
	}

	if r.IsStreamData() {
		if _, ok := s.T[r.TS()]; !ok {
			return false
		}
	}

	if rec, ok := s.T[r.TS()][r.Name()]; ok {
		//logger.L.Infof("in store.RequestMatched rec number = %d, r number = %d \n", rec.Number, r.Number())
		if r.Number() == rec.Number {
			return true
		}
	}
	return false
}
