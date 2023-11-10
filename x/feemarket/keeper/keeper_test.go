package keeper_test

import (
	"github.com/skip-mev/feemarket/x/feemarket/keeper"
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	auctionkeeper    keeper.Keeper
	encCfg           testutils.EncodingConfig
	ctx              sdk.Context
	key              *storetypes.KVStoreKey
	authorityAccount sdk.AccAddress
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.encCfg = testutils.CreateTestEncodingConfig()
	s.key = storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), s.key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	s.accountKeeper = mocks.NewAccountKeeper(s.T())
	s.accountKeeper.On("GetModuleAddress", types.ModuleName).Return(sdk.AccAddress{}).Maybe()

	s.bankKeeper = mocks.NewBankKeeper(s.T())
	s.distrKeeper = mocks.NewDistributionKeeper(s.T())
	s.stakingKeeper = mocks.NewStakingKeeper(s.T())
	s.authorityAccount = sdk.AccAddress([]byte("authority"))
	s.auctionkeeper = keeper.NewKeeper(
		s.encCfg.Codec,
		s.key,
		s.accountKeeper,
		s.bankKeeper,
		s.distrKeeper,
		s.stakingKeeper,
		s.authorityAccount.String(),
	)

	err := s.auctionkeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	s.msgServer = keeper.NewMsgServerImpl(s.auctionkeeper)
}
