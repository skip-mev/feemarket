package ante_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/mock"

	"github.com/skip-mev/feemarket/x/feemarket/ante"
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
			s := SetupTestSuite(t, false)
			acc := s.CreateTestAccounts(1)[0]
			if !tc.invalidCoin {
				s.bankKeeper.On("SendCoinsFromAccountToModule", s.ctx, acc.acc.GetAddress(), types.FeeCollectorName, tc.coins).Return(nil).Once()
			}

			if err := ante.DeductCoins(s.bankKeeper, s.ctx, acc.acc, tc.coins); (err != nil) != tc.wantErr {
				s.Errorf(err, "DeductCoins() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestAnteHandle(t *testing.T) {
	// Same data for every test case
	gasLimit := NewTestGasLimit()
	validFeeAmount := types.DefaultMinBaseFee.MulRaw(int64(gasLimit))
	validFee := sdk.NewCoins(sdk.NewCoin("stake", validFeeAmount))

	testCases := []TestCase{
		{
			"signer has no funds",
			func(suite *AnteTestSuite) TestCaseArgs {
				accs := suite.CreateTestAccounts(1)
				suite.bankKeeper.On("SendCoinsFromAccountToModule", mock.Anything, accs[0].acc.GetAddress(), types.FeeCollectorName, mock.Anything).Return(sdkerrors.ErrInsufficientFunds).Once()

				return TestCaseArgs{
					msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].acc.GetAddress())},
					gasLimit:  gasLimit,
					feeAmount: validFee,
				}
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"0 gas given should fail",
			func(suite *AnteTestSuite) TestCaseArgs {
				accs := suite.CreateTestAccounts(1)

				return TestCaseArgs{
					msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].acc.GetAddress())},
					gasLimit:  0,
					feeAmount: validFee,
				}
			},
			false,
			false,
			sdkerrors.ErrInvalidGasLimit,
		},
		{
			"signer has enough funds, should pass",
			func(suite *AnteTestSuite) TestCaseArgs {
				accs := suite.CreateTestAccounts(1)
				suite.bankKeeper.On("SendCoinsFromAccountToModule", mock.Anything, accs[0].acc.GetAddress(), types.FeeCollectorName, mock.Anything).Return(nil).Once()

				return TestCaseArgs{
					msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].acc.GetAddress())},
					gasLimit:  gasLimit,
					feeAmount: validFee,
				}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			s := SetupTestSuite(t, false)
			s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
			args := tc.malleate(s)

			s.RunTestCase(t, tc, args)
		})
	}
}
