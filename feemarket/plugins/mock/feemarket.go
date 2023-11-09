package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/feemarket"
)

var _ feemarket.FeeMarket = FeeMarket{}

// FeeMarket is a simple mock fee market implmentation that should only be used for testing.
type FeeMarket struct{}

// Init which initializes the fee market (in InitGenesis)
func (fm FeeMarket) Init(_ sdk.Context) error {
	return nil
}

// EndBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the EndBlock chain.
func (fm FeeMarket) EndBlockUpdateHandler(_ sdk.Context) feemarket.UpdateHandler {
	return nil
}

// EpochUpdateHandler allows the fee market to be updated
// after every given epoch identifier. This maps the epoch
// identifier to the UpdateHandler that should be executed.
func (fm FeeMarket) EpochUpdateHandler(_ sdk.Context) map[string]feemarket.UpdateHandler {
	return nil
}

// GetMinGasPrice retrieves the minimum gas price(s) needed
// to be included in the block for the given transaction
func (fm FeeMarket) GetMinGasPrice(_ sdk.Context, _ sdk.Tx) sdk.Coins {
	return sdk.NewCoins()
}

// GetFeeMarketInfo retrieves the fee market's information about
// how to pay for a transaction (min gas price, min tip,
// where the fees are being distributed, etc.).
func (fm FeeMarket) GetFeeMarketInfo(_ sdk.Context) map[string]string {
	return nil
}

// GetID returns the identifier of the fee market
func (fm FeeMarket) GetID() string {
	return "mock"
}

// FeeAnteHandler will be called in the module AnteHandler.
// Performs no actions.
func (fm FeeMarket) FeeAnteHandler(
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
func (fm FeeMarket) FeePostHandler(
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
