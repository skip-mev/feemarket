package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

// InitGenesis initializes the feemarket module's state from a given genesis state.
func (k *Keeper) InitGenesis(ctx sdk.Context, gs types.GenesisState) {
	if err := gs.Params.ValidateBasic(); err != nil {
		panic(err)
	}

	// Set the feemarket module's parameters.
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	// set the fee market implementation
	if err := k.plugin.Unmarshal(gs.Plugin); err != nil {
		panic(err)
	}

	if err := k.plugin.ValidateBasic(); err != nil {
		panic(err)
	}

	if err := k.Init(ctx); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context.
func (k *Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Get the feemarket module's parameters.
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(k.plugin, params)
}
