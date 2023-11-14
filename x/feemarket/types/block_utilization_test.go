package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func TestBlockUtilization(t *testing.T) {
	testCases := []struct {
		name      string
		state     types.BlockUtilization
		expectErr bool
	}{
		{
			name:      "default base EIP-1559 state",
			state:     types.DefaultBlockUtilization(),
			expectErr: false,
		},
		{
			name:      "default AIMD EIP-1559 state",
			state:     types.DefaultAIMDBlockUtilization(),
			expectErr: false,
		},
		{
			name:      "invalid window",
			state:     types.BlockUtilization{},
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
