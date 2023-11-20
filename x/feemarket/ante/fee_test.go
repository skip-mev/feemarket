package ante_test

import (
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/stretchr/testify/require"
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
	ctx         sdk.Context
	anteHandler sdk.AnteHandler
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

	s.bankKeeper = mocks.NewBankKeeper(s.T())
	s.feeGrantKeeper = mocks.NewFeeGrantKeeper(s.T())

	s.msgServer = keeper.NewMsgServer(*s.feemarketKeeper)
	s.queryServer = keeper.NewQueryServer(*s.feemarketKeeper)
}

// TestCase represents a test case used in test tables.
type TestCase struct {
	desc     string
	malleate func(*AnteTestSuite) TestCaseArgs
	simulate bool
	expPass  bool
	expErr   error
}

type TestCaseArgs struct {
	chainID   string
	accNums   []uint64
	accSeqs   []uint64
	feeAmount sdk.Coins
	gasLimit  uint64
	msgs      []sdk.Msg
	privs     []cryptotypes.PrivKey
}

// DeliverMsgs constructs a tx and runs it through the ante handler. This is used to set the context for a test case, for
// example to test for replay protection.
func (s *AnteTestSuite) DeliverMsgs(t *testing.T, privs []cryptotypes.PrivKey, msgs []sdk.Msg, feeAmount sdk.Coins, gasLimit uint64, accNums, accSeqs []uint64, chainID string, simulate bool) (sdk.Context, error) {
	require.NoError(t, s.txBuilder.SetMsgs(msgs...))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	tx, txErr := s.CreateTestTx(privs, accNums, accSeqs, chainID)
	require.NoError(t, txErr)
	return s.anteHandler(s.ctx, tx, simulate)
}

func (s *AnteTestSuite) RunTestCase(t *testing.T, tc TestCase, args TestCaseArgs) {
	require.NoError(t, s.txBuilder.SetMsgs(args.msgs...))
	s.txBuilder.SetFeeAmount(args.feeAmount)
	s.txBuilder.SetGasLimit(args.gasLimit)

	// Theoretically speaking, ante handler unit tests should only test
	// ante handlers, but here we sometimes also test the tx creation
	// process.
	tx, txErr := s.CreateTestTx(args.privs, args.accNums, args.accSeqs, args.chainID)
	newCtx, anteErr := s.anteHandler(s.ctx, tx, tc.simulate)

	if tc.expPass {
		require.NoError(t, txErr)
		require.NoError(t, anteErr)
		require.NotNil(t, newCtx)

		s.ctx = newCtx
	} else {
		switch {
		case txErr != nil:
			require.Error(t, txErr)
			require.ErrorIs(t, txErr, tc.expErr)

		case anteErr != nil:
			require.Error(t, anteErr)
			require.ErrorIs(t, anteErr, tc.expErr)

		default:
			t.Fatal("expected one of txErr, anteErr to be an error")
		}
	}
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (s *AnteTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  s.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := s.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			s.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			s.txBuilder, priv, s.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = s.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return s.txBuilder.GetTx(), nil
}

func (s *AnteTestSuite) TestDeductCoins() {
	accs := s.CreateTestAccounts(1)

	tests := []struct {
		name        string
		acc         TestAccount
		coins       sdk.Coins
		wantErr     bool
		invalidCoin bool
	}{
		{
			name:    "valid",
			acc:     accs[0],
			coins:   sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(10))),
			wantErr: false,
		},
		{
			name:    "valid no coins",
			acc:     accs[0],
			coins:   sdk.NewCoins(),
			wantErr: false,
		},
		{
			name:        "invalid coins",
			acc:         accs[0],
			coins:       sdk.Coins{sdk.Coin{Amount: sdk.NewInt(-1)}},
			wantErr:     true,
			invalidCoin: true,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			if !tc.invalidCoin {
				s.bankKeeper.On("SendCoinsFromAccountToModule", s.ctx, tc.acc.acc.GetAddress(), types.FeeCollectorName, tc.coins).Return(nil).Once()
			}

			if err := ante.DeductCoins(s.bankKeeper, s.ctx, tc.acc.acc, tc.coins); (err != nil) != tc.wantErr {
				s.Errorf(err, "DeductCoins() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func (s *AnteTestSuite) TestAnteHandle() {
	// Same data for every test cases
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	testCases := []TestCase{
		{
			"signer has no funds",
			func(suite *AnteTestSuite) TestCaseArgs {
				accs := suite.CreateTestAccounts(1)
				suite.bankKeeper.On("SendCoinsFromAccountToModule", s.ctx, accs[0].acc.GetAddress(), types.FeeCollectorName, feeAmount).Return(sdkerrors.ErrInsufficientFunds)

				return TestCaseArgs{
					msgs: []sdk.Msg{testdata.NewTestMsg(accs[0].acc.GetAddress())},
				}.WithAccountsInfo(accs)
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"signer has enough funds, should pass",
			func(suite *AnteTestSuite) TestCaseArgs {
				accs := suite.CreateTestAccounts(1)
				suite.bankKeeper.On("SendCoinsFromAccountToModule", s.ctx, accs[0].acc.GetAddress(), types.FeeCollectorName, feeAmount).Return(nil)

				return TestCaseArgs{
					msgs: []sdk.Msg{testdata.NewTestMsg(accs[0].acc.GetAddress())},
				}.WithAccountsInfo(accs)
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			args := tc.malleate(suite)
			args.feeAmount = feeAmount
			args.gasLimit = gasLimit

			suite.RunTestCase(t, tc, args)
		})
	}
}
