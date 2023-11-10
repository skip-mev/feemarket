package keeper

import (
	"fmt"

	"github.com/skip-mev/feemarket/x/feemarket/interfaces"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	// plugin is the fee market implementation to be used.
	plugin interfaces.FeeMarketImplementation

	// The address that is capable of executing a MsgParams message.
	// Typically, this will be the governance module's address.
	authority string
}

// NewKeeper is a wrapper around NewKeeperWithRewardsAddressProvider for backwards compatibility.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	plugin interfaces.FeeMarketImplementation,
	authority string,
) *Keeper {
	return &Keeper{
		cdc,
		storeKey,
		plugin,
		authority,
	}
}

// Logger returns a auction module-specific logger.
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetAuthority returns the address that is capable of executing a MsgUpdateParams message.
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Plugin returns the plugged fee market implementation of the keeper.
func (k *Keeper) Plugin() interfaces.FeeMarketImplementation {
	return k.plugin
}

// SetData sets arbitrary byte data in the keeper.
func (k *Keeper) SetData(ctx sdk.Context, data []byte) {
	// TODO: limit max data size?

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyData, data)
}

// GetData gets arbitrary byte data in the keeper.
func (k *Keeper) GetData(ctx sdk.Context) ([]byte, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyData)

	if len(bz) == 0 {
		return nil, fmt.Errorf("no data set in the keeper")
	}

	return bz, nil
}

// GetParams returns the feemarket module's parameters.
func (k *Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	store := ctx.KVStore(k.storeKey)

	key := types.KeyParams
	bz := store.Get(key)

	if len(bz) == 0 {
		return types.Params{}, fmt.Errorf("no params found for the feemarket module")
	}

	params := types.Params{}
	if err := params.Unmarshal(bz); err != nil {
		return types.Params{}, err
	}

	return params, nil
}

// SetParams sets the feemarket module's parameters.
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)

	bz, err := params.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.KeyParams, bz)

	return nil
}
