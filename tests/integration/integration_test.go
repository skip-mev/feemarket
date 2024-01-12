package integration_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/skip-mev/chaintestutil/encoding"
	"github.com/stretchr/testify/suite"

	"github.com/skip-mev/feemarket/tests/app"
	testkeeper "github.com/skip-mev/feemarket/testutils/keeper"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

type IntegrationTestSuite struct {
	suite.Suite
	testkeeper.TestKeepers
	testkeeper.TestMsgServers

	encCfg encoding.TestEncodingConfig
	ctx    sdk.Context
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	s.encCfg = encoding.MakeTestEncodingConfig(app.ModuleBasics.RegisterInterfaces)
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
			MinBaseFee:             math.NewInt(10),
			MinLearningRate:        math.LegacyMustNewDecFromStr("0.1"),
			MaxLearningRate:        math.LegacyMustNewDecFromStr("0.1"),
			TargetBlockUtilization: 5,
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
