// Package networksuite provides a base test suite for tests that need a local network instance
package networksuite

import (
	"math/rand"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/skip-mev/feemarket/testutils/network"
	"github.com/skip-mev/feemarket/testutils/sample"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
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
		cfg = network.DefaultConfig()
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

func populateFeeMarket(r *rand.Rand, feeMarketState feemarkettypes.GenesisState) feemarkettypes.GenesisState {
	return feeMarketState
}
