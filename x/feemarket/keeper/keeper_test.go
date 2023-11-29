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
	"github.com/skip-mev/feemarket/x/feemarket/types/mocks"
)

type KeeperTestSuite struct {
	suite.Suite

	accountKeeper    *mocks.AccountKeeper
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
	s.accountKeeper = mocks.NewAccountKeeper(s.T())
	s.accountKeeper.On("GetModuleAddress", "feemarket-fee-collector").Return(sdk.AccAddress("feemarket-fee-collector"))

	s.feemarketKeeper = keeper.NewKeeper(
		s.encCfg.Codec,
		s.key,
		s.accountKeeper,
		s.authorityAccount.String(),
	)

	err := s.feemarketKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	err = s.feemarketKeeper.SetState(s.ctx, types.DefaultState())
	s.Require().NoError(err)

	s.msgServer = keeper.NewMsgServer(*s.feemarketKeeper)
	s.queryServer = keeper.NewQueryServer(*s.feemarketKeeper)
}

func (s *KeeperTestSuite) TestState() {
	s.Run("set and get default eip1559 state", func() {
		state := types.DefaultState()

		err := s.feemarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		gotState, err := s.feemarketKeeper.GetState(s.ctx)
		s.Require().NoError(err)

		s.Require().EqualValues(state, gotState)
	})

	s.Run("set and get aimd eip1559 state", func() {
		state := types.DefaultAIMDState()

		err := s.feemarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		gotState, err := s.feemarketKeeper.GetState(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(state, gotState)
	})
}
