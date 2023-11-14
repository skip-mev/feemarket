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

	// TODO test further
}
