package keeper_test

import (
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func (s *KeeperTestSuite) TestMsgParams() {
	s.Run("accepts a req with no params", func() {
		req := &types.MsgParams{
			Authority: s.authorityAccount.String(),
		}
		resp, err := s.msgServer.Params(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)
	})

	s.Run("accepts a req with params", func() {
		req := &types.MsgParams{
			Authority: s.authorityAccount.String(),
			Params:    types.DefaultParams(),
		}
		resp, err := s.msgServer.Params(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		params, err := s.feemarketKeeper.GetParams(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(req.Params, params)
	})

	s.Run("rejects a req with invalid signer", func() {
		req := &types.MsgParams{
			Authority: "invalid",
		}
		_, err := s.msgServer.Params(s.ctx, req)
		s.Require().Error(err)
	})
}
