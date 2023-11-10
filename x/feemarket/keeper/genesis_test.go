package keeper_test

import (
	"github.com/skip-mev/feemarket/x/feemarket/plugins/defaultmarket"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func (s *KeeperTestSuite) TestInitGenesis() {
	s.Run("default genesis should not panic", func() {
		s.Require().NotPanics(func() {
			s.feemarketKeeper.InitGenesis(s.ctx, *types.NewDefaultGenesisState())
		})
	})

	s.Run("valid genesis should not panic", func() {
		bz, err := defaultmarket.NewDefaultFeeMarket().Marshal()
		s.Require().NoError(err)

		gs := types.GenesisState{
			Plugin: bz,
			Params: types.Params{},
		}

		s.Require().NotPanics(func() {
			s.feemarketKeeper.InitGenesis(s.ctx, gs)
		})
	})

	s.Run("0 bytes plugin bytes should panic", func() {
		gs := types.GenesisState{
			Plugin: make([]byte, 0),
			Params: types.Params{
				Enabled: true,
			}}

		s.Require().Panics(func() {
			s.feemarketKeeper.InitGenesis(s.ctx, gs)
		})
	})

	s.Run("invalid plugin bytes should panic if module enabled", func() {
		gs := types.GenesisState{
			Plugin: []byte("invalid"),
			Params: types.Params{
				Enabled: true,
			},
		}

		s.Require().Panics(func() {
			s.feemarketKeeper.InitGenesis(s.ctx, gs)
		})
	})

	s.Run("plugin init fail should panic", func() {

	})
}
