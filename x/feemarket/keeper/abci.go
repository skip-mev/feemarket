package keeper

import (
	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlock returns an beginblocker for the x/feemarket module.
func (k *Keeper) BeginBlock(ctx sdk.Context) ([]cometabci.ValidatorUpdate, error) {
	handler := k.plugin.BeginBlockUpdateHandler(ctx)

	err := handler(ctx)

	return []cometabci.ValidatorUpdate{}, err
}
