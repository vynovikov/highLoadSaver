package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vynovikov/postSaver/internal/adapters/driver/rpc/pb"
	"github.com/vynovikov/postSaver/internal/logger"
	"github.com/vynovikov/postSaver/internal/repo"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type mainSuite struct {
	suite.Suite
}

func TestMainSuite(t *testing.T) {
	suite.Run(t, new(mainSuite))
}

func (s *mainSuite) TestWorkFlow() {
	g, generatorChan := newGenerator()
	go main()
	go g.generate(generatorChan)
	time.Sleep(time.Millisecond * 100)
	tt := []struct {
		name        string
		ts          string
		reqs        []repo.Request
		wantTable   map[string]string
		wantContent map[string][]byte
		wantError   error
	}{

		{
			name: "1 unary field",
			ts:   "001",
			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    true,
						Ts:        "001",
						Name:      "alice",
						ByteChunk: []byte("azaza"),
					},
				},
			},

			wantTable: map[string]string{
				"alice": "azaza",
			},
			wantContent: map[string][]byte{},
		},

		{
			name: "1 unary file",
			ts:   "002",
			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    true,
						Ts:        "002",
						Name:      "alice",
						Filename:  "first.txt",
						ByteChunk: []byte("azaza"),
					},
				},
			},
			wantTable: map[string]string{
				"alice": "first.txt",
			},
			wantContent: map[string][]byte{
				"alice": []byte("azaza"),
			},
		},

		{
			name: "3 unaries files",
			ts:   "003",
			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    false,
						Ts:        "003",
						Name:      "alice",
						Filename:  "first.txt",
						ByteChunk: []byte("azaza"),
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   false,
						IsLast:    false,
						Ts:        "003",
						Name:      "bob",
						Filename:  "second.txt",
						ByteChunk: []byte("bzbzb"),
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   false,
						IsLast:    true,
						Ts:        "003",
						Name:      "cindel",
						Filename:  "third.txt",
						ByteChunk: []byte("czczc"),
					},
				},
			},
			wantTable: map[string]string{
				"alice":  "first.txt",
				"bob":    "second.txt",
				"cindel": "third.txt",
			},
			wantContent: map[string][]byte{
				"alice":  []byte("azaza"),
				"bob":    []byte("bzbzb"),
				"cindel": []byte("czczc"),
			},
		},

		{
			name: "3 unaries mixed",
			ts:   "004",
			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    false,
						Ts:        "004",
						Name:      "alice",
						ByteChunk: []byte("azaza"),
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   false,
						IsLast:    false,
						Ts:        "004",
						Name:      "bob",
						Filename:  "second.txt",
						ByteChunk: []byte("bzbzb"),
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   false,
						IsLast:    true,
						Ts:        "004",
						Name:      "cindel",
						Filename:  "third.txt",
						ByteChunk: []byte("czczc"),
					},
				},
			},
			wantTable: map[string]string{
				"alice":  "azaza",
				"bob":    "second.txt",
				"cindel": "third.txt",
			},
			wantContent: map[string][]byte{
				"bob":    []byte("bzbzb"),
				"cindel": []byte("czczc"),
			},
		},

		{
			name: "1 stream 2 parts correct order",
			ts:   "005",
			reqs: []repo.Request{
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								IsFirst:   true,
								Ts:        "005",
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
								Ts:        "005",
								FieldName: "alice",
								Number:    uint32(0),
								ByteChunk: []byte("azaza"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "005",
								FieldName: "alice",
								Number:    uint32(1),
								ByteChunk: []byte("bzbzbz"),
								IsLast:    true,
							},
						},
					},
				},
			},
			wantTable: map[string]string{
				"alice": "first.txt",
			},
			wantContent: map[string][]byte{
				"alice": []byte("azazabzbzbz"),
			},
		},

		{
			name: "1 stream 2 parts shuffled",
			ts:   "006",
			reqs: []repo.Request{
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								IsFirst:   true,
								Ts:        "006",
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
								Ts:        "006",
								FieldName: "alice",
								Number:    uint32(1),
								ByteChunk: []byte("bzbzbz"),
								IsLast:    true,
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "006",
								FieldName: "alice",
								Number:    uint32(0),
								ByteChunk: []byte("azaza"),
								IsLast:    false,
							},
						},
					},
				},
			},
			wantTable: map[string]string{
				"alice": "first.txt",
			},
			wantContent: map[string][]byte{
				"alice": []byte("azazabzbzbz"),
			},
		},
		{
			name: "2 streams 2 parts correct order",
			ts:   "007",
			reqs: []repo.Request{
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								IsFirst:   true,
								Ts:        "007",
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
								Ts:        "007",
								FieldName: "alice",
								Number:    uint32(0),
								ByteChunk: []byte("azaza"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "007",
								FieldName: "alice",
								Number:    uint32(1),
								ByteChunk: []byte("bzbzbz"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								Ts:        "007",
								FieldName: "bob",
								FileName:  "second.txt",
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "007",
								FieldName: "bob",
								Number:    uint32(0),
								ByteChunk: []byte("11111"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "007",
								FieldName: "bob",
								Number:    uint32(1),
								ByteChunk: []byte("22222"),
								IsLast:    true,
							},
						},
					},
				},
			},
			wantTable: map[string]string{
				"alice": "first.txt",
				"bob":   "second.txt",
			},
			wantContent: map[string][]byte{
				"alice": []byte("azazabzbzbz"),
				"bob":   []byte("1111122222"),
			},
		},

		{
			name: "2 streams 2 parts incorrect order",
			ts:   "007",
			reqs: []repo.Request{
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								IsFirst:   true,
								Ts:        "007",
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
								Ts:        "007",
								FieldName: "alice",
								Number:    uint32(1),
								ByteChunk: []byte("bzbzbz"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "007",
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
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								IsFirst:   true,
								Ts:        "007",
								FieldName: "bob",
								FileName:  "second.txt",
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "007",
								FieldName: "bob",
								Number:    uint32(1),
								ByteChunk: []byte("22222"),
								IsLast:    true,
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "007",
								FieldName: "bob",
								Number:    uint32(0),
								ByteChunk: []byte("11111"),
								IsLast:    false,
							},
						},
					},
				},
			},
			wantTable: map[string]string{
				"alice": "first.txt",
				"bob":   "second.txt",
			},
			wantContent: map[string][]byte{
				"alice": []byte("azazabzbzbz"),
				"bob":   []byte("1111122222"),
			},
		},

		{
			name: "2 streams 2 parts correct order && 3 unaries between them",
			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    false,
						Ts:        "009",
						Name:      "claire",
						ByteChunk: []byte("czczc"),
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								Ts:        "009",
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
								Ts:        "009",
								FieldName: "alice",
								Number:    uint32(0),
								ByteChunk: []byte("azaza"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "009",
								FieldName: "alice",
								Number:    uint32(1),
								ByteChunk: []byte("bzbzbz"),
							},
						},
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						Ts:        "009",
						Name:      "david",
						ByteChunk: []byte("dzdzd"),
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								Ts:        "009",
								FieldName: "bob",
								FileName:  "second.txt",
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "009",
								FieldName: "bob",
								Number:    uint32(0),
								ByteChunk: []byte("11111"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "009",
								FieldName: "bob",
								Number:    uint32(1),
								ByteChunk: []byte("22222"),
							},
						},
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsLast:    true,
						Ts:        "009",
						Name:      "erin",
						ByteChunk: []byte("ezeze"),
					},
				},
			},
			wantTable: map[string]string{
				"claire": "czczc",
				"alice":  "first.txt",
				"david":  "dzdzd",
				"bob":    "second.txt",
				"erin":   "ezeze",
			},
			wantContent: map[string][]byte{
				"alice": []byte("azazabzbzbz"),
				"bob":   []byte("1111122222"),
			},
		},

		{
			name: "2 streams 2 parts incorrect order && 3 unaries between them",
			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    false,
						Ts:        "010",
						Name:      "claire",
						ByteChunk: []byte("czczc"),
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								Ts:        "010",
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
								Ts:        "010",
								FieldName: "alice",
								Number:    uint32(1),
								ByteChunk: []byte("bzbzbz"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "010",
								FieldName: "alice",
								Number:    uint32(0),
								ByteChunk: []byte("azaza"),
							},
						},
					},
				},

				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						Ts:        "010",
						Name:      "david",
						ByteChunk: []byte("dzdzd"),
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								Ts:        "010",
								FieldName: "bob",
								FileName:  "second.txt",
							},
						},
					},
				},

				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "010",
								FieldName: "bob",
								Number:    uint32(1),
								ByteChunk: []byte("22222"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "010",
								FieldName: "bob",
								Number:    uint32(0),
								ByteChunk: []byte("11111"),
							},
						},
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsLast:    true,
						Ts:        "010",
						Name:      "erin",
						ByteChunk: []byte("ezeze"),
					},
				},
			},
			wantTable: map[string]string{
				"claire": "czczc",
				"alice":  "first.txt",
				"david":  "dzdzd",
				"bob":    "second.txt",
				"erin":   "ezeze",
			},
			wantContent: map[string][]byte{
				"alice": []byte("azazabzbzbz"),
				"bob":   []byte("1111122222"),
			},
		},

		{
			name: "2 streams 2 parts incorrect order && 3 unaries between them && one unary of different ts",
			reqs: []repo.Request{
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsFirst:   true,
						IsLast:    false,
						Ts:        "010",
						Name:      "claire",
						ByteChunk: []byte("czczc"),
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								Ts:        "011",
								FieldName: "alice",
								FileName:  "first.txt",
								IsFirst:   true,
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "011",
								FieldName: "alice",
								Number:    uint32(1),
								ByteChunk: []byte("bzbzbz"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "011",
								FieldName: "alice",
								Number:    uint32(0),
								ByteChunk: []byte("azaza"),
							},
						},
					},
				},

				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						Ts:        "011",
						Name:      "david",
						ByteChunk: []byte("dzdzd"),
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileInfo{
							FileInfo: &pb.FileInfo{
								Ts:        "011",
								FieldName: "bob",
								FileName:  "second.txt",
							},
						},
					},
				},

				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "011",
								FieldName: "bob",
								Number:    uint32(1),
								ByteChunk: []byte("22222"),
							},
						},
					},
				},
				&repo.ReqStream{
					R: &pb.FileUploadReq{
						Info: &pb.FileUploadReq_FileData{
							FileData: &pb.FileData{
								Ts:        "011",
								FieldName: "bob",
								Number:    uint32(0),
								ByteChunk: []byte("11111"),
							},
						},
					},
				},
				&repo.ReqUnary{
					R: &pb.TextFieldReq{
						IsLast:    true,
						Ts:        "011",
						Name:      "erin",
						ByteChunk: []byte("ezeze"),
					},
				},
			},
			wantTable: map[string]string{
				"alice": "first.txt",
				"david": "dzdzd",
				"bob":   "second.txt",
				"erin":  "ezeze",
			},
			wantContent: map[string][]byte{
				"alice": []byte("azazabzbzbz"),
				"bob":   []byte("1111122222"),
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			generatorChan <- v.reqs
			time.Sleep(time.Millisecond * 100)
			gotTable, gotContent, gotError := ResultExamine(v.reqs)
			if v.wantError != nil {
				s.Equal(v.wantError, gotError)
			}
			s.Equal(v.wantTable, gotTable)
			s.Equal(v.wantContent, gotContent)

			//time.Sleep(time.Minute * 20)

		})
	}

}

func (g *generator) generate(genChan chan []repo.Request) {
	for i := range genChan {
		for j, v := range i {
			if v.IsUnary() {
				if g.stream != nil {
					_, err := g.stream.CloseAndRecv()
					if err != nil && err != io.EOF {
						logger.L.Errorf("in main.TestWorkFlow request %v failed to close stream: %v\n", v, err)
					}
				}

				req, ok := v.Unwrap().(*pb.TextFieldReq)
				if !ok {
					logger.L.Errorf("in main.TestWorkFlow failed to convert %v to *TextFieldReq\n", v)
				}
				_, err := g.c.SinglePart(context.Background(), req)
				if err != nil {
					logger.L.Errorf("in main.TestWorkFlow faileg to send %v: %v\n", req, err)
				}

			}
			if v.IsStreamInfo() {
				if g.stream != nil {
					_, err := g.stream.CloseAndRecv()
					if err != nil && err != io.EOF {
						logger.L.Errorf("in main.TestWorkFlow request %v failed to close stream: %v\n", v, err)
					}
				}

				req, ok := v.Unwrap().(*pb.FileUploadReq)
				if !ok {
					logger.L.Errorf("in main.TestWorkFlow failed to convert %v to *FileUploadReq\n", v)
				}
				stream, err := g.c.MultiPart(context.Background())
				if err != nil {
					logger.L.Errorf("in main.TestWorkFlow for request %v failed to open stream: %v\n", req, err)
				}
				g.stream = stream

				err = stream.Send(req)
				if err != nil {
					logger.L.Errorf("in main.TestWorkFlow request %v failed to send to stream: %v\n", req, err)
				}
			}
			if v.IsStreamData() {
				req, ok := v.Unwrap().(*pb.FileUploadReq)
				if !ok {
					logger.L.Errorf("in main.TestWorkFlow failed to convert %v to *FileUploadReq_FileData\n", v)
				}
				err := g.stream.Send(req)
				if err != nil {
					logger.L.Errorf("in main.TestWorkFlow request %v failed to send to stream: %v\n", req, err)
				}
				if j == len(i)-1 {
					if g.stream != nil {
						_, err := g.stream.CloseAndRecv()
						if err != nil && err != io.EOF {
							logger.L.Errorf("in main.TestWorkFlow request %v failed to close stream: %v\n", v, err)
						}
					}
				}
			}

		}
	}
}

type generator struct {
	c       pb.SaverClient
	stream  pb.Saver_MultiPartClient
	genChan chan []repo.Request
}

func newGenerator() (*generator, chan []repo.Request) {

	connTest, err := grpc.Dial("localhost:3100", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		logger.L.Errorf("in main.TestWorkFlow err: %v\n", err)
	}

	generatorChan := make(chan []repo.Request, 0)

	return &generator{
			c:       pb.NewSaverClient(connTest),
			genChan: generatorChan,
		},
		generatorChan
}
func ResultExamine(reqs []repo.Request) (map[string]string, map[string][]byte, error) {
	ts := ""
	if len(reqs) == 0 {
		return nil, nil, fmt.Errorf("in main.ResultExamine passed zero len request slice")
	}
	for _, v := range reqs {
		if v.IsLast() {
			ts = v.TS()
		}
	}
	rootPath := "results" + "/" + ts
	tm, fcm := make(map[string]string), make(map[string][]byte)
	_, err := os.Stat(rootPath)
	if err != nil {
		return nil, nil, fmt.Errorf("in main.ResultExamine unable to check \"%s\": %v", rootPath, err)
	}
	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("in main.ResultExamine unable to walk through %q: %v\n", rootPath, err)
		}

		if !d.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("in main.ResultExamine unable to open %q: %v\n", path, err)
			}
			filaName := d.Name()[:strings.Index(d.Name(), ".")]
			fileExt := d.Name()[len(filaName)+1:]

			//if repo.IsTS(d.Name()) {
			if filaName == ts && fileExt == "json" {

				//logger.L.Infof("in main.ResultExamine trying to decode \"%s\"\n", path)
				d := json.NewDecoder(f)
				err = d.Decode(&tm)
				if err != nil {
					return fmt.Errorf("in main.ResultExamine unable to decode file \"%s\": %v\n", path, err)
				}
				//logger.L.Infof("in main.ResultExamine \"%s\": decoded into %v len %d\n", path, tm, len(tm))

				f.Close()
				return nil
			}
			// other files
			bs, err := ioutil.ReadAll(f)
			if err != nil {
				return fmt.Errorf("in main.ResultExamine unable to read file \"%s\": %v\n", path, err)
			}
			for _, v := range reqs {
				if v.FileName() == d.Name() {
					fcm[v.Name()] = append(fcm[v.Name()], bs...)
				}

			}
			f.Close()

		}
		//logger.L.Infof("in main.ResultExamine fcm: %q\n", fcm)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return tm, fcm, nil

}
