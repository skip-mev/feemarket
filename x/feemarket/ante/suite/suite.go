package suite

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	coretesting "cosmossdk.io/core/testing"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/accounts"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	consensuskeeper "cosmossdk.io/x/consensus/keeper"

	testfixture "github.com/skip-mev/feemarket/testutils/fixture"
	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"
	"github.com/skip-mev/feemarket/x/feemarket/ante/mocks"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarketpost "github.com/skip-mev/feemarket/x/feemarket/post"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

type TestSuite struct {
	suite.Suite

	Ctx         sdk.Context
	AnteHandler sdk.AnteHandler
	PostHandler sdk.PostHandler
	ClientCtx   client.Context
	TxBuilder   client.TxBuilder

	AuthKeeper      feemarketante.AccountKeeper
	ConsensusKeeper consensuskeeper.Keeper
	FeeMarketKeeper *feemarketkeeper.Keeper
	BankKeeper      bankkeeper.Keeper
	FeeGrantKeeper  feemarketante.FeeGrantKeeper
	AccountsKeeper  accounts.Keeper

	MockBankKeeper     *mocks.BankKeeper
	MockFeeGrantKeeper *mocks.FeeGrantKeeper
	EncCfg             testutil.TestEncodingConfig

	MsgServer feemarkettypes.MsgServer
}

// TestAccount represents an account used in the tests in x/auth/ante.
type TestAccount struct {
	Account sdk.AccountI
	Priv    cryptotypes.PrivKey
}

type TestAccountBalance struct {
	TestAccount
	sdk.Coins
}

func (s *TestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	s.T().Helper()

	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := s.AuthKeeper.NewAccountWithAddress(s.Ctx, addr)
		err := acc.SetAccountNumber(uint64(i + 1000))
		if err != nil {
			panic(err)
		}
		s.AuthKeeper.SetAccount(s.Ctx, acc)
		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

func (s *TestSuite) SetAccountBalances(accounts []TestAccountBalance) {
	s.T().Helper()

	oldState, err := s.BankKeeper.ExportGenesis(s.Ctx)
	s.Require().NoError(err)

	balances := make([]banktypes.Balance, len(accounts))
	for i, acc := range accounts {
		balances[i] = banktypes.Balance{
			Address: acc.Account.GetAddress().String(),
			Coins:   acc.Coins,
		}
	}

	oldState.Balances = balances
	s.Require().NoError(s.BankKeeper.InitGenesis(s.Ctx, oldState))
}

// SetupTestSuite setups a new test, with new app, context, and anteHandler.
func SetupTestSuite(t *testing.T, mock bool) *TestSuite {
	s := &TestSuite{}

	fixture := testfixture.NewTestFixture(t, nil)
	s.Ctx = fixture.Ctx
	s.EncCfg = fixture.EncodingConfig

	s.AuthKeeper = fixture.AuthKeeper
	s.FeeMarketKeeper = fixture.FeeMarketKeeper
	s.BankKeeper = fixture.BankKeeper
	s.FeeGrantKeeper = fixture.FeeGrantKeeper
	s.ConsensusKeeper = fixture.ConsensusKeeper
	s.AccountsKeeper = fixture.AccountsKeeper

	s.MockBankKeeper = mocks.NewBankKeeper(t)
	s.MockFeeGrantKeeper = mocks.NewFeeGrantKeeper(t)

	s.ClientCtx = client.Context{}.WithTxConfig(s.EncCfg.TxConfig)
	s.TxBuilder = s.ClientCtx.TxConfig.NewTxBuilder()

	s.FeeMarketKeeper.SetEnabledHeight(s.Ctx, -1)
	s.MsgServer = feemarketkeeper.NewMsgServer(s.FeeMarketKeeper)

	s.SetupHandlers(mock)
	s.SetT(t)

	s.Require().NoError(s.BankKeeper.InitGenesis(s.Ctx, &banktypes.GenesisState{}))

	return s
}

func (s *TestSuite) SetupHandlers(mock bool) {
	bankKeeper := s.BankKeeper
	feeGrantKeeper := s.FeeGrantKeeper

	if mock {
		bankKeeper = s.MockBankKeeper
		feeGrantKeeper = s.MockFeeGrantKeeper
	}

	env := runtime.NewEnvironment(nil, coretesting.NewNopLogger())
	// create basic antehandler with the feemarket decorator
	anteDecorators := []sdk.AnteDecorator{
		authante.NewSetUpContextDecorator(env, s.ConsensusKeeper), // outermost AnteDecorator. SetUpContext must be called first
		feemarketante.NewFeeMarketCheckDecorator( // fee market replaces fee deduct decorator
			s.AuthKeeper,
			bankKeeper,
			feeGrantKeeper,
			s.FeeMarketKeeper,
			authante.NewDeductFeeDecorator(
				s.AuthKeeper,
				bankKeeper,
				feeGrantKeeper,
				nil,
			),
		),
	}

	s.AnteHandler = sdk.ChainAnteDecorators(anteDecorators...)

	// create basic postHandler with the feemarket decorator
	postDecorators := []sdk.PostDecorator{
		feemarketpost.NewFeeMarketDeductDecorator(
			s.AuthKeeper,
			bankKeeper,
			s.FeeMarketKeeper,
		),
	}

	s.PostHandler = sdk.ChainPostDecorators(postDecorators...)
}

// TestCase represents a test case used in test tables.
type TestCase struct {
	Name              string
	Malleate          func(*TestSuite) TestCaseArgs
	StateUpdate       func(*TestSuite)
	RunAnte           bool
	RunPost           bool
	Simulate          bool
	ExpPass           bool
	ExpErr            error
	ExpectConsumedGas uint64
	Mock              bool
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
		newCtx  sdk.Context
		anteErr error
		postErr error
	)

	// reset gas meter
	s.Ctx = s.Ctx.WithGasMeter(storetypes.NewGasMeter(NewTestGasLimit()))

	if tc.RunAnte {
		newCtx, anteErr = s.AnteHandler(s.Ctx, tx, tc.Simulate)
	}

	// perform mid-tx state update if configured
	if tc.StateUpdate != nil {
		tc.StateUpdate(s)
	}

	if tc.RunPost && anteErr == nil {
		newCtx, postErr = s.PostHandler(s.Ctx, tx, tc.Simulate, true)
	}

	if tc.ExpPass {
		require.NoError(t, txErr)
		require.NoError(t, anteErr)
		require.NoError(t, postErr)
		require.NotNil(t, newCtx)

		s.Ctx = newCtx
		if tc.RunPost {
			consumedGas := newCtx.GasMeter().GasConsumed()
			require.Equal(t, tc.ExpectConsumedGas, consumedGas)
		}

	} else {
		switch {
		case txErr != nil:
			require.Error(t, txErr)
			require.ErrorIs(t, txErr, tc.ExpErr)

		case anteErr != nil:
			require.Error(t, anteErr)
			require.NoError(t, postErr)
			require.ErrorIs(t, anteErr, tc.ExpErr)

		case postErr != nil:
			require.NoError(t, anteErr)
			require.Error(t, postErr)
			require.ErrorIs(t, postErr, tc.ExpErr)

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
			s.Ctx,
			s.ClientCtx.TxConfig.SignModeHandler().DefaultMode(),
			signerData,
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

// NewTestGasLimit is a test fee gas limit.
func NewTestGasLimit() uint64 {
	return 200000
}
