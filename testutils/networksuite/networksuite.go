// Package networksuite provides a base test suite for tests that need a local network instance
package networksuite

import (
	"math/rand"

	tmdb "github.com/cometbft/cometbft-db"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/pruning/types"
	sdknetwork "github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/gogoproto/proto"

	"github.com/skip-mev/chaintestutil/network"
	"github.com/skip-mev/chaintestutil/sample"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/skip-mev/feemarket/tests/app"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

var (
	chainID = "chain-" + tmrand.NewRand().Str(6)

	DefaultAppConstructor = func(val sdknetwork.ValidatorI) servertypes.Application {
		return app.New(
			val.GetCtx().Logger,
			tmdb.NewMemDB(),
			nil,
			true,
			simtestutil.EmptyAppOptions{},
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
			baseapp.SetChainID(chainID),
		)
	}
)

// NetworkTestSuite is a test suite for query tests that initializes a network instance.
type NetworkTestSuite struct {
	suite.Suite

	Network        *network.Network
	FeeMarketState feemarkettypes.GenesisState
}

// SetupSuite setups the local network with a genesis state.
func (nts *NetworkTestSuite) SetupSuite() {
	var (
		r   = sample.Rand()
		cfg = network.NewConfig(DefaultAppConstructor, app.ModuleBasics, chainID)
	)

	updateGenesisConfigState := func(moduleName string, moduleState proto.Message) {
		buf, err := cfg.Codec.MarshalJSON(moduleState)
		require.NoError(nts.T(), err)
		cfg.GenesisState[moduleName] = buf
	}

	// initialize fee market
	require.NoError(nts.T(), cfg.Codec.UnmarshalJSON(cfg.GenesisState[feemarkettypes.ModuleName], &nts.FeeMarketState))
	nts.FeeMarketState = populateFeeMarket(r, nts.FeeMarketState)
	updateGenesisConfigState(feemarkettypes.ModuleName, &nts.FeeMarketState)

	nts.Network = network.New(nts.T(), cfg)
}

func populateFeeMarket(_ *rand.Rand, feeMarketState feemarkettypes.GenesisState) feemarkettypes.GenesisState {
	// TODO intercept and populate state randomly if desired
	return feeMarketState
}
