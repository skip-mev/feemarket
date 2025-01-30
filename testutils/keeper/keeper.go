// Package keeper provides methods to initialize SDK keepers with local storage for test purposes
package keeper

import (
	"testing"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/skip-mev/feemarket/x/feemarket"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	distributionkeeper "cosmossdk.io/x/distribution/keeper"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	govtypes "cosmossdk.io/x/gov/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"

	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

// TestKeepers holds all keepers used during keeper tests for all modules
type TestKeepers struct {
	AccountKeeper  authkeeper.AccountKeeper
	BankKeeper     bankkeeper.Keeper
	DistrKeeper    distributionkeeper.Keeper
	StakingKeeper  stakingkeeper.Keeper
	FeeGrantKeeper feegrantkeeper.Keeper

	FeeMarketKeeper *feemarketkeeper.Keeper
}

// TestMsgServers holds all message servers used during keeper tests for all modules
type TestMsgServers struct {
	FeeMarketMsgServer feemarkettypes.MsgServer
}

var additionalMaccPerms = map[string][]string{
	feemarkettypes.ModuleName:       nil,
	feemarkettypes.FeeCollectorName: {authtypes.Burner},
}

// NewTestSetup returns initialized instances of all the keepers and message servers of the modules
func NewTestSetup(t *testing.T) (sdk.Context, TestKeepers, TestMsgServers) {
	tk := TestKeepers{} // TODO(technicallyty): fill this out.
	// initialize extra keeper
	feeMarketKeeper := FeeMarket(tk.AccountKeeper)

	// initialize msg servers
	feeMarketMsgSrv := feemarketkeeper.NewMsgServer(feeMarketKeeper)

	key := storetypes.NewKVStoreKey(feemarkettypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx

	err := feeMarketKeeper.SetState(ctx, feemarkettypes.DefaultState())
	require.NoError(t, err)
	err = feeMarketKeeper.SetParams(ctx, feemarkettypes.DefaultParams())
	require.NoError(t, err)

	testKeepers := TestKeepers{
		FeeMarketKeeper: feeMarketKeeper,
	}

	testMsgServers := TestMsgServers{
		FeeMarketMsgServer: feeMarketMsgSrv,
	}

	return ctx, testKeepers, testMsgServers
}

// FeeMarket initializes the fee market module using the testkeepers intializer.
func FeeMarket(
	authKeeper authkeeper.AccountKeeper,
) *feemarketkeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(feemarkettypes.StoreKey)
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, feemarket.AppModule{})

	return feemarketkeeper.NewKeeper(
		encCfg.Codec,
		storeKey,
		authKeeper,
		&feemarkettypes.TestDenomResolver{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}
