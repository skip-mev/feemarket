package keeper

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

// Keeper is the x/feemarket keeper.
type Keeper struct {
	appmodule.Environment
	cdc      codec.BinaryCodec
	ak       types.AccountKeeper
	resolver types.DenomResolver

	// The address that is capable of executing a MsgParams message.
	// Typically, this will be the governance module's address.
	authority string
}

// NewKeeper constructs a new feemarket keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	env appmodule.Environment,
	authKeeper types.AccountKeeper,
	resolver types.DenomResolver,
	authority string,
) *Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	k := &Keeper{
		Environment: env,
		cdc:         cdc,
		ak:          authKeeper,
		resolver:    resolver,
		authority:   authority,
	}

	return k
}

// Logger returns a feemarket module-specific logger.
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetAuthority returns the address that is capable of executing a MsgUpdateParams message.
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// GetEnabledHeight returns the height at which the feemarket was enabled.
func (k *Keeper) GetEnabledHeight(ctx sdk.Context) (int64, error) {
	store := k.Environment.KVStoreService.OpenKVStore(ctx)

	key := types.KeyEnabledHeight
	bz, err := store.Get(key)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return -1, nil
	}

	return strconv.ParseInt(string(bz), 10, 64)
}

// SetEnabledHeight sets the height at which the feemarket was enabled.
func (k *Keeper) SetEnabledHeight(ctx sdk.Context, height int64) {
	store := k.Environment.KVStoreService.OpenKVStore(ctx)

	bz := []byte(strconv.FormatInt(height, 10))

	if err := store.Set(types.KeyEnabledHeight, bz); err != nil {
		panic(err) // TODO(technicallyty): fix this.
	}
}

// ResolveToDenom converts the given coin to the given denomination.
func (k *Keeper) ResolveToDenom(ctx sdk.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error) {
	if k.resolver == nil {
		return sdk.DecCoin{}, types.ErrResolverNotSet
	}

	return k.resolver.ConvertToDenom(ctx, coin, denom)
}

// SetDenomResolver sets the keeper's denom resolver.
func (k *Keeper) SetDenomResolver(resolver types.DenomResolver) {
	k.resolver = resolver
}

// GetState returns the feemarket module's state.
func (k *Keeper) GetState(ctx sdk.Context) (types.State, error) {
	store := k.Environment.KVStoreService.OpenKVStore(ctx)

	key := types.KeyState
	bz, err := store.Get(key)
	if err != nil {
		return types.State{}, err
	}

	state := types.State{}
	if err := state.Unmarshal(bz); err != nil {
		return types.State{}, err
	}

	return state, nil
}

// SetState sets the feemarket module's state.
func (k *Keeper) SetState(ctx sdk.Context, state types.State) error {
	store := k.Environment.KVStoreService.OpenKVStore(ctx)

	bz, err := state.Marshal()
	if err != nil {
		return err
	}

	if err := store.Set(types.KeyState, bz); err != nil {
		return err
	}

	return nil
}

// GetParams returns the feemarket module's parameters.
func (k *Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	store := k.Environment.KVStoreService.OpenKVStore(ctx)

	key := types.KeyParams
	bz, err := store.Get(key)
	if err != nil {
		return types.Params{}, err
	}

	params := types.Params{}
	if err := params.Unmarshal(bz); err != nil {
		return types.Params{}, err
	}

	return params, nil
}

// SetParams sets the feemarket module's parameters.
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := k.Environment.KVStoreService.OpenKVStore(ctx)

	bz, err := params.Marshal()
	if err != nil {
		return err
	}

	if err := store.Set(types.KeyParams, bz); err != nil {
		return err
	}

	return nil
}
