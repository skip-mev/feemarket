package e2e_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/suite"

	"github.com/skip-mev/feemarket/tests/e2e"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

var (
	// config params
	numValidators = 3
	numFullNodes  = 1
	denom         = "stake"

	image = ibc.DockerImage{
		Repository: "feemarket-e2e",
		Version:    "latest",
		UidGid:     "1000:1000",
	}
	encodingConfig = MakeEncodingConfig()
	noHostMount    = false
	gasAdjustment  = 10.0

	genesisKV = []cosmos.GenesisKV{
		{
			Key: "app_state.feemarket.params",
			Value: feemarkettypes.NewParams(
				feemarkettypes.DefaultWindow,
				feemarkettypes.DefaultAlpha,
				feemarkettypes.DefaultBeta,
				feemarkettypes.DefaultTheta,
				feemarkettypes.DefaultDelta,
				feemarkettypes.DefaultTargetBlockUtilization,
				feemarkettypes.DefaultMaxBlockUtilization,
				sdkmath.LegacyNewDec(1000),
				feemarkettypes.DefaultMinLearningRate,
				feemarkettypes.DefaultMaxLearningRate,
				feemarkettypes.DefaultFeeDenom,
				true,
			),
		},
		{
			Key: "app_state.feemarket.state",
			Value: feemarkettypes.NewState(
				feemarkettypes.DefaultWindow,
				sdkmath.LegacyNewDec(1000),
				feemarkettypes.DefaultMaxLearningRate,
			),
		},
	}

	// interchain specification
	spec = &interchaintest.ChainSpec{
		ChainName:     "feemarket",
		Name:          "feemarket",
		NumValidators: &numValidators,
		NumFullNodes:  &numFullNodes,
		Version:       "latest",
		NoHostMount:   &noHostMount,
		ChainConfig: ibc.ChainConfig{
			EncodingConfig: encodingConfig,
			Images: []ibc.DockerImage{
				image,
			},
			Type:           "cosmos",
			Name:           "feemarket",
			Denom:          denom,
			ChainID:        "chain-id-0",
			Bin:            "feemarketd",
			Bech32Prefix:   "cosmos",
			CoinType:       "118",
			GasAdjustment:  gasAdjustment,
			GasPrices:      fmt.Sprintf("0%s", denom),
			TrustingPeriod: "48h",
			NoHostMount:    noHostMount,
			ModifyGenesis:  cosmos.ModifyGenesis(genesisKV),
		},
	}
)

func MakeEncodingConfig() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()
	feemarkettypes.RegisterInterfaces(cfg.InterfaceRegistry)
	return &cfg
}

func TestE2ETestSuite(t *testing.T) {
	s := e2e.NewE2ETestSuiteFromSpec(spec)
	suite.Run(t, s)
}
