package saver

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type saverSuite struct {
	suite.Suite
}

func TestSaverSuite(t *testing.T) {
	suite.Run(t, new(saverSuite))
}
