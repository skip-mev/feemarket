package mock

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

var _ types.FeeMarketImplementation = &MockFeeMarket{}

type MockFeeMarket struct{} //nolint

// ValidateBasic is a no-op.
func (fm *MockFeeMarket) ValidateBasic() error {
	return nil
}

// Init which initializes the fee market (in InitGenesis)
func (fm *MockFeeMarket) Init(_ sdk.Context) error {
	return nil
}

// Export which exports the fee market (in ExportGenesis)
func (fm *MockFeeMarket) Export(_ sdk.Context) (json.RawMessage, error) {
	return nil, nil
}

// BeginBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the BeginBlock chain.
func (fm *MockFeeMarket) BeginBlockUpdateHandler(_ sdk.Context) types.UpdateHandler {
	return func(ctx sdk.Context) error {
		return nil
	}
}

// EndBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the EndBlock chain.
func (fm *MockFeeMarket) EndBlockUpdateHandler(_ sdk.Context) types.UpdateHandler {
	return func(ctx sdk.Context) error {
		return nil
	}
}

// GetFeeMarketInfo retrieves the fee market's information about
// how to pay for a transaction (min gas price, min tip,
// where the fees are being distributed, etc.).
func (fm *MockFeeMarket) GetFeeMarketInfo(_ sdk.Context) map[string]string {
	return nil
}

// GetID returns the identifier of the fee market
func (fm *MockFeeMarket) GetID() string {
	return "mock"
}

// FeeAnteHandler will be called in the module AnteHandler.
// Performs no actions.
func (fm *MockFeeMarket) FeeAnteHandler(
	_ sdk.Context,
	_ sdk.Tx,
	_ bool,
	next sdk.AnteHandler,
) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return next(ctx, tx, simulate)
	}
}

// FeePostHandler will be called in the module PostHandler
// if PostHandlers are implemented. Performs no actions.
func (fm *MockFeeMarket) FeePostHandler(
	_ sdk.Context,
	_ sdk.Tx,
	_,
	_ bool,
	next sdk.PostHandler,
) sdk.PostHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate, success bool) (newCtx sdk.Context, err error) {
		return next(ctx, tx, simulate, success)
	}
}
