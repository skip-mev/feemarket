// Package keeper provides methods to initialize SDK keepers with local storage for test purposes
package keeper

import (
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/skip-mev/chaintestutil/keeper"

	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

// TestKeepers holds all keepers used during keeper tests for all modules
type TestKeepers struct {
	keeper.TestKeepers
	FeeMarketKeeper *feemarketkeeper.Keeper
}

// TestMsgServers holds all message servers used during keeper tests for all modules
type TestMsgServers struct {
	keeper.TestMsgServers
	FeeMarketMsgServer feemarkettypes.MsgServer
}

var additionalMaccPerms = map[string][]string{
	feemarkettypes.ModuleName:       nil,
	feemarkettypes.FeeCollectorName: {authtypes.Burner},
}

// NewTestSetup returns initialized instances of all the keepers and message servers of the modules
func NewTestSetup(t testing.TB, options ...keeper.SetupOption) (sdk.Context, TestKeepers, TestMsgServers) {
	options = append(options, keeper.WithAdditionalModuleAccounts(additionalMaccPerms))

	_, tk, tms := keeper.NewTestSetup(t, options...)

	// initialize extra keeper
	feeMarketKeeper := FeeMarket(tk.Initializer, tk.AccountKeeper)
	require.NoError(t, tk.Initializer.LoadLatest())

	// initialize msg servers
	feeMarketMsgSrv := feemarketkeeper.NewMsgServer(*feeMarketKeeper)

	ctx := sdk.NewContext(tk.Initializer.StateStore, tmproto.Header{
		Time:   keeper.ExampleTimestamp,
		Height: keeper.ExampleHeight,
	}, false, log.NewNopLogger())

	err := feeMarketKeeper.SetState(ctx, feemarkettypes.DefaultState())
	require.NoError(t, err)
	err = feeMarketKeeper.SetParams(ctx, feemarkettypes.DefaultParams())
	require.NoError(t, err)

	testKeepers := TestKeepers{
		TestKeepers:     tk,
		FeeMarketKeeper: feeMarketKeeper,
	}

	testMsgServers := TestMsgServers{
		TestMsgServers:     tms,
		FeeMarketMsgServer: feeMarketMsgSrv,
	}

	return ctx, testKeepers, testMsgServers
}

func FeeMarket(
	initializer *keeper.Initializer,
	authKeeper authkeeper.AccountKeeper,
) *feemarketkeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(feemarkettypes.StoreKey)
	initializer.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, initializer.DB)

	return feemarketkeeper.NewKeeper(
		initializer.Codec,
		storeKey,
		authKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}
