package application

import (
	"testing"
	"time"

	"github.com/vynovikov/postSaver/internal/adapters/driven/store"
	"github.com/vynovikov/postSaver/internal/adapters/driver/rpc/pb"
	"github.com/vynovikov/postSaver/internal/repo"

	"github.com/stretchr/testify/suite"
)

type applicationSuite struct {
	suite.Suite
}

func TestApplicationSuite(t *testing.T) {
	suite.Run(t, new(applicationSuite))
}

var testApp *ApplicationStruct

func (s *applicationSuite) TestHandleUnary() {
	tt := []struct {
		name      string
		store     store.Store
		reqs      []repo.Request
		wantStore store.Store
	}{
		{
			name: "correct order",
			store: &store.StoreStruct{
				T: map[string]map[string]repo.NameNumber{},
			},

			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    false,
						Ts:        "qqq",
						Name:      "alice",
						ByteChunk: []byte("azaza"),
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   false,
						IsLast:    false,
						Ts:        "qqq",
						Name:      "bob",
						ByteChunk: []byte("bzbzb"),
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   false,
						IsLast:    true,
						Ts:        "qqq",
						Name:      "cindel",
						ByteChunk: []byte("czczc"),
					},
				},
			},
			wantStore: &store.StoreStruct{
				T: map[string]map[string]repo.NameNumber{
					"qqq": {
						"alice": repo.NameNumber{
							Name: "azaza"},
						"bob": repo.NameNumber{
							Name: "bzbzb"},
						"cindel": repo.NameNumber{
							Name: "czczc"},
					},
				},
			},
		},
		{
			name: "incorrect order",
			store: &store.StoreStruct{
				T: map[string]map[string]repo.NameNumber{},
			},

			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    false,
						Ts:        "qqq",
						Name:      "alice",
						ByteChunk: []byte("azaza"),
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   false,
						IsLast:    true,
						Ts:        "qqq",
						Name:      "cindel",
						ByteChunk: []byte("czczc"),
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   false,
						IsLast:    false,
						Ts:        "qqq",
						Name:      "bob",
						ByteChunk: []byte("bzbzb"),
					},
				},
			},
			wantStore: &store.StoreStruct{
				T: map[string]map[string]repo.NameNumber{
					"qqq": {
						"alice": repo.NameNumber{
							Name: "azaza"},
						"bob": repo.NameNumber{
							Name: "bzbzb"},
						"cindel": repo.NameNumber{
							Name: "czczc"},
					},
				},
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			app := NewAppStoreOnly(v.store)
			for _, w := range v.reqs {
				app.HandleUnary(w)
			}
			time.Sleep(time.Millisecond * 70)
			s.Equal(v.wantStore, app.St)
		})
	}
}
