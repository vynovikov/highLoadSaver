package repo

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type repoSuite struct {
	suite.Suite
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(repoSuite))
}
