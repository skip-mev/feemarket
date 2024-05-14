package ante_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	_ "github.com/cosmos/cosmos-sdk/x/auth"

	antesuite "github.com/skip-mev/feemarket/x/feemarket/ante/suite"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func TestAnteHandle(t *testing.T) {
	// Same data for every test case
	gasLimit := antesuite.NewTestGasLimit()

	validFeeAmount := types.DefaultMinBaseFee.MulInt64(int64(gasLimit))
	validFee := sdk.NewCoins(sdk.NewCoin("stake", validFeeAmount.TruncateInt()))
	validFeeDifferentDenom := sdk.NewCoins(sdk.NewCoin("atom", math.Int(validFeeAmount)))

	testCases := []antesuite.TestCase{
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
			RunPost:  false,
			Simulate: false,
			ExpPass:  false,
			ExpErr:   sdkerrors.ErrInvalidGasLimit,
		},
		// test --gas=auto flag settings
		// when --gas=auto is set, cosmos-sdk sets gas=0 and simulate=true
		{
			Name: "--gas=auto behaviour test",
			Malleate: func(suite *antesuite.TestSuite) antesuite.TestCaseArgs {
				accs := suite.CreateTestAccounts(1)

				return antesuite.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].Account.GetAddress())},
					GasLimit:  0,
					FeeAmount: validFee,
				}
			},
			RunAnte:  true,
			RunPost:  false,
			Simulate: true,
			ExpPass:  true,
		},
		{
			Name: "0 gas given should fail with resolvable denom",
			Malleate: func(suite *antesuite.TestSuite) antesuite.TestCaseArgs {
				accs := suite.CreateTestAccounts(1)

				return antesuite.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].Account.GetAddress())},
					GasLimit:  0,
					FeeAmount: validFeeDifferentDenom,
				}
			},
			RunAnte:  true,
			RunPost:  false,
			Simulate: false,
			ExpPass:  false,
			ExpErr:   sdkerrors.ErrInvalidGasLimit,
		},
		{
			Name: "signer has enough funds, should pass",
			Malleate: func(suite *antesuite.TestSuite) antesuite.TestCaseArgs {
				accs := suite.CreateTestAccounts(1)
				return antesuite.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].Account.GetAddress())},
					GasLimit:  gasLimit,
					FeeAmount: validFee,
				}
			},
			RunAnte:  true,
			RunPost:  false,
			Simulate: false,
			ExpPass:  true,
			ExpErr:   nil,
		},
		{
			Name: "signer has enough funds in resolvable denom, should pass",
			Malleate: func(suite *antesuite.TestSuite) antesuite.TestCaseArgs {
				accs := suite.CreateTestAccounts(1)
				return antesuite.TestCaseArgs{
					Msgs:      []sdk.Msg{testdata.NewTestMsg(accs[0].Account.GetAddress())},
					GasLimit:  gasLimit,
					FeeAmount: validFeeDifferentDenom,
				}
			},
			RunAnte:  true,
			RunPost:  false,
			Simulate: false,
			ExpPass:  true,
			ExpErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.Name), func(t *testing.T) {
			s := antesuite.SetupTestSuite(t, true)
			s.TxBuilder = s.ClientCtx.TxConfig.NewTxBuilder()
			args := tc.Malleate(s)

			s.RunTestCase(t, tc, args)
		})
	}
}
