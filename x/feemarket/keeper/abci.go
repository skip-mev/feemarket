package keeper

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlock returns a beginblocker for the x/feemarket module.
func (k *Keeper) BeginBlock(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	handler := k.plugin.BeginBlockUpdateHandler(ctx)

	err := handler(ctx)

	return []abci.ValidatorUpdate{}, err
}

// EndBlock returns an endblocker for the x/feemarket module.
func (k *Keeper) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	handler := k.plugin.EndBlockUpdateHandler(ctx)

	err := handler(ctx)
	if err != nil {
		k.Logger(ctx).Error("error in end block", "error", err)
	}

	return []abci.ValidatorUpdate{}
}