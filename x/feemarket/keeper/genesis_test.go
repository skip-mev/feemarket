package keeper_test

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/skip-mev/feemarket/testutils"
	"github.com/skip-mev/feemarket/x/feemarket/keeper"
	"github.com/skip-mev/feemarket/x/feemarket/plugins/defaultmarket"
	"github.com/skip-mev/feemarket/x/feemarket/plugins/mocks"
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
			},
		}

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
		encCfg := testutils.CreateTestEncodingConfig()
		key := storetypes.NewKVStoreKey(types.StoreKey)
		testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("new  test"))
		ctx := testCtx.Ctx

		plugin := mocks.NewPanicMarket()
		k := keeper.NewKeeper(encCfg.Codec, key, plugin, s.authorityAccount.String())
		bz, err := plugin.Marshal()
		s.Require().NoError(err)

		gs := types.GenesisState{
			Plugin: bz,
			Params: types.Params{
				Enabled: true,
			},
		}

		s.Require().Panics(func() {
			k.InitGenesis(ctx, gs)
		})
	})
}
