package repo

import (
	"log"
	"postSaver/internal/adapters/driver/rpc/pb"
	"testing"

	"github.com/stretchr/testify/suite"
)

type repoSuite struct {
	suite.Suite
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(repoSuite))
}

func (s *repoSuite) TestAdd() {
	tt := []struct {
		name    string
		ids     IDsToRemove
		id      IDToRemove
		wantIDS IDsToRemove
	}{
		{
			name: "Add first",
			ids:  IDsToRemove{},
			id: IDToRemove{
				TS: "qqq",
				I:  1,
			},
			wantIDS: IDsToRemove{
				TS: "qqq",
				I:  []int{1},
			},
		},

		{
			name: "Add fifth",
			ids: IDsToRemove{
				TS: "qqq",
				I:  []int{1, 2, 3, 4},
			},
			id: IDToRemove{
				TS: "qqq",
				I:  5,
			},
			wantIDS: IDsToRemove{
				TS: "qqq",
				I:  []int{1, 2, 3, 4, 5},
			},
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {
			v.ids.Add(v.id)
			s.Equal(v.wantIDS, v.ids)
		})
	}

}

func (s *repoSuite) TestUnwrap() {
	req := NewReqUnary(&pb.TextFieldReq{
		IsFirst:   false,
		IsLast:    false,
		Ts:        "qqq",
		Name:      "bob",
		Filename:  "second.txt",
		ByteChunk: []byte("bzbzb"),
	})
	log.Printf("in repo.TestUnwrap unwrapped req: %T %v\n", req.Unwrap(), req.Unwrap())
}
