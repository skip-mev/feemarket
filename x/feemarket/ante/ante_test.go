package ante_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/skip-mev/feemarket/testutils"
	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"
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
	feemarketKeeper  *keeper.Keeper
	bankKeeper       *mocks.BankKeeper
	feeGrantKeeper   *mocks.FeeGrantKeeper
	encCfg           testutils.EncodingConfig
	key              *storetypes.KVStoreKey
	authorityAccount sdk.AccAddress
}

// TestAccount represents an account used in the tests in x/feemarket/ante.
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

// SetupTest setups a new test, with new app, context, and anteHandler.
func SetupTestSuite(t *testing.T) *AnteTestSuite {
	s := &AnteTestSuite{}

	s.encCfg = testutils.CreateTestEncodingConfig()
	s.key = storetypes.NewKVStoreKey(types.StoreKey)
	tkey := storetypes.NewTransientStoreKey("transient_test_feemarket")
	testCtx := testutil.DefaultContextWithDB(t, s.key, tkey)
	s.ctx = testCtx.Ctx.WithIsCheckTx(false).WithBlockHeight(1)
	cms, db := testCtx.CMS, testCtx.DB

	authKey := storetypes.NewKVStoreKey(authtypes.StoreKey)
	tkey = storetypes.NewTransientStoreKey("transient_test_auth")
	cms.MountStoreWithDB(authKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, db)
	err := cms.LoadLatestVersion()
	require.NoError(t, err)

	maccPerms := map[string][]string{
		types.ModuleName:       nil,
		types.FeeCollectorName: {"burner"},
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
	require.NoError(t, err)

	err = s.feemarketKeeper.SetState(s.ctx, types.DefaultState())
	require.NoError(t, err)

	s.bankKeeper = mocks.NewBankKeeper(t)
	s.feeGrantKeeper = mocks.NewFeeGrantKeeper(t)

	s.clientCtx = client.Context{}.WithTxConfig(s.encCfg.TxConfig)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// create basic antehandler with the feemarket decorator
	anteDecorators := []sdk.AnteDecorator{
		authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		feemarketante.NewFeeMarketDecorator( // fee market replaces fee deduct decorator
			s.accountKeeper,
			s.bankKeeper,
			s.feeGrantKeeper,
			s.feemarketKeeper,
		),
		authante.NewSigGasConsumeDecorator(s.accountKeeper, authante.DefaultSigVerificationGasConsumer),
	}

	s.anteHandler = sdk.ChainAnteDecorators(anteDecorators...)
	return s
}

// TestCase represents a test case used in test tables.
type TestCase struct {
	name     string
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
func (s *AnteTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (authsigning.Tx, error) {
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
		signerData := authsigning.SignerData{
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

// NewTestFeeAmount is a test fee amount.
func NewTestFeeAmount() sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin("stake", 150))
}

// NewTestGasLimit is a test fee gas limit.
func NewTestGasLimit() uint64 {
	return 200000
}
