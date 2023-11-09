package mock_test

import (
	"github.com/skip-mev/feemarket/x/feemarket/interfaces"
	"github.com/skip-mev/feemarket/x/feemarket/plugins/mock"
)

// type assertion in test to prevent import cycle
var _ interfaces.FeeMarketImplementation = &mock.MockFeeMarket{}
