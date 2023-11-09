package defaultmarket

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/interfaces"
)

var _ interfaces.FeeMarketImplementation = &DefaultMarket{}

// NewDefaultFeeMarket returns an instance of a new DefaultFeeMarket.
func NewDefaultFeeMarket() *DefaultMarket {
	return &DefaultMarket{
		Data: []byte("default"),
	}
}

// ValidateBasic is a no-op.
func (fm *DefaultMarket) ValidateBasic() error {
	return nil
}

// Init which initializes the fee market (in InitGenesis).
func (fm *DefaultMarket) Init(_ sdk.Context) error {
	return nil
}

// Export which exports the fee market (in ExportGenesis).
func (fm *DefaultMarket) Export(_ sdk.Context) (json.RawMessage, error) {
	return nil, nil
}

// BeginBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the BeginBlock chain.
func (fm *DefaultMarket) BeginBlockUpdateHandler(_ sdk.Context) interfaces.UpdateHandler {
	return func(ctx sdk.Context) error {
		return nil
	}
}

// EndBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the EndBlock chain.
func (fm *DefaultMarket) EndBlockUpdateHandler(_ sdk.Context) interfaces.UpdateHandler {
	return func(ctx sdk.Context) error {
		return nil
	}
}

// GetFeeMarketInfo retrieves the fee market's information about
// how to pay for a transaction (min gas price, min tip,
// where the fees are being distributed, etc.).
func (fm *DefaultMarket) GetFeeMarketInfo(_ sdk.Context) map[string]string {
	return nil
}

// GetID returns the identifier of the fee market.
func (fm *DefaultMarket) GetID() string {
	return "default"
}

// FeeAnteHandler will be called in the module AnteHandler.
// Performs no actions.
func (fm *DefaultMarket) FeeAnteHandler(
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
func (fm *DefaultMarket) FeePostHandler(
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
