package e2e_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	ictestutil "github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/suite"

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
				sdk.NewInt(1000),
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
				sdk.NewInt(1000),
				feemarkettypes.DefaultMaxLearningRate,
			),
		},
	}

	consensusParams = ictestutil.Toml{
		"timeout_commit": "3500ms",
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
			Type:                "cosmos",
			Name:                "feemarket",
			Denom:               denom,
			ChainID:             "chain-id-0",
			Bin:                 "feemarketd",
			Bech32Prefix:        "cosmos",
			CoinType:            "118",
			GasAdjustment:       gasAdjustment,
			GasPrices:           fmt.Sprintf("50%s", denom),
			TrustingPeriod:      "48h",
			NoHostMount:         noHostMount,
			ModifyGenesis:       cosmos.ModifyGenesis(genesisKV),
			ConfigFileOverrides: map[string]any{"config/config.toml": ictestutil.Toml{"consensus": consensusParams}},
		},
	}
)

func MakeEncodingConfig() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()
	feemarkettypes.RegisterInterfaces(cfg.InterfaceRegistry)
	return &cfg
}

func TestIntegrationTestSuite(t *testing.T) {
	s := integration.NewIntegrationTestSuiteFromSpec(spec)
	suite.Run(t, s)
}
