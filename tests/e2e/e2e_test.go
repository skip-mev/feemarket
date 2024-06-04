package e2e_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"

	"github.com/skip-mev/feemarket/x/feemarket"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/suite"

	"github.com/skip-mev/feemarket/tests/e2e"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

var (
	minBaseGasPrice = sdkmath.LegacyMustNewDecFromStr("0.001")
	baseGasPrice    = sdkmath.LegacyMustNewDecFromStr("0.1")

	// config params
	numValidators = 4
	numFullNodes  = 1
	denom         = "stake"

	image = ibc.DockerImage{
		Repository: "feemarket-e2e",
		Version:    "latest",
		UidGid:     "1000:1000",
	}
	oracleImage = ibc.DockerImage{
		Repository: "ghcr.io/skip-mev/slinky-sidecar",
		Version:    "latest",
		UidGid:     "1000:1000",
	}
	encodingConfig = testutil.MakeTestEncodingConfig(
		bank.AppModuleBasic{},
		gov.AppModuleBasic{},
		auth.AppModuleBasic{},
		feemarket.AppModuleBasic{},
	)
	noHostMount   = false
	gasAdjustment = 10.0

	genesisKV = []cosmos.GenesisKV{
		{
			Key: "app_state.feemarket.params",
			Value: feemarkettypes.Params{
				Alpha:               feemarkettypes.DefaultAlpha,
				Beta:                feemarkettypes.DefaultBeta,
				Gamma:               feemarkettypes.DefaultAIMDGamma,
				Delta:               feemarkettypes.DefaultDelta,
				MinBaseGasPrice:     minBaseGasPrice,
				MinLearningRate:     feemarkettypes.DefaultMinLearningRate,
				MaxLearningRate:     feemarkettypes.DefaultMaxLearningRate,
				MaxBlockUtilization: feemarkettypes.DefaultMaxBlockUtilization,
				Window:              feemarkettypes.DefaultWindow,
				FeeDenom:            feemarkettypes.DefaultFeeDenom,
				Enabled:             true,
				DistributeFees:      false,
			},
		},
		{
			Key: "app_state.feemarket.state",
			Value: feemarkettypes.State{
				BaseGasPrice: baseGasPrice,
				LearningRate: feemarkettypes.DefaultMaxLearningRate,
				Window:       make([]uint64, feemarkettypes.DefaultWindow),
				Index:        0,
			},
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
			EncodingConfig: &encodingConfig,
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

func TestE2ETestSuite(t *testing.T) {
	s := e2e.NewIntegrationSuite(
		spec,
		oracleImage,
	)

	suite.Run(t, s)
}
