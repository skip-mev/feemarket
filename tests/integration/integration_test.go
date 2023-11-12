package integration_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	ictestutil "github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/suite"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

var (
	// config params
	numValidators = 1
	numFullNodes  = 0
	denom         = "stake"

	image = ibc.DockerImage{
		Repository: "block-sdk-integration",
		Version:    "latest",
		UidGid:     "1000:1000",
	}
	encodingConfig = MakeEncodingConfig()
	noHostMount    = false
	gasAdjustment  = 2.0

	genesisKV []cosmos.GenesisKV

	consensusParams = ictestutil.Toml{
		"timeout_commit": "3500ms",
	}

	// interchain specification
	spec = &interchaintest.ChainSpec{
		ChainName:     "block-sdk",
		Name:          "block-sdk",
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
			Name:                "block-sdk",
			Denom:               denom,
			ChainID:             "chain-id-0",
			Bin:                 "testappd",
			Bech32Prefix:        "cosmos",
			CoinType:            "118",
			GasAdjustment:       gasAdjustment,
			GasPrices:           fmt.Sprintf("0%s", denom),
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
	suite.Run(t, NewIntegrationTestSuiteFromSpec(spec))
}
