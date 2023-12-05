package keeper

import (
	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/skip-mev/feemarket/testutils/sample"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

var moduleAccountPerms = map[string][]string{
	authtypes.FeeCollectorName:      nil,
	distrtypes.ModuleName:           nil,
	minttypes.ModuleName:            {authtypes.Minter},
	stakingtypes.BondedPoolName:     {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:  {authtypes.Burner, authtypes.Staking},
	feemarkettypes.ModuleName:       nil,
	feemarkettypes.FeeCollectorName: {authtypes.Burner},
}

// initializer allows to initialize each module keeper
type initializer struct {
	Codec      codec.Codec
	Amino      *codec.LegacyAmino
	DB         *tmdb.MemDB
	StateStore store.CommitMultiStore
}

func newInitializer() initializer {
	db := tmdb.NewMemDB()
	return initializer{
		DB:         db,
		Codec:      sample.Codec(),
		StateStore: store.NewCommitMultiStore(db),
	}
}

// ModuleAccountAddrs returns all the app's module account addresses.
func ModuleAccountAddrs(maccPerms map[string][]string) map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

func (i *initializer) Param() paramskeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	tkeys := sdk.NewTransientStoreKey(paramstypes.TStoreKey)

	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)
	i.StateStore.MountStoreWithDB(tkeys, storetypes.StoreTypeTransient, i.DB)

	return paramskeeper.NewKeeper(
		i.Codec,
		i.Amino,
		storeKey,
		tkeys,
	)
}

func (i *initializer) Auth(paramKeeper paramskeeper.Keeper) authkeeper.AccountKeeper {
	storeKey := sdk.NewKVStoreKey(authtypes.StoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)
	paramKeeper.Subspace(authtypes.ModuleName)

	return authkeeper.NewAccountKeeper(
		i.Codec,
		storeKey,
		authtypes.ProtoBaseAccount,
		moduleAccountPerms,
		sdk.Bech32PrefixAccAddr,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

func (i *initializer) Bank(paramKeeper paramskeeper.Keeper, authKeeper authkeeper.AccountKeeper) bankkeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(banktypes.StoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)
	paramKeeper.Subspace(banktypes.ModuleName)
	modAccAddrs := ModuleAccountAddrs(moduleAccountPerms)

	return bankkeeper.NewBaseKeeper(
		i.Codec,
		storeKey,
		authKeeper,
		modAccAddrs,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

// create mock ProtocolVersionSetter for UpgradeKeeper

type ProtocolVersionSetter struct{}

func (vs ProtocolVersionSetter) SetProtocolVersion(uint64) {}

func (i *initializer) Upgrade() *upgradekeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(upgradetypes.StoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)

	skipUpgradeHeights := make(map[int64]bool)
	vs := ProtocolVersionSetter{}

	return upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		storeKey,
		i.Codec,
		"",
		vs,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

func (i *initializer) Staking(
	authKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	paramKeeper paramskeeper.Keeper,
) *stakingkeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(stakingtypes.StoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)
	paramKeeper.Subspace(stakingtypes.ModuleName)

	return stakingkeeper.NewKeeper(
		i.Codec,
		storeKey,
		authKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

func (i *initializer) Distribution(
	authKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
) distrkeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(distrtypes.StoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)

	return distrkeeper.NewKeeper(
		i.Codec,
		storeKey,
		authKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

func (i *initializer) FeeMarket(
	authKeeper authkeeper.AccountKeeper,
) *feemarketkeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(feemarkettypes.StoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)

	return feemarketkeeper.NewKeeper(
		i.Codec,
		storeKey,
		authKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

func (i *initializer) FeeGrant(
	authKeeper authkeeper.AccountKeeper,
) feegrantkeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(feegrant.StoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)

	return feegrantkeeper.NewKeeper(
		i.Codec,
		storeKey,
		authKeeper,
	)
}
