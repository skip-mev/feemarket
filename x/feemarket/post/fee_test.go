package post_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/mock"

	sdk "github.com/cosmos/cosmos-sdk/types"

	antesuite "github.com/skip-mev/feemarket/x/feemarket/ante/suite"
	"github.com/skip-mev/feemarket/x/feemarket/post"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func TestDeductCoins(t *testing.T) {
	tests := []struct {
		name        string
		coins       sdk.Coins
		wantErr     bool
		invalidCoin bool
	}{
		{
			name:    "valid",
			coins:   sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(10))),
			wantErr: false,
		},
		{
			name:    "valid no coins",
			coins:   sdk.NewCoins(),
			wantErr: false,
		},
		{
			name:        "invalid coins",
			coins:       sdk.Coins{sdk.Coin{Amount: sdk.NewInt(-1)}},
			wantErr:     true,
			invalidCoin: true,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			s := antesuite.SetupTestSuite(t)
			acc := s.CreateTestAccounts(1)[0]
			if !tc.invalidCoin {
				s.BankKeeper.On("SendCoinsFromAccountToModule", s.Ctx, acc.Account.GetAddress(), types.FeeCollectorName, tc.coins).Return(nil).Once()
			}

			if err := post.DeductCoins(s.BankKeeper, s.Ctx, acc.Account, tc.coins); (err != nil) != tc.wantErr {
				s.Errorf(err, "DeductCoins() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestPostHandle(t *testing.T) {
	// Same data for every test case
	gasLimit := antesuite.NewTestGasLimit()
	validFeeAmount := types.DefaultMinBaseFee.MulRaw(int64(gasLimit))
	validFee := sdk.NewCoins(sdk.NewCoin("stake", validFeeAmount))

	testCases := []antesuite.TestCase{
		{
			Name: "signer has no funds",
			Malleate: func(suite *antesuite.TestSuite) antesuite.TestCaseArgs {
				accs := suite.CreateTestAccounts(1)
				suite.BankKeeper.On("SendCoinsFromAccountToModule", mock.Anything, accs[0].Account.GetAddress(), types.FeeCollectorName, mock.Anything).Return(sdkerrors.ErrInsufficientFunds)

				return antesuite.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].Account.GetAddress())},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
			RunAnte:  true,
			RunPost:  true,
			Simulate: false,
			ExpPass:  false,
			ExpErr:   sdkerrors.ErrInsufficientFunds,
		},
		{
			Name: "0 gas given should fail",
			Malleate: func(suite *antesuite.TestSuite) antesuite.TestCaseArgs {
				accs := suite.CreateTestAccounts(1)

				return antesuite.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].Account.GetAddress())},
					GasLimit:  0,
					FeeAmount: validFee,
				}
			},
			RunAnte:  true,
			RunPost:  true,
			Simulate: false,
			ExpPass:  false,
			ExpErr:   sdkerrors.ErrInvalidGasLimit,
		},
		{
			Name: "signer has enough funds, should pass",
			Malleate: func(suite *antesuite.TestSuite) antesuite.TestCaseArgs {
				accs := suite.CreateTestAccounts(1)
				suite.BankKeeper.On("SendCoinsFromAccountToModule", mock.Anything, accs[0].Account.GetAddress(), types.FeeCollectorName, mock.Anything).Return(nil)

				return antesuite.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].Account.GetAddress())},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
			RunAnte:  true,
			RunPost:  true,
			Simulate: false,
			ExpPass:  true,
			ExpErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.Name), func(t *testing.T) {
			s := antesuite.SetupTestSuite(t)
			s.TxBuilder = s.ClientCtx.TxConfig.NewTxBuilder()
			args := tc.Malleate(s)

			s.RunTestCase(t, tc, args)
		})
	}
}
