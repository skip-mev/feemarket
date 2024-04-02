package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func (s *KeeperTestSuite) TestUpdateFeeMarket() {
	s.Run("empty block with default eip1559 with min base fee", func() {
		state := types.DefaultState()
		params := types.DefaultParams()
		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(fee, params.MinBaseFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("empty block with default eip1559 with preset base fee", func() {
		state := types.DefaultState()
		state.BaseFee = state.BaseFee.Mul(math.LegacyNewDec(2))
		params := types.DefaultParams()
		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to decrease by 1/8th.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)

		factor := math.LegacyMustNewDecFromStr("0.875")
		expectedFee := state.BaseFee.Mul(factor)
		s.Require().Equal(fee, expectedFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("empty block default eip1559 with preset base fee that should default to min", func() {
		// Set the base fee to just below the expected threshold decrease of 1/8th. This means it
		// should default to the minimum base fee.
		state := types.DefaultState()
		factor := math.LegacyMustNewDecFromStr("0.125")
		increase := state.BaseFee.Mul(factor)
		state.BaseFee = types.DefaultMinBaseFee.Add(increase).Sub(math.LegacyNewDec(1))

		params := types.DefaultParams()
		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to decrease by 1/8th.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(fee, params.MinBaseFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("target block with default eip1559 at min base fee", func() {
		state := types.DefaultState()
		params := types.DefaultParams()

		// Reaching the target block size means that we expect this to not
		// increase.
		err := state.Update(params.TargetBlockUtilization, params)
		s.Require().NoError(err)

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to remain the same.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(fee, params.MinBaseFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("target block with default eip1559 at preset base fee", func() {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.BaseFee = state.BaseFee.Mul(math.LegacyNewDec(2))
		// Reaching the target block size means that we expect this to not
		// increase.
		err := state.Update(params.TargetBlockUtilization, params)
		s.Require().NoError(err)

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to remain the same.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(state.BaseFee, fee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("max block with default eip1559 at min base fee", func() {
		state := types.DefaultState()
		params := types.DefaultParams()

		// Reaching the target block size means that we expect this to not
		// increase.
		err := state.Update(params.MaxBlockUtilization, params)
		s.Require().NoError(err)

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to increase by 1/8th.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)

		factor := math.LegacyMustNewDecFromStr("1.125")
		expectedFee := state.BaseFee.Mul(factor)
		s.Require().Equal(fee, expectedFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("max block with default eip1559 at preset base fee", func() {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.BaseFee = state.BaseFee.Mul(math.LegacyNewDec(2))
		// Reaching the target block size means that we expect this to not
		// increase.
		err := state.Update(params.MaxBlockUtilization, params)
		s.Require().NoError(err)

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to increase by 1/8th.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)

		factor := math.LegacyMustNewDecFromStr("1.125")
		expectedFee := state.BaseFee.Mul(factor)
		s.Require().Equal(fee, expectedFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("in-between min and target block with default eip1559 at min base fee", func() {
		state := types.DefaultState()
		params := types.DefaultParams()
		params.MaxBlockUtilization = 100
		params.TargetBlockUtilization = 50

		err := state.Update(25, params)
		s.Require().NoError(err)

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to remain the same since it is at min base fee.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(fee, params.MinBaseFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("in-between min and target block with default eip1559 at preset base fee", func() {
		state := types.DefaultState()
		state.BaseFee = state.BaseFee.Mul(math.LegacyNewDec(2))

		params := types.DefaultParams()
		params.MaxBlockUtilization = 100
		params.TargetBlockUtilization = 50
		err := state.Update(25, params)

		s.Require().NoError(err)

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to decrease by 1/8th * 1/2.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)

		factor := math.LegacyMustNewDecFromStr("0.9375")
		expectedFee := state.BaseFee.Mul(factor)
		s.Require().Equal(fee, expectedFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("in-between target and max block with default eip1559 at min base fee", func() {
		state := types.DefaultState()
		params := types.DefaultParams()
		params.MaxBlockUtilization = 100
		params.TargetBlockUtilization = 50

		err := state.Update(75, params)
		s.Require().NoError(err)

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to increase by 1/8th * 1/2.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)

		factor := math.LegacyMustNewDecFromStr("1.0625")
		expectedFee := state.BaseFee.Mul(factor)
		s.Require().Equal(fee, expectedFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("in-between target and max block with default eip1559 at preset base fee", func() {
		state := types.DefaultState()
		state.BaseFee = state.BaseFee.Mul(math.LegacyNewDec(2))
		params := types.DefaultParams()
		params.MaxBlockUtilization = 100
		params.TargetBlockUtilization = 50

		err := state.Update(75, params)
		s.Require().NoError(err)

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to increase by 1/8th * 1/2.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)

		factor := math.LegacyMustNewDecFromStr("1.0625")
		expectedFee := state.BaseFee.Mul(factor)
		s.Require().Equal(fee, expectedFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(math.LegacyMustNewDecFromStr("0.125"), lr)
	})

	s.Run("empty blocks with aimd eip1559 with min base fee", func() {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()
		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(fee, params.MinBaseFee)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		expectedLR := state.LearningRate.Add(params.Alpha)
		s.Require().Equal(expectedLR, lr)
	})

	s.Run("empty blocks with aimd eip1559 with preset base fee", func() {
		state := types.DefaultAIMDState()
		state.BaseFee = state.BaseFee.Mul(math.LegacyNewDec(2))
		params := types.DefaultAIMDParams()
		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to decrease by 1/8th and the learning rate to
		// increase by alpha.
		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		expectedLR := state.LearningRate.Add(params.Alpha)
		s.Require().Equal(expectedLR, lr)

		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		factor := math.LegacyOneDec().Add(math.LegacyMustNewDecFromStr("-1.0").Mul(lr))
		expectedFee := state.BaseFee.Mul(factor)
		s.Require().Equal(fee, expectedFee)
	})

	s.Run("empty blocks aimd eip1559 with preset base fee that should default to min", func() {
		params := types.DefaultAIMDParams()

		state := types.DefaultAIMDState()
		lr := math.LegacyMustNewDecFromStr("0.125")
		increase := state.BaseFee.Mul(lr).TruncateInt()

		state.BaseFee = types.DefaultMinBaseFee.Add(math.LegacyNewDecFromInt(increase)).Sub(math.LegacyNewDec(1))
		state.LearningRate = lr

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		expectedLR := state.LearningRate.Add(params.Alpha)
		s.Require().Equal(expectedLR, lr)

		// We expect the base fee to decrease by 1/8th and the learning rate to
		// increase by alpha.
		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(fee, params.MinBaseFee)
	})

	s.Run("target block with aimd eip1559 at min base fee + LR", func() {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		// Reaching the target block size means that we expect this to not
		// increase.
		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.TargetBlockUtilization
		}

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to remain the same and the learning rate to
		// remain at minimum.
		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(params.MinLearningRate, lr)

		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(state.BaseFee, fee)
	})

	s.Run("target block with aimd eip1559 at preset base fee + LR", func() {
		state := types.DefaultAIMDState()
		state.BaseFee = state.BaseFee.Mul(math.LegacyNewDec(2))
		state.LearningRate = math.LegacyMustNewDecFromStr("0.125")
		params := types.DefaultAIMDParams()

		// Reaching the target block size means that we expect this to not
		// increase.
		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.TargetBlockUtilization
		}

		s.setGenesisState(params, state)

		s.Require().NoError(s.feeMarketKeeper.UpdateFeeMarket(s.ctx))

		// We expect the base fee to decrease by 1/8th and the learning rate to
		// decrease by lr * beta.
		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		expectedLR := state.LearningRate.Mul(params.Beta)
		s.Require().Equal(expectedLR, lr)

		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(state.BaseFee, fee)
	})
}

func (s *KeeperTestSuite) TestGetBaseFee() {
	s.Run("can retrieve base fee with default eip-1559", func() {
		gs := types.DefaultGenesisState()
		s.feeMarketKeeper.InitGenesis(s.ctx, *gs)

		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(fee, gs.State.BaseFee)
	})

	s.Run("can retrieve base fee with aimd eip-1559", func() {
		gs := types.DefaultAIMDGenesisState()
		s.feeMarketKeeper.InitGenesis(s.ctx, *gs)

		fee, err := s.feeMarketKeeper.GetBaseFee(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(fee, gs.State.BaseFee)
	})
}

func (s *KeeperTestSuite) TestGetLearningRate() {
	s.Run("can retrieve learning rate with default eip-1559", func() {
		gs := types.DefaultGenesisState()
		s.feeMarketKeeper.InitGenesis(s.ctx, *gs)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(lr, gs.State.LearningRate)
	})

	s.Run("can retrieve learning rate with aimd eip-1559", func() {
		gs := types.DefaultAIMDGenesisState()
		s.feeMarketKeeper.InitGenesis(s.ctx, *gs)

		lr, err := s.feeMarketKeeper.GetLearningRate(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(lr, gs.State.LearningRate)
	})
}

func (s *KeeperTestSuite) TestGetMinGasPrices() {
	s.Run("can retrieve min gas prices with default eip-1559", func() {
		gs := types.DefaultGenesisState()
		s.feeMarketKeeper.InitGenesis(s.ctx, *gs)

		expected := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, gs.State.BaseFee))

		mgp, err := s.feeMarketKeeper.GetMinGasPrices(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(expected, mgp)
	})

	s.Run("can retrieve min gas prices with aimd eip-1559", func() {
		gs := types.DefaultAIMDGenesisState()
		s.feeMarketKeeper.InitGenesis(s.ctx, *gs)

		expected := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, gs.State.BaseFee))

		mgp, err := s.feeMarketKeeper.GetMinGasPrices(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(expected, mgp)
	})
}

func (s *KeeperTestSuite) setGenesisState(params types.Params, state types.State) {
	gs := types.NewGenesisState(params, state)
	s.NotPanics(func() {
		s.feeMarketKeeper.InitGenesis(s.ctx, *gs)
	})
}
