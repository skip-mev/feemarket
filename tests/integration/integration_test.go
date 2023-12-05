package integration_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	appparams "github.com/skip-mev/feemarket/tests/app/params"
	"github.com/skip-mev/feemarket/testutils/encoding"
	testkeeper "github.com/skip-mev/feemarket/testutils/keeper"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	testKeepers    testkeeper.TestKeepers
	testMsgServers testkeeper.TestMsgServers
	encCfg         appparams.EncodingConfig
	ctx            sdk.Context
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	s.encCfg = encoding.MakeTestEncodingConfig()

	s.ctx, s.testKeepers, s.testMsgServers = testkeeper.NewTestSetup(s.T())
}

func (s *IntegrationTestSuite) TestState() {
	s.Run("set and get default eip1559 state", func() {
		state := types.DefaultState()

		err := s.testKeepers.FeeMarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		gotState, err := s.testKeepers.FeeMarketKeeper.GetState(s.ctx)
		s.Require().NoError(err)

		s.Require().EqualValues(state, gotState)
	})

	s.Run("set and get aimd eip1559 state", func() {
		state := types.DefaultAIMDState()

		err := s.testKeepers.FeeMarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		gotState, err := s.testKeepers.FeeMarketKeeper.GetState(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(state, gotState)
	})
}

func (s *IntegrationTestSuite) TestParams() {
	s.Run("set and get default params", func() {
		params := types.DefaultParams()

		err := s.testKeepers.FeeMarketKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		gotParams, err := s.testKeepers.FeeMarketKeeper.GetParams(s.ctx)
		s.Require().NoError(err)

		s.Require().EqualValues(params, gotParams)
	})

	s.Run("set and get custom params", func() {
		params := types.Params{
			Alpha:                  math.LegacyMustNewDecFromStr("0.1"),
			Beta:                   math.LegacyMustNewDecFromStr("0.1"),
			Theta:                  math.LegacyMustNewDecFromStr("0.1"),
			Delta:                  math.LegacyMustNewDecFromStr("0.1"),
			MinBaseFee:             math.NewInt(10),
			MinLearningRate:        math.LegacyMustNewDecFromStr("0.1"),
			MaxLearningRate:        math.LegacyMustNewDecFromStr("0.1"),
			TargetBlockUtilization: 5,
			MaxBlockUtilization:    10,
			Window:                 1,
			Enabled:                true,
		}

		err := s.testKeepers.FeeMarketKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		gotParams, err := s.testKeepers.FeeMarketKeeper.GetParams(s.ctx)
		s.Require().NoError(err)

		s.Require().EqualValues(params, gotParams)
	})
}
