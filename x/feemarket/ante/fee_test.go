package ante_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	"github.com/skip-mev/feemarket/testutils"
	"github.com/skip-mev/feemarket/x/feemarket/ante"
	"github.com/skip-mev/feemarket/x/feemarket/ante/mocks"
	"github.com/skip-mev/feemarket/x/feemarket/keeper"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

type AnteTestSuite struct {
	suite.Suite
	anteHandler sdk.AnteHandler
	ctx         sdk.Context
	clientCtx   client.Context
	txBuilder   client.TxBuilder

	accountKeeper    authkeeper.AccountKeeper
	bankKeeper       *mocks.BankKeeper
	feeGrantKeeper   *mocks.FeeGrantKeeper
	feemarketKeeper  *keeper.Keeper
	encCfg           testutils.EncodingConfig
	key              *storetypes.KVStoreKey
	authorityAccount sdk.AccAddress

	// Message server variables
	msgServer types.MsgServer

	// Query server variables
	queryServer types.QueryServer
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

// TestAccount represents an account used in the tests in x/auth/ante.
type TestAccount struct {
	acc  authtypes.AccountI
	priv cryptotypes.PrivKey
}

func (s *AnteTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := s.accountKeeper.NewAccountWithAddress(s.ctx, addr)
		err := acc.SetAccountNumber(uint64(i + 1000))
		if err != nil {
			panic(err)
		}
		s.accountKeeper.SetAccount(s.ctx, acc)
		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

func (s *AnteTestSuite) SetupTest() {
	s.encCfg = testutils.CreateTestEncodingConfig()
	s.key = storetypes.NewKVStoreKey(types.StoreKey)
	tkey := storetypes.NewTransientStoreKey("transient_test_feemarket")
	testCtx := testutil.DefaultContextWithDB(s.T(), s.key, tkey)
	s.ctx = testCtx.Ctx.WithIsCheckTx(true).WithBlockHeight(1)
	cms, db := testCtx.CMS, testCtx.DB

	authKey := storetypes.NewKVStoreKey(authtypes.StoreKey)
	tkey = storetypes.NewTransientStoreKey("transient_test_auth")
	cms.MountStoreWithDB(authKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, db)
	err := cms.LoadLatestVersion()
	s.Require().NoError(err)

	maccPerms := map[string][]string{
		"fee_collector":          nil,
		types.ModuleName:         nil,
		types.FeeCollectorName:   {"burner"},
		"mint":                   {"minter"},
		"bonded_tokens_pool":     {"burner", "staking"},
		"not_bonded_tokens_pool": {"burner", "staking"},
		"multiPerm":              {"burner", "minter", "staking"},
		"random":                 {"random"},
	}

	s.authorityAccount = authtypes.NewModuleAddress("gov")
	s.accountKeeper = authkeeper.NewAccountKeeper(
		s.encCfg.Codec, authKey, authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix, s.authorityAccount.String(),
	)

	s.feemarketKeeper = keeper.NewKeeper(
		s.encCfg.Codec,
		s.key,
		s.accountKeeper,
		s.authorityAccount.String(),
	)

	err = s.feemarketKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	s.msgServer = keeper.NewMsgServer(*s.feemarketKeeper)
	s.queryServer = keeper.NewQueryServer(*s.feemarketKeeper)
}

func (s *AnteTestSuite) DeductCoins(t *testing.T) {

}

func (s *AnteTestSuite) TestDeductCoins() {
	accs := s.CreateTestAccounts(1)

	tests := []struct {
		name    string
		acc     TestAccount
		coins   sdk.Coins
		wantErr bool
	}{
		{
			name: "valid no coins",
			acc:  accs[0],
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.bankKeeper.On("SendCoinsFromAccountToModule", s.ctx, tc.acc.acc.GetAddress(), types.FeeCollectorName, tc.coins).Return(nil)

			if err := ante.DeductCoins(s.bankKeeper, s.ctx, tc.acc.acc, tc.coins); (err != nil) != tc.wantErr {
				s.Errorf(err, "DeductCoins() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
