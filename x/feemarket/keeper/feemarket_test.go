package keeper_test

import (
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func (s *KeeperTestSuite) TestUpdateFeeMarket() {
	// TODO: add tests.
}

func (s *KeeperTestSuite) TestGetBaseFee() {
	s.Run("can retrieve base fee with default eip-1559", func() {
		gs := types.DefaultGenesisState()
		s.feemarketKeeper.InitGenesis(s.ctx, *gs)

		fee, err := s.feemarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(fee, gs.State.BaseFee)
	})

	s.Run("can retrieve base fee with aimd eip-1559", func() {
		gs := types.DefaultAIMDGenesisState()
		s.feemarketKeeper.InitGenesis(s.ctx, *gs)

		fee, err := s.feemarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(fee, gs.State.BaseFee)
	})
}

func (s *KeeperTestSuite) TestGetLearningRate() {
	s.Run("can retrieve learning rate with default eip-1559", func() {
		gs := types.DefaultGenesisState()
		s.feemarketKeeper.InitGenesis(s.ctx, *gs)

		lr, err := s.feemarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(lr, gs.State.LearningRate)
	})

	s.Run("can retrieve learning rate with aimd eip-1559", func() {
		gs := types.DefaultAIMDGenesisState()
		s.feemarketKeeper.InitGenesis(s.ctx, *gs)

		lr, err := s.feemarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(lr, gs.State.LearningRate)
	})
}
