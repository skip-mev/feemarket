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

// BeginBlocker allows the fee market to be updated
// after every block. This will be added to the BeginBlock chain.
func (k *Keeper) BeginBlocker(_ sdk.Context) error {
	return nil
}

// EndBlocker allows the fee market to be updated
// after every block. This will be added to the EndBlock chain.
func (k *Keeper) EndBlocker(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	if !params.Enabled {
		return nil
	}

	state, err := k.GetState(ctx)
	if err != nil {
		return err
	}

	// Update the learning rate based on the block utilization seen in the
	// current block. This is the AIMD learning rate adjustment algorithm.
	newLR := state.UpdateLearningRate(
		params.Theta,
		params.Alpha,
		params.Beta,
		params.MinLearningRate,
		params.MaxLearningRate,
	)

	// Update the base fee based with the new learning rate and delta adjustment.
	newBaseFee := state.UpdateBaseFee(params.Delta)

	k.Logger(ctx).Info(
		"updated the fee market",
		"height", ctx.BlockHeight(),
		"new_base_fee", newBaseFee,
		"new_learning_rate", newLR,
		"average_block_utilization", state.GetAverageUtilization(),
		"net_block_utilization", state.GetNetUtilization(),
	)

	// Increment the height of the state and set the new state.
	state.IncrementHeight()
	return k.SetState(ctx, state)
}
