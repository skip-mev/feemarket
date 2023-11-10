package keeper_test

import (
	"github.com/skip-mev/feemarket/x/feemarket/plugins/defaultmarket"
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
		params := types.NewParams(false)
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

func (s *KeeperTestSuite) TestFeeMarketInfo() {
	s.Run("can get default feemarket info", func() {
		req := &types.FeeMarketInfoRequest{}
		resp, err := s.queryServer.FeeMarketInfo(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Equal(defaultmarket.Info, resp.Info)
	})
}
