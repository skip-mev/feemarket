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
			name:        "invalid coins negative amount",
			coins:       sdk.Coins{sdk.Coin{Denom: "test", Amount: sdk.NewInt(-1)}},
			wantErr:     true,
			invalidCoin: true,
		},
		{
			name:        "invalid coins invalid denom",
			coins:       sdk.Coins{sdk.Coin{Amount: sdk.NewInt(1)}},
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

func TestSendTip(t *testing.T) {
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
			accs := s.CreateTestAccounts(2)
			if !tc.invalidCoin {
				s.BankKeeper.On("SendCoins", s.Ctx, mock.Anything, mock.Anything, tc.coins).Return(nil).Once()
			}

			if err := post.SendTip(s.BankKeeper, s.Ctx, accs[0].Account.GetAddress(), accs[1].Account.GetAddress(), tc.coins); (err != nil) != tc.wantErr {
				s.Errorf(err, "SendTip() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestPostHandle(t *testing.T) {
	// Same data for every test case
	gasLimit := antesuite.NewTestGasLimit()
	validFeeAmount := types.DefaultMinBaseFee.MulRaw(int64(gasLimit))
	validFeeAmountWithTip := validFeeAmount.Add(sdk.NewInt(100))
	validFee := sdk.NewCoins(sdk.NewCoin("stake", validFeeAmount))
	validFeeWithTip := sdk.NewCoins(sdk.NewCoin("stake", validFeeAmountWithTip))

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
			Name: "signer has enough funds, should pass, no tip",
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
		{
			Name: "signer has enough funds, should pass with tip",
			Malleate: func(suite *antesuite.TestSuite) antesuite.TestCaseArgs {
				accs := suite.CreateTestAccounts(1)
				suite.BankKeeper.On("SendCoinsFromAccountToModule", mock.Anything, accs[0].Account.GetAddress(), types.FeeCollectorName, mock.Anything).Return(nil)
				suite.BankKeeper.On("SendCoins", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

				return antesuite.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].Account.GetAddress())},
					GasLimit:  gasLimit,
					FeeAmount: validFeeWithTip,
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
