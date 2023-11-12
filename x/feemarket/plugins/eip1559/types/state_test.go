package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/skip-mev/feemarket/x/feemarket/plugins/eip1559/types"
)

func TestState(t *testing.T) {
	testCases := []struct {
		name      string
		state     types.State
		expectErr bool
	}{
		{
			name:      "default base EIP-1559 state",
			state:     types.DefaultState(),
			expectErr: false,
		},
		{
			name:      "default AIMD EIP-1559 state",
			state:     types.DefaultAIMDState(),
			expectErr: false,
		},
		{
			name:      "nil base fee",
			state:     types.State{},
			expectErr: true,
		},
		{
			name: "negative base fee",
			state: types.State{
				BaseFee: math.NewInt(-1),
			},
			expectErr: true,
		},
		{
			name: "nil learning rate",
			state: types.State{
				BaseFee: math.NewInt(1),
			},
			expectErr: true,
		},
		{
			name: "negative learning rate",
			state: types.State{
				BaseFee:      math.NewInt(1),
				LearningRate: math.LegacyMustNewDecFromStr("-1.0"),
			},
			expectErr: true,
		},
		{
			name: "nil block utilization window",
			state: types.State{
				BaseFee:      math.NewInt(1),
				LearningRate: math.LegacyMustNewDecFromStr("0.0"),
			},
			expectErr: true,
		},
		{
			name: "empty block utilization window",
			state: types.State{
				BaseFee:                math.NewInt(1),
				LearningRate:           math.LegacyMustNewDecFromStr("0.0"),
				BlockUtilizationWindow: []uint64{},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.state.ValidateBasic()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
