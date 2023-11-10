package keeper_test

import (
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/skip-mev/feemarket/testutils"
	"github.com/skip-mev/feemarket/x/feemarket/keeper"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

type KeeperTestSuite struct {
	suite.Suite

	feemarketKeeper  *keeper.Keeper
	encCfg           testutils.EncodingConfig
	ctx              sdk.Context
	key              *storetypes.KVStoreKey
	authorityAccount sdk.AccAddress

	// Message server variables
	msgServer types.MsgServer

	// Query server variables
	queryServer types.QueryServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.encCfg = testutils.CreateTestEncodingConfig()
	s.key = storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), s.key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	s.authorityAccount = []byte("authority")
	s.feemarketKeeper = keeper.NewKeeper(
		s.encCfg.Codec,
		s.key,
		s.authorityAccount.String(),
	)

	err := s.feemarketKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	s.msgServer = keeper.NewMsgServer(*s.feemarketKeeper)
	s.queryServer = keeper.NewQueryServer(*s.feemarketKeeper)
}

func (s *KeeperTestSuite) TestData() {
	s.Run("get with no data returns error", func() {
		_, err := s.feemarketKeeper.GetData(s.ctx)
		s.Require().Error(err)
	})

	s.Run("set and get valid data", func() {
		data := []byte("testdata")

		s.feemarketKeeper.SetData(s.ctx, data)

		gotData, err := s.feemarketKeeper.GetData(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(data, gotData)
	})
}
