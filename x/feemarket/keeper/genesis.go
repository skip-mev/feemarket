package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

// InitGenesis initializes the feemarket module's state from a given genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) {
	if err := gs.ValidateBasic(); err != nil {
		panic(err)
	}

	if gs.Params.Window != uint64(len(gs.State.Window)) {
		panic("genesis state and parameters do not match for window")
	}

	// Initialize the fee market state and parameters.
	if err := k.SetParams(sdk.UnwrapSDKContext(ctx), gs.Params); err != nil {
		panic(err)
	}

	if err := k.SetState(sdk.UnwrapSDKContext(ctx), gs.State); err != nil {
		panic(err)
	}

	// always init enabled height to -1 until it is explicitly set later in the application
	if err := k.SetEnabledHeight(sdk.UnwrapSDKContext(ctx), -1); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context.
func (k *Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	// Get the feemarket module's parameters.
	params, err := k.GetParams(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		panic(err)
	}

	// Get the feemarket module's state.
	state, err := k.GetState(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(params, state)
}
