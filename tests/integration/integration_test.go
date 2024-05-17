package integration_test

import (
	"testing"

	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/std"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	testkeeper "github.com/skip-mev/feemarket/testutils/keeper"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

type IntegrationTestSuite struct {
	suite.Suite
	testkeeper.TestKeepers
	testkeeper.TestMsgServers

	encCfg TestEncodingConfig
	ctx    sdk.Context
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	s.encCfg = MakeTestEncodingConfig()

	s.ctx, s.TestKeepers, s.TestMsgServers = testkeeper.NewTestSetup(s.T())
}

func (s *IntegrationTestSuite) TestState() {
	s.Run("set and get default eip1559 state", func() {
		state := types.DefaultState()

		err := s.TestKeepers.FeeMarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		gotState, err := s.TestKeepers.FeeMarketKeeper.GetState(s.ctx)
		s.Require().NoError(err)

		s.Require().EqualValues(state, gotState)
	})

	s.Run("set and get aimd eip1559 state", func() {
		state := types.DefaultAIMDState()

		err := s.TestKeepers.FeeMarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		gotState, err := s.TestKeepers.FeeMarketKeeper.GetState(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(state, gotState)
	})
}

func (s *IntegrationTestSuite) TestParams() {
	s.Run("set and get default params", func() {
		params := types.DefaultParams()

		err := s.TestKeepers.FeeMarketKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		gotParams, err := s.TestKeepers.FeeMarketKeeper.GetParams(s.ctx)
		s.Require().NoError(err)

		s.Require().EqualValues(params, gotParams)
	})

	s.Run("set and get custom params", func() {
		params := types.Params{
			Alpha:                  math.LegacyMustNewDecFromStr("0.1"),
			Beta:                   math.LegacyMustNewDecFromStr("0.1"),
			Theta:                  math.LegacyMustNewDecFromStr("0.1"),
			Delta:                  math.LegacyMustNewDecFromStr("0.1"),
			MinBaseFee:             math.LegacyNewDec(10),
			MinLearningRate:        math.LegacyMustNewDecFromStr("0.1"),
			MaxLearningRate:        math.LegacyMustNewDecFromStr("0.1"),
			TargetBlockUtilization: 5,
			MaxBlockUtilization:    10,
			Window:                 1,
			Enabled:                true,
		}

		err := s.TestKeepers.FeeMarketKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		gotParams, err := s.TestKeepers.FeeMarketKeeper.GetParams(s.ctx)
		s.Require().NoError(err)

		s.Require().EqualValues(params, gotParams)
	})
}

// TestEncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type TestEncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeTestEncodingConfig creates a test EncodingConfig for a test configuration.
func MakeTestEncodingConfig() TestEncodingConfig {
	amino := codec.NewLegacyAmino()

	interfaceRegistry := InterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	txCfg := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(amino)
	std.RegisterInterfaces(interfaceRegistry)

	return TestEncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             cdc,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}

func (s *IntegrationTestSuite) TestGetGasPrices() {
	s.Run("set and get default base fee", func() {
		params := types.DefaultParams()
		testDenom := "skip"

		err := s.TestKeepers.FeeMarketKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		state := types.DefaultState()

		err = s.TestKeepers.FeeMarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		prices, err := s.TestKeepers.FeeMarketKeeper.GetMinGasPrices(s.ctx, testDenom)
		s.Require().NoError(err)

		expected := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, state.BaseFee), sdk.NewDecCoinFromDec(testDenom, state.BaseFee))
		s.Require().EqualValues(expected, prices)
	})
}

func InterfaceRegistry() codectypes.InterfaceRegistry {
	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: txsigning.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	if err != nil {
		panic(err)
	}

	// always register
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)

	// call extra registry functions
	types.RegisterInterfaces(interfaceRegistry)

	return interfaceRegistry
}
