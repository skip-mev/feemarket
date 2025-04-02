package post_test

import (
	"fmt"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	antesuite "github.com/skip-mev/feemarket/x/feemarket/ante/suite"
	"github.com/skip-mev/feemarket/x/feemarket/post"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func TestPostHandleDistributeFeesMock(t *testing.T) {
	// Same data for every test case
	const (
		baseDenom              = "stake"
		expectedConsumedGas    = 11700
		expectedConsumedSimGas = expectedConsumedGas + post.BankSendGasConsumption
		gasLimit               = expectedConsumedSimGas
	)

	validFeeAmount := types.DefaultMinBaseGasPrice.MulInt64(int64(gasLimit))
	validFeeAmountWithTip := validFeeAmount.Add(math.LegacyNewDec(100))
	validFeeWithTip := sdk.NewCoins(sdk.NewCoin(baseDenom, validFeeAmountWithTip.TruncateInt()))

	testCases := []antesuite.TestCase{
		{
			Name: "signer has enough funds, should pass with tip",
			Malleate: func(s *antesuite.TestSuite) antesuite.TestCaseArgs {
				accs := s.CreateTestAccounts(1)
				s.MockBankKeeper.On("SendCoinsFromAccountToModule", mock.Anything, accs[0].Account.GetAddress(),
					types.FeeCollectorName, mock.Anything).Return(nil).Once()
				s.MockBankKeeper.On("SendCoinsFromModuleToModule", mock.Anything, types.FeeCollectorName, authtypes.FeeCollectorName, mock.Anything).Return(nil).Once()
				s.MockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, types.FeeCollectorName, mock.Anything, mock.Anything).Return(nil).Once()
				return antesuite.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].Account.GetAddress())},
					GasLimit:  gasLimit,
					FeeAmount: validFeeWithTip,
				}
			},
			RunAnte:           true,
			RunPost:           true,
			Simulate:          false,
			ExpPass:           true,
			MsgRunSuccess:     true,
			ExpErr:            nil,
			ExpectConsumedGas: expectedConsumedGas,
			Mock:              true,
			DistributeFees:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.Name), func(t *testing.T) {
			s := antesuite.SetupTestSuite(t, tc.Mock, true)
			s.TxBuilder = s.ClientCtx.TxConfig.NewTxBuilder()
			args := tc.Malleate(s)

			s.RunTestCase(t, tc, args)
		})
	}
}

func TestPostHandleDistributeFees(t *testing.T) {
	// Same data for every test case
	const (
		baseDenom           = "stake"
		expectedConsumedGas = 65558
		gasLimit            = 100000
	)

	validFeeAmount := types.DefaultMinBaseGasPrice.MulInt64(int64(gasLimit))
	validFeeAmountWithTip := validFeeAmount.Add(math.LegacyNewDec(100))
	validFeeWithTip := sdk.NewCoins(sdk.NewCoin(baseDenom, validFeeAmountWithTip.TruncateInt()))

	testCases := []antesuite.TestCase{
		{
			Name: "signer has enough funds, gaslimit is not enough to complete entire transaction, should pass",
			Malleate: func(s *antesuite.TestSuite) antesuite.TestCaseArgs {
				accs := s.CreateTestAccounts(1)

				balance := antesuite.TestAccountBalance{
					TestAccount: accs[0],
					Coins:       validFeeWithTip,
				}
				s.SetAccountBalances([]antesuite.TestAccountBalance{balance})

				return antesuite.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].Account.GetAddress())},
					GasLimit:  gasLimit,
					FeeAmount: validFeeWithTip,
				}
			},
			RunAnte:           true,
			RunPost:           true,
			Simulate:          false,
			MsgRunSuccess:     false,
			ExpPass:           false,
			ExpErr:            sdkerrors.ErrOutOfGas,
			ExpectConsumedGas: expectedConsumedGas,
			Mock:              false,
			DistributeFees:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.Name), func(t *testing.T) {
			s := antesuite.SetupTestSuite(t, tc.Mock, true)
			s.TxBuilder = s.ClientCtx.TxConfig.NewTxBuilder()
			args := tc.Malleate(s)

			s.RunTestCase(t, tc, args)
		})
	}
}
