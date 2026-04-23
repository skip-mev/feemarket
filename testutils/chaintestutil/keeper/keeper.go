// Package keeper provides methods to initialize SDK keepers with local storage for test purposes
package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log/v2"
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
)

var (
	// ExampleTimestamp is a timestamp used as the current time for the context of the keepers returned from the package
	ExampleTimestamp = time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC)

	// ExampleHeight is a block height used as the current block height for the context of test keeper
	ExampleHeight = int64(1111)
)

// TestKeepers holds all keepers used during keeper tests for all modules
type TestKeepers struct {
	T              testing.TB
	Initializer    *Initializer
	AccountKeeper  authkeeper.AccountKeeper
	BankKeeper     bankkeeper.Keeper
	DistrKeeper    distrkeeper.Keeper
	StakingKeeper  *stakingkeeper.Keeper
	FeeGrantKeeper feegrantkeeper.Keeper
}

// TestMsgServers holds all message servers used during keeper tests for all modules
type TestMsgServers struct {
	T testing.TB
}

// SetupOption represents an option that can be provided to NewTestSetup
type SetupOption func(*SetupOptions)

// SetupOptions represents the options to configure the setup of a keeper-level integration test.
type SetupOptions struct {
	// AdditionalModuleAccountPerms represents any added module account permissions that need to
	// be passed to the keeper initializer
	AdditionalModuleAccountPerms map[string][]string
}

// WithAdditionalModuleAccounts adds additional module accounts to the testing config.
func WithAdditionalModuleAccounts(maccPerms map[string][]string) SetupOption {
	return func(options *SetupOptions) {
		options.AdditionalModuleAccountPerms = maccPerms
	}
}

// NewTestSetup returns initialized instances of all the keepers and message servers of the modules
func NewTestSetup(t testing.TB, options ...SetupOption) (sdk.Context, TestKeepers, TestMsgServers) {
	// run all options before setup
	var so SetupOptions
	for _, option := range options {
		option(&so)
	}

	initializer := newInitializer()

	authKeeper := initializer.Auth(so.AdditionalModuleAccountPerms)
	bankKeeper := initializer.Bank(authKeeper, so.AdditionalModuleAccountPerms)
	stakingKeeper := initializer.Staking(authKeeper, bankKeeper)
	distrKeeper := initializer.Distribution(authKeeper, bankKeeper, stakingKeeper)
	feeGrantKeeper := initializer.FeeGrant(authKeeper)
	require.NoError(t, initializer.LoadLatest())

	// Create a context using a custom timestamp
	ctx := sdk.NewContext(initializer.StateStore, tmproto.Header{
		Time:   ExampleTimestamp,
		Height: ExampleHeight,
	}, false, log.NewNopLogger())

	// initialize params
	err := distrKeeper.Params.Set(ctx, distrtypes.DefaultParams())
	if err != nil {
		panic(err)
	}
	err = stakingKeeper.SetParams(ctx, stakingtypes.DefaultParams())
	if err != nil {
		panic(err)
	}

	return ctx,
		TestKeepers{
			T:              t,
			Initializer:    &initializer,
			AccountKeeper:  authKeeper,
			BankKeeper:     bankKeeper,
			DistrKeeper:    distrKeeper,
			StakingKeeper:  stakingKeeper,
			FeeGrantKeeper: feeGrantKeeper,
		},
		TestMsgServers{
			T: t,
		}
}
