package keeper

import (
<<<<<<< HEAD
=======
	"fmt"

>>>>>>> 136c8fa (basic keeper funcs)
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	// The address that is capable of executing a MsgParams message.
	// Typically, this will be the governance module's address.
	authority string
}

<<<<<<< HEAD
// NewKeeper constructs a new feemarket keeper.
=======
// NewKeeper is a wrapper around NewKeeperWithRewardsAddressProvider for backwards compatibility.
>>>>>>> 136c8fa (basic keeper funcs)
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
<<<<<<< HEAD
) *Keeper {
	k := &Keeper{
=======
) Keeper {
	return Keeper{
>>>>>>> 136c8fa (basic keeper funcs)
		cdc,
		storeKey,
		authority,
	}
<<<<<<< HEAD

	return k
}

// Logger returns a feemarket module-specific logger.
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
=======
}

// Logger returns a auction module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
>>>>>>> 136c8fa (basic keeper funcs)
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetAuthority returns the address that is capable of executing a MsgUpdateParams message.
<<<<<<< HEAD
func (k *Keeper) GetAuthority() string {
=======
func (k Keeper) GetAuthority() string {
>>>>>>> 136c8fa (basic keeper funcs)
	return k.authority
}

// GetParams returns the feemarket module's parameters.
<<<<<<< HEAD
func (k *Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
=======
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
>>>>>>> 136c8fa (basic keeper funcs)
	store := ctx.KVStore(k.storeKey)

	key := types.KeyParams
	bz := store.Get(key)

<<<<<<< HEAD
=======
	if len(bz) == 0 {
		return types.Params{}, fmt.Errorf("no params found for the feemarket module")
	}

>>>>>>> 136c8fa (basic keeper funcs)
	params := types.Params{}
	if err := params.Unmarshal(bz); err != nil {
		return types.Params{}, err
	}

	return params, nil
}

// SetParams sets the feemarket module's parameters.
<<<<<<< HEAD
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) error {
=======
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
>>>>>>> 136c8fa (basic keeper funcs)
	store := ctx.KVStore(k.storeKey)

	bz, err := params.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.KeyParams, bz)

	return nil
}
