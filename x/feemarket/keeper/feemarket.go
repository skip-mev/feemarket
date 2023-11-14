package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UpdateHandler is responsible for updating the parameters of the
// fee market plugin.
// Fees can optionally also be extracted here.
type UpdateHandler func(ctx sdk.Context) error

// ------------------- Fee Market Updates ------------------- //

// Init initializes the fee market (in InitGenesis).
func (k *Keeper) Init(ctx sdk.Context) error {
	// TODO initialize fee market state with params

	return nil
}

// Export exports the fee market (in ExportGenesis).
func (k *Keeper) Export(ctx sdk.Context) (json.RawMessage, error) {
	// TODO export state from fee market state

	return nil, nil
}

// BeginBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the BeginBlock chain.
func (k *Keeper) BeginBlockUpdateHandler(ctx sdk.Context) UpdateHandler {
	return func(ctx sdk.Context) error {
		return nil // TODO return handler
	}
}

// EndBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the EndBlock chain.
func (k *Keeper) EndBlockUpdateHandler(ctx sdk.Context) UpdateHandler {
	return func(ctx sdk.Context) error {
		return nil // TODO return handler
	}
}

// ------------------- Fee Market Queries ------------------- //

// TODO add fee market state query
