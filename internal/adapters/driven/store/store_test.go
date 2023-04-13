package store

import (
	"errors"
	"postSaver/internal/adapters/driver/rpc/pb"
	"postSaver/internal/repo"
	"testing"

	"github.com/stretchr/testify/suite"
)

type storeSuite struct {
	suite.Suite
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(storeSuite))
}

func (s *storeSuite) TestToTable() {
	tt := []struct {
		name      string
		store     Store
		reqs      []repo.Request
		wantStore Store
		wantError error
	}{
		{
			name: "1 unary",
			store: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{},
			},
			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    true,
						Ts:        "qqq",
						Name:      "alice",
						ByteChunk: []byte("azaza"),
					},
				},
			},
			wantStore: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{
					"qqq": {
						"alice": repo.NameNumber{
							Name: "azaza"},
					},
				},
			},
		},

		{
			name: "3 unaries",
			store: &StoreStruct{
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
			wantStore: &StoreStruct{
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
			name: "3 unaries, one is short file",
			store: &StoreStruct{
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
						Filename:  "long.txt",
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
			wantStore: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{
					"qqq": {
						"alice": repo.NameNumber{
							Name: "azaza"},
						"bob": repo.NameNumber{
							Name: "long.txt"},
						"cindel": repo.NameNumber{
							Name: "czczc"},
					},
				},
			},
		},

		{
			name: "3 unaries, all are short files",
			store: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{},
			},
			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    false,
						Ts:        "qqq",
						Name:      "alice",
						Filename:  "first.txt",
						ByteChunk: []byte("azaza"),
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   false,
						IsLast:    false,
						Ts:        "qqq",
						Name:      "bob",
						Filename:  "second.txt",
						ByteChunk: []byte("bzbzb"),
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   false,
						IsLast:    true,
						Ts:        "qqq",
						Name:      "cindel",
						Filename:  "third.txt",
						ByteChunk: []byte("czczc"),
					},
				},
			},
			wantStore: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{
					"qqq": {
						"alice": repo.NameNumber{
							Name: "first.txt"},
						"bob": repo.NameNumber{
							Name: "second.txt"},
						"cindel": repo.NameNumber{
							Name: "third.txt"},
					},
				},
			},
		},

		{
			name: "3 stream requests number is matcked",
			store: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{},
			},
			reqs: []repo.Request{
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								IsFirst:   true,
								Ts:        "qqq",
								FieldName: "alice",
								FileName:  "first.txt",
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "qqq",
								FieldName: "alice",
								Number:    uint32(0),
								ByteChunk: []byte("azaza"),
								IsLast:    false,
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "qqq",
								FieldName: "alice",
								Number:    uint32(1),
								ByteChunk: []byte("bzbzbz"),
								IsLast:    true,
							},
						},
					},
				},
			},
			wantStore: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{
					"qqq": {
						"alice": repo.NameNumber{
							Name:   "first.txt",
							Number: 2,
						},
					},
				},
			},
		},

		{
			name: "3 stream requests 2 unaries",
			store: &StoreStruct{
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
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								IsFirst:   false,
								Ts:        "qqq",
								FieldName: "bob",
								FileName:  "first.txt",
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "qqq",
								FieldName: "bob",
								Number:    uint32(0),
								ByteChunk: []byte("azaza"),
								IsLast:    false,
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "qqq",
								FieldName: "bob",
								Number:    uint32(1),
								ByteChunk: []byte("bzbzbz"),
								IsLast:    false,
							},
						},
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
			wantStore: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{
					"qqq": {
						"alice": repo.NameNumber{
							Name: "azaza"},
						"bob": repo.NameNumber{
							Name:   "first.txt",
							Number: 2,
						},
						"cindel": repo.NameNumber{
							Name: "czczc"},
					},
				},
			},
		},
		{
			name: "3 stream requests, number is unmatched",
			store: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{},
			},
			reqs: []repo.Request{
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								IsFirst:   true,
								Ts:        "qqq",
								FieldName: "alice",
								FileName:  "first.txt",
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "qqq",
								FieldName: "alice",
								Number:    uint32(0),
								ByteChunk: []byte("azaza"),
								IsLast:    false,
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "qqq",
								FieldName: "alice",
								Number:    uint32(2),
								ByteChunk: []byte("bzbzbz"),
								IsLast:    true,
							},
						},
					},
				},
			},
			wantStore: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{
					"qqq": {
						"alice": repo.NameNumber{
							Name:   "first.txt",
							Number: 1,
						},
					},
				},
			},
			wantError: errors.New("in store.ToTable request number 2 is not matched with store number 1"),
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			for _, w := range v.reqs {
				v.store.ToTable(w)
			}
			s.Equal(v.wantStore, v.store)

		})
	}
}

func (s *storeSuite) TestBufferAdd() {
	tt := []struct {
		name      string
		store     *StoreStruct
		r         repo.Request
		wantStore *StoreStruct
	}{

		{
			name:  "len(s.B)==0",
			store: &StoreStruct{},
			r: &repo.ReqStream{
				R: &pb.FileUploadReq{
					Info: &pb.FileUploadReq_FileData{
						FileData: &pb.FileData{
							Ts:        "qqq",
							FieldName: "alice",
							Number:    uint32(2),
							ByteChunk: []byte("azaza"),
							IsLast:    true,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(2),
										ByteChunk: []byte("azaza"),
										IsLast:    true,
									},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "len(s.B)==1, r.Number > mapRecordNumber => appending",
			store: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(2),
										ByteChunk: []byte("azaza"),
										IsLast:    true,
									},
								},
							},
						},
					},
				},
			},
			r: &repo.ReqStream{
				R: &pb.FileUploadReq{
					Info: &pb.FileUploadReq_FileData{
						FileData: &pb.FileData{
							Ts:        "qqq",
							FieldName: "alice",
							Number:    uint32(4),
							ByteChunk: []byte("bzbzb"),
							IsLast:    true,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(2),
										ByteChunk: []byte("azaza"),
										IsLast:    true,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(4),
										ByteChunk: []byte("bzbzb"),
										IsLast:    true,
									},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "len(s.B)==1, r.Number < mapRecordNumber => prepending",
			store: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(4),
										ByteChunk: []byte("azaza"),
										IsLast:    true,
									},
								},
							},
						},
					},
				},
			},
			r: &repo.ReqStream{
				R: &pb.FileUploadReq{
					Info: &pb.FileUploadReq_FileData{
						FileData: &pb.FileData{
							Ts:        "qqq",
							FieldName: "alice",
							Number:    uint32(3),
							ByteChunk: []byte("bzbzb"),
							IsLast:    true,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(3),
										ByteChunk: []byte("bzbzb"),
										IsLast:    true,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(4),
										ByteChunk: []byte("azaza"),
										IsLast:    true,
									},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "len(s.B)==2, firstRecordNumber + 1 == r.Number == lastRecordNumber - 1  => prepending + swapping",
			store: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(2),
										ByteChunk: []byte("azaza"),
										IsLast:    true,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(4),
										ByteChunk: []byte("bzbzb"),
										IsLast:    true,
									},
								},
							},
						},
					},
				},
			},
			r: &repo.ReqStream{
				R: &pb.FileUploadReq{
					Info: &pb.FileUploadReq_FileData{
						FileData: &pb.FileData{
							Ts:        "qqq",
							FieldName: "alice",
							Number:    uint32(3),
							ByteChunk: []byte("czczc"),
							IsLast:    true,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(2),
										ByteChunk: []byte("azaza"),
										IsLast:    true,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(3),
										ByteChunk: []byte("czczc"),
										IsLast:    true,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(4),
										ByteChunk: []byte("bzbzb"),
										IsLast:    true,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "len(s.B)==2, firstRecordNumber + 1 < r.Number < lastRecordNumber - 1  => sorting",
			store: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(2),
										ByteChunk: []byte("azaza"),
										IsLast:    true,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(6),
										ByteChunk: []byte("bzbzb"),
										IsLast:    true,
									},
								},
							},
						},
					},
				},
			},
			r: &repo.ReqStream{
				R: &pb.FileUploadReq{
					Info: &pb.FileUploadReq_FileData{
						FileData: &pb.FileData{
							Ts:        "qqq",
							FieldName: "alice",
							Number:    uint32(4),
							ByteChunk: []byte("czczc"),
							IsLast:    true,
						},
					},
				},
			},
			wantStore: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(2),
										ByteChunk: []byte("azaza"),
										IsLast:    true,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(4),
										ByteChunk: []byte("czczc"),
										IsLast:    true,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(6),
										ByteChunk: []byte("bzbzb"),
										IsLast:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			v.store.BufferAdd(v.r)
			s.Equal(v.wantStore, v.store)

		})
	}
}

func (s *storeSuite) TestReleaseuffer() {
	tt := []struct {
		name       string
		store      *StoreStruct
		wantStore  *StoreStruct
		wantReqs   []repo.Request
		wantErrors []error
	}{

		{
			name:      "nil buffer",
			store:     &StoreStruct{},
			wantStore: &StoreStruct{},
			wantErrors: []error{
				errors.New("in store.ReleaseBuffer buffer has no elements"),
			},
		},

		{
			name: "empty buffer",
			store: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {},
				},
			},
			wantStore: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {},
				},
			},
			wantErrors: []error{
				errors.New("in store.ReleaseBuffer buffer has no elements"),
			},
		},

		{
			name: "1 buffer element is matched, 1 isn't",
			store: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{
					"qqq": {
						"alice": repo.NameNumber{
							Name:   "first.txt",
							Number: 3,
						},
					},
				},
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(3),
										ByteChunk: []byte("azaza"),
										IsLast:    false,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(6),
										ByteChunk: []byte("bzbzb"),
										IsLast:    false,
									},
								},
							},
						},
					},
				},
			},
			wantStore: &StoreStruct{
				T: map[string]map[string]repo.NameNumber{
					"qqq": {
						"alice": repo.NameNumber{
							Name:   "first.txt",
							Number: 4,
						},
					},
				},
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(6),
										ByteChunk: []byte("bzbzb"),
										IsLast:    false,
									},
								},
							},
						},
					},
				},
			},
			wantReqs: []repo.Request{
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "qqq",
								FieldName: "alice",
								Number:    uint32(3),
								ByteChunk: []byte("azaza"),
								IsLast:    false,
							},
						},
					},
				},
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			gotReqs, gotErr := v.store.ReleaseBuffer()
			if gotErr != nil {
				s.Equal(v.wantErrors, gotErr)
			}
			s.Equal(v.wantStore, v.store)
			s.Equal(v.wantReqs, gotReqs)

		})
	}
}

func (s *storeSuite) TestCleanBuffer() {
	tt := []struct {
		name      string
		store     *StoreStruct
		ids       repo.IDsToRemove
		wantStore *StoreStruct
	}{
		{
			name: "len(ids.I) == 0",
			store: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(6),
										ByteChunk: []byte("bzbzb"),
										IsLast:    false,
									},
								},
							},
						},
					},
				},
			},
			wantStore: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(6),
										ByteChunk: []byte("bzbzb"),
										IsLast:    false,
									},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "len(ids.I) == len(s.B[ids.TS])",
			store: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(2),
										ByteChunk: []byte("bzbzb"),
										IsLast:    false,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(3),
										ByteChunk: []byte("bzbzb"),
										IsLast:    false,
									},
								},
							},
						},
					},
				},
			},
			ids: repo.IDsToRemove{
				TS: "qqq",
				I:  []int{0, 1},
			},
			wantStore: &StoreStruct{
				B: map[string][]repo.Request{},
			},
		},

		{
			name: "release 2, leave 1 alone",
			store: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(2),
										ByteChunk: []byte("bzbzb"),
										IsLast:    false,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(3),
										ByteChunk: []byte("bzbzb"),
										IsLast:    false,
									},
								},
							},
						},
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(5),
										ByteChunk: []byte("czczc"),
										IsLast:    false,
									},
								},
							},
						},
					},
				},
			},
			ids: repo.IDsToRemove{
				TS: "qqq",
				I:  []int{0, 1},
			},
			wantStore: &StoreStruct{
				B: map[string][]repo.Request{
					"qqq": {
						&repo.ReqStream{
							R: &pb.FileUploadReq{
								Info: &pb.FileUploadReq_FileData{
									FileData: &pb.FileData{
										Ts:        "qqq",
										FieldName: "alice",
										Number:    uint32(5),
										ByteChunk: []byte("czczc"),
										IsLast:    false,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			v.store.CleanBuffer(v.ids)
			s.Equal(v.wantStore, v.store)

		})
	}
}
