package saver

import (
	"os"
	"postSaver/internal/adapters/driver/rpc/pb"
	"postSaver/internal/logger"
	"postSaver/internal/repo"
	"testing"

	"github.com/stretchr/testify/suite"
)

type saverSuite struct {
	suite.Suite
}

func TestSaverSuite(t *testing.T) {
	suite.Run(t, new(saverSuite))
}

func (s *saverSuite) TestFileCreate() {
	tt := []struct {
		name      string
		saver     Saver
		r         repo.Request
		wantError error
		wantPath  string
	}{
		{
			name: "1",
			saver: &SaverStruct{
				Path: "../../../../results",
				F:    map[string]*os.File{},
			},
			r: &repo.ReqStream{
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
			wantPath: "../../../../results/qqq/first.txt",
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			path, err := v.saver.FileCreate(v.r)
			if err != nil {
				s.Equal(v.wantError, err)
			}
			s.Equal(v.wantPath, path)
			logger.L.Infof("in saver.TestFileCreate F: %v\n", v.saver)
		})
	}
}

func (s *saverSuite) TestFileWrite() {
	tt := []struct {
		name      string
		saver     Saver
		rCreate   repo.Request
		rWrite    []repo.Request
		wantError error
	}{
		{
			name: "1",
			saver: &SaverStruct{
				Path: "../../../../results",
				F:    map[string]*os.File{},
			},
			rCreate: &repo.ReqStream{
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
			rWrite: []repo.Request{
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
								Number:    uint32(0),
								ByteChunk: []byte("bzbzb"),
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
			v.saver.FileCreate(v.rCreate)
			for _, w := range v.rWrite {
				v.saver.FileWrite(w)
			}
			v.saver.FileClose(v.rCreate)
		})
	}
}

func (s *saverSuite) TestSimplify() {
	tt := []struct {
		name            string
		m               map[string]repo.NameNumber
		wantMSimplified map[string]string
	}{
		{
			name: "1",
			m: map[string]repo.NameNumber{
				"alice": {
					Name: "azaza",
				},
				"bob": {
					Name:   "first.txt",
					Number: 2,
				},
			},
			wantMSimplified: map[string]string{
				"alice": "azaza",
				"bob":   "first.txt",
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			gotSimplified := simplify(v.m)
			s.Equal(v.wantMSimplified, gotSimplified)
		})
	}
}

func (s *saverSuite) TestSaveTable() {
	tt := []struct {
		name      string
		sv        Saver
		m         map[string]repo.NameNumber
		ts        string
		wantError error
	}{
		{
			name: "1",
			sv: &SaverStruct{
				Path: "../../../../results",
			},
			m: map[string]repo.NameNumber{
				"alice": {
					Name: "azaza",
				},
				"bob": {
					Name:   "first.txt",
					Number: 2,
				},
			},
			ts: "qqq",
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			simplidied := simplify(v.m)
			gotError := v.sv.tableSave(simplidied, v.ts)
			if gotError != nil {
				s.Equal(v.wantError, gotError)
			}
		})
	}
}
