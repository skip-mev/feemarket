package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

// InitGenesis initializes the feemarket module's state from a given genesis state.
func (k *Keeper) InitGenesis(ctx sdk.Context, gs types.GenesisState) {
	if err := gs.ValidateBasic(); err != nil {
		panic(err)
	}

	if gs.Params.WindowSize != uint64(len(gs.State.Window)) {
		panic("genesis state and parameters do not match for window")
	}

	maxGas, err := k.GetMaxGasUtilization(ctx)
	if err != nil {
		panic("unable to get consensus params. ensure that the feemarket module is initialized after the consensus params module")
	}

	maxUtilization := uint64(maxGas)
	if gs.Params.TargetBlockUtilization > maxUtilization {
		panic(fmt.Sprintf("target block size of %d cannot be greater than max block size of %d", gs.Params.TargetBlockUtilization, maxUtilization))
	}

	if maxUtilization/gs.Params.TargetBlockUtilization > types.MaxBlockUtilizationRatio {
		panic(fmt.Sprintf("max block size of %d cannot be greater than target block of %d size times %d",
			maxUtilization,
			gs.Params.TargetBlockUtilization,
			types.MaxBlockUtilizationRatio,
		))
	}
	// Initialize the fee market state and parameters.
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	if err := k.SetState(ctx, gs.State); err != nil {
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

	// Get the feemarket module's state.
	state, err := k.GetState(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(params, state)
}
