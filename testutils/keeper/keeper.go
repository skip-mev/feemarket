// Package keeper provides methods to initialize SDK keepers with local storage for test purposes
package keeper

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

var (
	// ExampleTimestamp is a timestamp used as the current time for the context of the keepers returned from the package
	ExampleTimestamp = time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC)

	// ExampleHeight is a block height used as the current block height for the context of test keeper
	ExampleHeight = int64(1111)
)

// TestKeepers holds all keepers used during keeper tests for all modules
type TestKeepers struct {
	T               testing.TB
	AccountKeeper   authkeeper.AccountKeeper
	BankKeeper      bankkeeper.Keeper
	DistrKeeper     distrkeeper.Keeper
	StakingKeeper   *stakingkeeper.Keeper
	FeeMarketKeeper *feemarketkeeper.Keeper
	FeeGrantKeeper  feegrantkeeper.Keeper
}

// TestMsgServers holds all message servers used during keeper tests for all modules
type TestMsgServers struct {
	T                  testing.TB
	FeeMarketMsgServer feemarkettypes.MsgServer
}

// SetupOption represents an option that can be provided to NewTestSetup
type SetupOption func(*setupOptions)

// setupOptions represents the set of SetupOption
type setupOptions struct{}

// NewTestSetup returns initialized instances of all the keepers and message servers of the modules
func NewTestSetup(t testing.TB, options ...SetupOption) (sdk.Context, TestKeepers, TestMsgServers) {
	// setup options
	var so setupOptions
	for _, option := range options {
		option(&so)
	}

	initializer := newInitializer()

	paramKeeper := initializer.Param()
	authKeeper := initializer.Auth(paramKeeper)
	bankKeeper := initializer.Bank(paramKeeper, authKeeper)
	stakingKeeper := initializer.Staking(authKeeper, bankKeeper, paramKeeper)
	distrKeeper := initializer.Distribution(authKeeper, bankKeeper, stakingKeeper)
	feeMarketKeeper := initializer.FeeMarket(authKeeper)
	feeGrantKeeper := initializer.FeeGrant(authKeeper)

	require.NoError(t, initializer.StateStore.LoadLatestVersion())

	// Create a context using a custom timestamp
	ctx := sdk.NewContext(initializer.StateStore, tmproto.Header{
		Time:   ExampleTimestamp,
		Height: ExampleHeight,
	}, false, log.NewNopLogger())

	// initialize params
	err := distrKeeper.SetParams(ctx, distrtypes.DefaultParams())
	if err != nil {
		panic(err)
	}
	err = stakingKeeper.SetParams(ctx, stakingtypes.DefaultParams())
	if err != nil {
		panic(err)
	}
	err = feeMarketKeeper.SetState(ctx, feemarkettypes.DefaultState())
	if err != nil {
		panic(err)
	}
	err = feeMarketKeeper.SetParams(ctx, feemarkettypes.DefaultParams())
	if err != nil {
		panic(err)
	}

	// initialize msg servers
	feeMarketMsgSrv := feemarketkeeper.NewMsgServer(*feeMarketKeeper)

	return ctx,
		TestKeepers{
			T:               t,
			AccountKeeper:   authKeeper,
			BankKeeper:      bankKeeper,
			DistrKeeper:     distrKeeper,
			StakingKeeper:   stakingKeeper,
			FeeMarketKeeper: feeMarketKeeper,
			FeeGrantKeeper:  feeGrantKeeper,
		},
		TestMsgServers{
			T:                  t,
			FeeMarketMsgServer: feeMarketMsgSrv,
		}
}
