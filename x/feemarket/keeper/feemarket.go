package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ------------------- Fee Market Updates ------------------- //

// Init initializes the fee market (in InitGenesis).
func (k *Keeper) Init(_ sdk.Context) error {
	// TODO initialize fee market state with params

	return nil
}

// Export exports the fee market (in ExportGenesis).
func (k *Keeper) Export(_ sdk.Context) (json.RawMessage, error) {
	// TODO export state from fee market state

	return nil, nil
}

// BeginBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the BeginBlock chain.
func (k *Keeper) BeginBlockUpdateHandler(_ sdk.Context) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		return nil // TODO return handler
	}
}

// EndBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the EndBlock chain.
func (k *Keeper) EndBlockUpdateHandler(_ sdk.Context) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		return nil // TODO return handler
	}
}

// ------------------- Fee Market Queries ------------------- //

// TODO add fee market state query
