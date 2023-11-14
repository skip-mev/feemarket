package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/skip-mev/feemarket/x/feemarket/plugins/eip1559/types"
	"github.com/stretchr/testify/require"
)

func TestParams(t *testing.T) {
	testCases := []struct {
		name        string
		p           types.Params
		expectedErr bool
	}{
		{
			name:        "valid base eip-1559 params",
			p:           types.DefaultParams(),
			expectedErr: false,
		},
		{
			name:        "valid aimd eip-1559 params",
			p:           types.DefaultAIMDParams(),
			expectedErr: false,
		},
		{
			name:        "nil alpha",
			p:           types.Params{},
			expectedErr: true,
		},
		{
			name: "negative alpha",
			p: types.Params{
				Alpha: math.LegacyMustNewDecFromStr("-0.1"),
			},
			expectedErr: true,
		},
		{
			name: "beta is nil",
			p: types.Params{
				Alpha: math.LegacyMustNewDecFromStr("0.1"),
			},
			expectedErr: true,
		},
		{
			name: "beta is negative",
			p: types.Params{
				Alpha: math.LegacyMustNewDecFromStr("0.1"),
				Beta:  math.LegacyMustNewDecFromStr("-0.1"),
			},
			expectedErr: true,
		},
		{
			name: "beta is greater than 1",
			p: types.Params{
				Alpha: math.LegacyMustNewDecFromStr("0.1"),
				Beta:  math.LegacyMustNewDecFromStr("1.1"),
			},
			expectedErr: true,
		},
		{
			name: "theta is nil",
			p: types.Params{
				Alpha: math.LegacyMustNewDecFromStr("0.1"),
				Beta:  math.LegacyMustNewDecFromStr("0.1"),
			},
			expectedErr: true,
		},
		{
			name: "theta is negative",
			p: types.Params{
				Alpha: math.LegacyMustNewDecFromStr("0.1"),
				Beta:  math.LegacyMustNewDecFromStr("0.1"),
				Theta: math.LegacyMustNewDecFromStr("-0.1"),
			},
			expectedErr: true,
		},
		{
			name: "theta is greater than 1",
			p: types.Params{
				Alpha: math.LegacyMustNewDecFromStr("0.1"),
				Beta:  math.LegacyMustNewDecFromStr("0.1"),
				Theta: math.LegacyMustNewDecFromStr("1.1"),
			},
			expectedErr: true,
		},
		{
			name: "delta is nil",
			p: types.Params{
				Alpha: math.LegacyMustNewDecFromStr("0.1"),
				Beta:  math.LegacyMustNewDecFromStr("0.1"),
				Theta: math.LegacyMustNewDecFromStr("0.1"),
			},
			expectedErr: true,
		},
		{
			name: "delta is negative",
			p: types.Params{
				Alpha: math.LegacyMustNewDecFromStr("0.1"),
				Beta:  math.LegacyMustNewDecFromStr("0.1"),
				Theta: math.LegacyMustNewDecFromStr("0.1"),
				Delta: math.LegacyMustNewDecFromStr("-0.1"),
			},
			expectedErr: true,
		},
		{
			name: "min base fee is nil",
			p: types.Params{
				Alpha: math.LegacyMustNewDecFromStr("0.1"),
				Beta:  math.LegacyMustNewDecFromStr("0.1"),
				Theta: math.LegacyMustNewDecFromStr("0.1"),
				Delta: math.LegacyMustNewDecFromStr("0.1"),
			},
			expectedErr: true,
		},
		{
			name: "min base fee is negative",
			p: types.Params{
				Alpha:      math.LegacyMustNewDecFromStr("0.1"),
				Beta:       math.LegacyMustNewDecFromStr("0.1"),
				Theta:      math.LegacyMustNewDecFromStr("0.1"),
				Delta:      math.LegacyMustNewDecFromStr("0.1"),
				MinBaseFee: math.NewInt(-1),
			},
			expectedErr: true,
		},
		{
			name: "min learning rate is nil",
			p: types.Params{
				Alpha:      math.LegacyMustNewDecFromStr("0.1"),
				Beta:       math.LegacyMustNewDecFromStr("0.1"),
				Theta:      math.LegacyMustNewDecFromStr("0.1"),
				Delta:      math.LegacyMustNewDecFromStr("0.1"),
				MinBaseFee: math.NewInt(1),
			},
			expectedErr: true,
		},
		{
			name: "min learning rate is negative",
			p: types.Params{
				Alpha:           math.LegacyMustNewDecFromStr("0.1"),
				Beta:            math.LegacyMustNewDecFromStr("0.1"),
				Theta:           math.LegacyMustNewDecFromStr("0.1"),
				Delta:           math.LegacyMustNewDecFromStr("0.1"),
				MinBaseFee:      math.NewInt(1),
				MinLearningRate: math.LegacyMustNewDecFromStr("-0.1"),
			},
			expectedErr: true,
		},
		{
			name: "max learning rate is nil",
			p: types.Params{
				Alpha:           math.LegacyMustNewDecFromStr("0.1"),
				Beta:            math.LegacyMustNewDecFromStr("0.1"),
				Theta:           math.LegacyMustNewDecFromStr("0.1"),
				Delta:           math.LegacyMustNewDecFromStr("0.1"),
				MinBaseFee:      math.NewInt(1),
				MinLearningRate: math.LegacyMustNewDecFromStr("0.1"),
			},
			expectedErr: true,
		},
		{
			name: "max learning rate is negative",
			p: types.Params{
				Alpha:           math.LegacyMustNewDecFromStr("0.1"),
				Beta:            math.LegacyMustNewDecFromStr("0.1"),
				Theta:           math.LegacyMustNewDecFromStr("0.1"),
				Delta:           math.LegacyMustNewDecFromStr("0.1"),
				MinBaseFee:      math.NewInt(1),
				MinLearningRate: math.LegacyMustNewDecFromStr("0.1"),
				MaxLearningRate: math.LegacyMustNewDecFromStr("-0.1"),
			},
			expectedErr: true,
		},
		{
			name: "min learning rate is greater than max learning rate",
			p: types.Params{
				Alpha:           math.LegacyMustNewDecFromStr("0.1"),
				Beta:            math.LegacyMustNewDecFromStr("0.1"),
				Theta:           math.LegacyMustNewDecFromStr("0.1"),
				Delta:           math.LegacyMustNewDecFromStr("0.1"),
				MinBaseFee:      math.NewInt(1),
				MinLearningRate: math.LegacyMustNewDecFromStr("0.1"),
				MaxLearningRate: math.LegacyMustNewDecFromStr("0.05"),
			},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.ValidateBasic()
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
