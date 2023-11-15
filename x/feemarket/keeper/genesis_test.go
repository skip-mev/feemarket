package keeper_test

import (
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func (s *KeeperTestSuite) TestInitGenesis() {
	s.Run("default genesis should not panic", func() {
		s.Require().NotPanics(func() {
			s.feemarketKeeper.InitGenesis(s.ctx, *types.DefaultGenesisState())
		})
	})

	s.Run("default AIMD genesis should not panic", func() {
		s.Require().NotPanics(func() {
			s.feemarketKeeper.InitGenesis(s.ctx, *types.DefaultAIMDGenesisState())
		})
	})

	s.Run("bad genesis state should panic", func() {
		gs := types.DefaultGenesisState()
		gs.Params.Window = 0
		s.Require().Panics(func() {
			s.feemarketKeeper.InitGenesis(s.ctx, *gs)
		})
	})

	s.Run("mismatch in params and state for window should panic", func() {
		gs := types.DefaultAIMDGenesisState()
		gs.Params.Window = 1

		s.Require().Panics(func() {
			s.feemarketKeeper.InitGenesis(s.ctx, *gs)
		})
	})

	s.Run("mismatch in params and state for target utilization should panic", func() {
		gs := types.DefaultAIMDGenesisState()
		gs.Params.TargetBlockUtilization = 1

		s.Require().Panics(func() {
			s.feemarketKeeper.InitGenesis(s.ctx, *gs)
		})
	})

	s.Run("mismatch in params and state for max utilization should panic", func() {
		gs := types.DefaultAIMDGenesisState()
		gs.Params.MaxBlockUtilization = 1

		s.Require().Panics(func() {
			s.feemarketKeeper.InitGenesis(s.ctx, *gs)
		})
	})
}

func (s *KeeperTestSuite) TestExportGenesis() {
	s.Run("export genesis should not panic for default eip-1559", func() {
		gs := types.DefaultGenesisState()
		s.feemarketKeeper.InitGenesis(s.ctx, *gs)

		var exportedGenesis *types.GenesisState
		s.Require().NotPanics(func() {
			exportedGenesis = s.feemarketKeeper.ExportGenesis(s.ctx)
		})

		s.Require().Equal(gs, exportedGenesis)
	})

	s.Run("export genesis should not panic for default AIMD eip-1559", func() {
		gs := types.DefaultAIMDGenesisState()
		s.feemarketKeeper.InitGenesis(s.ctx, *gs)

		var exportedGenesis *types.GenesisState
		s.Require().NotPanics(func() {
			exportedGenesis = s.feemarketKeeper.ExportGenesis(s.ctx)
		})

		s.Require().Equal(gs, exportedGenesis)
	})
}
