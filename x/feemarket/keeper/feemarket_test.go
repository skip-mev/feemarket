package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

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

func (s *KeeperTestSuite) TestGetMinGasPrices() {
	s.Run("can retrieve min gas prices with default eip-1559", func() {
		gs := types.DefaultGenesisState()
		s.feemarketKeeper.InitGenesis(s.ctx, *gs)

		expected := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, gs.State.BaseFee))

		mgp, err := s.feemarketKeeper.GetMinGasPrices(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(expected, mgp)
	})

	s.Run("can retrieve min gas prices with aimd eip-1559", func() {
		gs := types.DefaultAIMDGenesisState()
		s.feemarketKeeper.InitGenesis(s.ctx, *gs)

		expected := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, gs.State.BaseFee))

		mgp, err := s.feemarketKeeper.GetMinGasPrices(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(expected, mgp)
	})
}
