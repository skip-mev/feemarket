package suite

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

	appparams "github.com/skip-mev/feemarket/tests/app/params"
	"github.com/skip-mev/feemarket/testutils/encoding"
	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"
	"github.com/skip-mev/feemarket/x/feemarket/ante/mocks"
	"github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarketpost "github.com/skip-mev/feemarket/x/feemarket/post"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

type TestSuite struct {
	suite.Suite

	Ctx         sdk.Context
	AnteHandler sdk.AnteHandler
	PostHandler sdk.PostHandler
	ClientCtx   client.Context
	TxBuilder   client.TxBuilder

	AccountKeeper    authkeeper.AccountKeeper
	FeemarketKeeper  *keeper.Keeper
	BankKeeper       *mocks.BankKeeper
	FeeGrantKeeper   *mocks.FeeGrantKeeper
	EncCfg           appparams.EncodingConfig
	Key              *storetypes.KVStoreKey
	AuthorityAccount sdk.AccAddress
}

// TestAccount represents an account used in the tests in x/auth/ante.
type TestAccount struct {
	Account authtypes.AccountI
	Priv    cryptotypes.PrivKey
}

func (s *TestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := s.AccountKeeper.NewAccountWithAddress(s.Ctx, addr)
		err := acc.SetAccountNumber(uint64(i + 1000))
		if err != nil {
			panic(err)
		}
		s.AccountKeeper.SetAccount(s.Ctx, acc)
		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

// SetupTestSuite setups a new test, with new app, context, and anteHandler.
func SetupTestSuite(t *testing.T) *TestSuite {
	s := &TestSuite{}

	s.EncCfg = encoding.MakeTestEncodingConfig()
	s.Key = storetypes.NewKVStoreKey(types.StoreKey)
	tkey := storetypes.NewTransientStoreKey("transient_test_feemarket")
	testCtx := testutil.DefaultContextWithDB(t, s.Key, tkey)
	s.Ctx = testCtx.Ctx.WithIsCheckTx(false).WithBlockHeight(1)
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

	s.AuthorityAccount = authtypes.NewModuleAddress("gov")
	s.AccountKeeper = authkeeper.NewAccountKeeper(
		s.EncCfg.Codec, authKey, authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix, s.AuthorityAccount.String(),
	)

	s.FeemarketKeeper = keeper.NewKeeper(
		s.EncCfg.Codec,
		s.Key,
		s.AccountKeeper,
		s.AuthorityAccount.String(),
	)

	err = s.FeemarketKeeper.SetParams(s.Ctx, types.DefaultParams())
	require.NoError(t, err)

	err = s.FeemarketKeeper.SetState(s.Ctx, types.DefaultState())
	require.NoError(t, err)

	s.BankKeeper = mocks.NewBankKeeper(t)
	s.FeeGrantKeeper = mocks.NewFeeGrantKeeper(t)

	s.ClientCtx = client.Context{}.WithTxConfig(s.EncCfg.TxConfig)
	s.TxBuilder = s.ClientCtx.TxConfig.NewTxBuilder()

	// create basic antehandler with the feemarket decorator
	anteDecorators := []sdk.AnteDecorator{
		authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		feemarketante.NewFeeMarketCheckDecorator( // fee market replaces fee deduct decorator
			s.FeemarketKeeper,
		),
		authante.NewSigGasConsumeDecorator(s.AccountKeeper, authante.DefaultSigVerificationGasConsumer),
	}

	s.AnteHandler = sdk.ChainAnteDecorators(anteDecorators...)

	// create basic postHandler with the feemarket decorator
	postDecorators := []sdk.PostDecorator{
		feemarketpost.NewFeeMarketDeductDecorator(
			s.AccountKeeper,
			s.BankKeeper,
			s.FeeGrantKeeper,
			s.FeemarketKeeper,
		),
	}

	s.PostHandler = sdk.ChainPostDecorators(postDecorators...)
	return s
}

// TestCase represents a test case used in test tables.
type TestCase struct {
	Name     string
	Malleate func(*TestSuite) TestCaseArgs
	RunAnte  bool
	RunPost  bool
	Simulate bool
	ExpPass  bool
	ExpErr   error
}

type TestCaseArgs struct {
	ChainID   string
	AccNums   []uint64
	AccSeqs   []uint64
	FeeAmount sdk.Coins
	GasLimit  uint64
	Msgs      []sdk.Msg
	Privs     []cryptotypes.PrivKey
}

// DeliverMsgs constructs a tx and runs it through the ante handler. This is used to set the context for a test case, for
// example to test for replay protection.
func (s *TestSuite) DeliverMsgs(t *testing.T, privs []cryptotypes.PrivKey, msgs []sdk.Msg, feeAmount sdk.Coins, gasLimit uint64, accNums, accSeqs []uint64, chainID string, simulate bool) (sdk.Context, error) {
	require.NoError(t, s.TxBuilder.SetMsgs(msgs...))
	s.TxBuilder.SetFeeAmount(feeAmount)
	s.TxBuilder.SetGasLimit(gasLimit)

	tx, txErr := s.CreateTestTx(privs, accNums, accSeqs, chainID)
	require.NoError(t, txErr)
	return s.AnteHandler(s.Ctx, tx, simulate)
}

func (s *TestSuite) RunTestCase(t *testing.T, tc TestCase, args TestCaseArgs) {
	require.NoError(t, s.TxBuilder.SetMsgs(args.Msgs...))
	s.TxBuilder.SetFeeAmount(args.FeeAmount)
	s.TxBuilder.SetGasLimit(args.GasLimit)

	// Theoretically speaking, ante handler unit tests should only test
	// ante handlers, but here we sometimes also test the tx creation
	// process.
	tx, txErr := s.CreateTestTx(args.Privs, args.AccNums, args.AccSeqs, args.ChainID)

	var (
		newCtx    sdk.Context
		handleErr error
	)

	if tc.RunAnte {
		newCtx, handleErr = s.AnteHandler(s.Ctx, tx, tc.Simulate)
	}

	if tc.RunPost {
		newCtx, handleErr = s.PostHandler(s.Ctx, tx, tc.Simulate, true)
	}

	if tc.ExpPass {
		require.NoError(t, txErr)
		require.NoError(t, handleErr)
		require.NotNil(t, newCtx)

		s.Ctx = newCtx
	} else {
		switch {
		case txErr != nil:
			require.Error(t, txErr)
			require.ErrorIs(t, txErr, tc.ExpErr)

		case handleErr != nil:
			require.Error(t, handleErr)
			require.ErrorIs(t, handleErr, tc.ExpErr)

		default:
			t.Fatal("expected one of txErr, handleErr to be an error")
		}
	}
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (s *TestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (authsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  s.ClientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := s.TxBuilder.SetSignatures(sigsV2...)
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
			s.ClientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			s.TxBuilder, priv, s.ClientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = s.TxBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return s.TxBuilder.GetTx(), nil
}

// NewTestFeeAmount is a test fee amount.
func NewTestFeeAmount() sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin("stake", 150))
}

// NewTestGasLimit is a test fee gas limit.
func NewTestGasLimit() uint64 {
	return 200000
}
