package integration_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// QueryTestSuite is a test suite for query tests
type NetworkTestSuite struct {
	networksuite.NetworkTestSuite
}

// TestQueryTestSuite runs test of the query suite
func TestNetworkTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkTestSuite))
}
