package mocks

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/interfaces"
)

var _ interfaces.FeeMarketImplementation = &PanicMarket{}

// NewPanicMarket returns an instance of a new PanicMarket.
func NewPanicMarket() *PanicMarket {
	return &PanicMarket{
		Data: []byte("panic"),
	}
}

// ValidateBasic is a no-op.
func (fm *PanicMarket) ValidateBasic() error {
	panic("panic feemarket")

	return nil
}

// Init which initializes the fee market (in InitGenesis).
func (fm *PanicMarket) Init(_ sdk.Context) error {
	panic("panic feemarket")

	return nil
}

// Export which exports the fee market (in ExportGenesis).
func (fm *PanicMarket) Export(_ sdk.Context) (json.RawMessage, error) {
	panic("panic feemarket")

	return nil, nil
}

// BeginBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the BeginBlock chain.
func (fm *PanicMarket) BeginBlockUpdateHandler(_ sdk.Context) interfaces.UpdateHandler {
	return func(ctx sdk.Context) error {
		panic("panic feemarket")

		return nil
	}
}

// EndBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the EndBlock chain.
func (fm *PanicMarket) EndBlockUpdateHandler(_ sdk.Context) interfaces.UpdateHandler {
	return func(ctx sdk.Context) error {
		panic("panic feemarket")

		return nil
	}
}

// GetFeeMarketInfo retrieves the fee market's information about
// how to pay for a transaction (min gas price, min tip,
// where the fees are being distributed, etc.).
func (fm *PanicMarket) GetFeeMarketInfo(_ sdk.Context) map[string]string {
	panic("panic feemarket")

	return nil
}

// GetID returns the identifier of the fee market.
func (fm *PanicMarket) GetID() string {
	return "panic"
}

// FeeAnteHandler will be called in the module AnteHandler.
// Performs no actions.
func (fm *PanicMarket) FeeAnteHandler(
	_ sdk.Context,
	_ sdk.Tx,
	_ bool,
	next sdk.AnteHandler,
) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		panic("panic feemarket")

		return next(ctx, tx, simulate)
	}
}

// FeePostHandler will be called in the module PostHandler
// if PostHandlers are implemented. Performs no actions.
func (fm *PanicMarket) FeePostHandler(
	_ sdk.Context,
	_ sdk.Tx,
	_,
	_ bool,
	next sdk.PostHandler,
) sdk.PostHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate, success bool) (newCtx sdk.Context, err error) {
		panic("panic feemarket")

		return next(ctx, tx, simulate, success)
	}
}
