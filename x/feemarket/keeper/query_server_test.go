package keeper_test

import (
	"cosmossdk.io/math"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func (s *KeeperTestSuite) TestParamsRequest() {
	s.Run("can get default params", func() {
		req := &types.ParamsRequest{}
		resp, err := s.queryServer.Params(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Equal(types.DefaultParams(), resp.Params)

		params, err := s.feemarketKeeper.GetParams(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(resp.Params, params)
	})

	s.Run("can get updated params", func() {
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
		err := s.feemarketKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		req := &types.ParamsRequest{}
		resp, err := s.queryServer.Params(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Equal(params, resp.Params)

		params, err = s.feemarketKeeper.GetParams(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(resp.Params, params)
	})
}

func (s *KeeperTestSuite) TestStateRequest() {
	s.Run("can get default state", func() {
		req := &types.StateRequest{}
		resp, err := s.queryServer.State(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Equal(types.DefaultState(), resp.State)

		state, err := s.feemarketKeeper.GetState(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(resp.State, state)
	})

	s.Run("can get updated state", func() {
		state := types.State{
			BaseFee:      math.OneInt(),
			LearningRate: math.LegacyOneDec(),
			Window:       []uint64{1},
			Index:        0,
		}
		err := s.feemarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		req := &types.StateRequest{}
		resp, err := s.queryServer.State(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Equal(state, resp.State)

		state, err = s.feemarketKeeper.GetState(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(resp.State, state)
	})
}

func (s *KeeperTestSuite) TestBaseFeeRequest() {
	s.Run("can get default base fee", func() {
		req := &types.BaseFeeRequest{}
		resp, err := s.queryServer.BaseFee(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		fees, err := s.feemarketKeeper.GetMinGasPrices(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(resp.Fees, fees)
	})

	s.Run("can get updated base fee", func() {
		state := types.State{
			BaseFee: math.OneInt(),
		}
		err := s.feemarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		params := types.Params{
			FeeDenom: "test",
		}
		err = s.feemarketKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		req := &types.BaseFeeRequest{}
		resp, err := s.queryServer.BaseFee(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		fees, err := s.feemarketKeeper.GetMinGasPrices(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(resp.Fees, fees)
	})
}
