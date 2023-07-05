package application

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type applicationSuite struct {
	suite.Suite
}

func TestApplicationSuite(t *testing.T) {
	suite.Run(t, new(applicationSuite))
}
