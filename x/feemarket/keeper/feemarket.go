package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/interfaces"
)

// ------------------- Fee Market Parameters ------------------- //

// Init which initializes the fee market (in InitGenesis)
func (k *Keeper) Init(ctx sdk.Context) error {
	return k.plugin.Init(ctx)
}

// Export which exports the fee market (in ExportGenesis)
func (k *Keeper) Export(ctx sdk.Context) (json.RawMessage, error) {
	return k.plugin.Export(ctx)
}

// BeginBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the BeginBlock chain.
func (k *Keeper) BeginBlockUpdateHandler(ctx sdk.Context) interfaces.UpdateHandler {
	return k.plugin.BeginBlockUpdateHandler(ctx)
}

// EndBlockUpdateHandler allows the fee market to be updated
// after every block. This will be added to the EndBlock chain.
func (k *Keeper) EndBlockUpdateHandler(ctx sdk.Context) interfaces.UpdateHandler {
	return k.plugin.EndBlockUpdateHandler(ctx)
}

// ------------------- Fee Market Queries ------------------- //

// GetFeeMarketInfo retrieves the fee market's information about
// how to pay for a transaction (min gas price, min tip,
// where the fees are being distributed, etc.).
func (k *Keeper) GetFeeMarketInfo(ctx sdk.Context) map[string]string {
	return k.plugin.GetFeeMarketInfo(ctx)
}
