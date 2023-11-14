package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/skip-mev/feemarket/x/feemarket/plugins/eip1559/types"
	"github.com/stretchr/testify/require"
)

func TestGetNetUtilization(t *testing.T) {
	t.Run("target block size is always met with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		err := state.Update(state.TargetBlockUtilization)
		require.NoError(t, err)

		netUtilization := state.GetNetUtilization()
		require.True(t, math.ZeroInt().Equal(netUtilization))

		state.IncrementHeight()

		err = state.Update(state.TargetBlockUtilization)
		require.NoError(t, err)

		netUtilization = state.GetNetUtilization()
		require.True(t, math.ZeroInt().Equal(netUtilization))
	})

	t.Run("target block size is always met with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.BlockUtilizationWindow)-1; i++ {
			err := state.Update(state.TargetBlockUtilization)
			require.NoError(t, err)

			state.IncrementHeight()
		}
		err := state.Update(state.TargetBlockUtilization)
		require.NoError(t, err)

		netUtilization := state.GetNetUtilization()
		require.True(t, math.ZeroInt().Equal(netUtilization))
	})

	t.Run("target block size is always met with max and lows", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.BlockUtilizationWindow)-1; i++ {
			if i%2 == 0 {
				err := state.Update(state.MaxBlockUtilization)
				require.NoError(t, err)
			} else {
				err := state.Update(0)
				require.NoError(t, err)
			}

			state.IncrementHeight()
		}
		if len(state.BlockUtilizationWindow)%2 == 0 {
			err := state.Update(0)
			require.NoError(t, err)
		} else {
			err := state.Update(state.MaxBlockUtilization)
			require.NoError(t, err)
		}

		netUtilization := state.GetNetUtilization()
		require.True(t, math.ZeroInt().Equal(netUtilization))
	})

	t.Run("target block size is always exceeded", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.BlockUtilizationWindow)-1; i++ {
			err := state.Update(state.MaxBlockUtilization)
			require.NoError(t, err)

			state.IncrementHeight()
		}

		err := state.Update(state.MaxBlockUtilization)
		require.NoError(t, err)

		netUtilization := state.GetNetUtilization()
		require.True(t, math.ZeroInt().LT(netUtilization))

		expectedNetUtilization := math.NewIntFromUint64(state.MaxBlockUtilization - state.TargetBlockUtilization).Mul(math.NewIntFromUint64(uint64(len(state.BlockUtilizationWindow))))
		require.True(t, expectedNetUtilization.Equal(netUtilization))
	})
}

func TestGetAverageUtilization(t *testing.T) {
	t.Run("empty block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		averageUtilization := state.GetAverageUtilization()
		require.True(t, math.LegacyZeroDec().Equal(averageUtilization))
	})

	t.Run("full block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		err := state.Update(state.MaxBlockUtilization)
		require.NoError(t, err)

		averageUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("2.0")
		require.True(t, expectedUtilization.Equal(averageUtilization))
	})

	t.Run("target block size with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		err := state.Update(state.TargetBlockUtilization)
		require.NoError(t, err)

		averageUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("1.0")
		require.True(t, expectedUtilization.Equal(averageUtilization))
	})

	t.Run("75 target rate with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		err := state.Update((state.TargetBlockUtilization / 2) + state.TargetBlockUtilization)
		require.NoError(t, err)

		averageUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("1.5")
		require.True(t, expectedUtilization.Equal(averageUtilization))
	})

	t.Run("empty block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		averageUtilization := state.GetAverageUtilization()
		require.True(t, math.LegacyZeroDec().Equal(averageUtilization))
	})

	t.Run("full blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.BlockUtilizationWindow)-1; i++ {
			err := state.Update(state.MaxBlockUtilization)
			require.NoError(t, err)

			state.IncrementHeight()
		}
		err := state.Update(state.MaxBlockUtilization)
		require.NoError(t, err)

		averageUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("2.0")
		require.True(t, expectedUtilization.Equal(averageUtilization))
	})

	t.Run("target block size with aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.BlockUtilizationWindow)-1; i++ {
			err := state.Update(state.TargetBlockUtilization)
			require.NoError(t, err)

			state.IncrementHeight()
		}
		err := state.Update(state.TargetBlockUtilization)
		require.NoError(t, err)

		averageUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("1.0")
		require.True(t, expectedUtilization.Equal(averageUtilization))
	})

	t.Run("75 target rate with aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.BlockUtilizationWindow)-1; i++ {
			err := state.Update((state.TargetBlockUtilization / 2) + state.TargetBlockUtilization)
			require.NoError(t, err)

			state.IncrementHeight()
		}
		err := state.Update((state.TargetBlockUtilization / 2) + state.TargetBlockUtilization)
		require.NoError(t, err)

		averageUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("1.5")
		require.True(t, expectedUtilization.Equal(averageUtilization))
	})

	t.Run("full and empty blocks with aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.BlockUtilizationWindow)-1; i++ {
			if i%2 == 0 {
				err := state.Update(state.MaxBlockUtilization)
				require.NoError(t, err)
			} else {
				err := state.Update(0)
				require.NoError(t, err)
			}

			state.IncrementHeight()
		}
		if len(state.BlockUtilizationWindow)%2 == 0 {
			err := state.Update(0)
			require.NoError(t, err)
		} else {
			err := state.Update(state.MaxBlockUtilization)
			require.NoError(t, err)
		}

		averageUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("1.0")
		require.True(t, expectedUtilization.Equal(averageUtilization))
	})

	t.Run("increasing utilization with aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.BlockUtilizationWindow = make([]uint64, 4)
		state.TargetBlockUtilization = 10
		state.MaxBlockUtilization = 20

		err := state.Update(0)
		require.NoError(t, err)
		state.IncrementHeight()

		err = state.Update(5)
		require.NoError(t, err)
		state.IncrementHeight()

		err = state.Update(10)
		require.NoError(t, err)
		state.IncrementHeight()

		err = state.Update(15)
		require.NoError(t, err)

		averageUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("0.75")
		require.True(t, expectedUtilization.Equal(averageUtilization))
	})
}

func TestUpdate(t *testing.T) {
	t.Run("update base fee", func(t *testing.T) {
		state := types.DefaultState()
		err := state.Update(1)
		require.NoError(t, err)
		require.Equal(t, uint64(1), state.BlockUtilizationWindow[0])

		state.IncrementHeight()

		err = state.Update(2)
		require.NoError(t, err)
		require.Equal(t, uint64(2), state.BlockUtilizationWindow[0])
	})

	t.Run("update base fee in several blocks", func(t *testing.T) {
		state := types.DefaultAIMDState()
		err := state.Update(1)
		require.NoError(t, err)

		state.IncrementHeight()

		err = state.Update(2)
		require.NoError(t, err)

		require.Equal(t, uint64(1), state.BlockUtilizationWindow[0])
		require.Equal(t, uint64(2), state.BlockUtilizationWindow[1])
	})

	t.Run("updates base fee with to the max window size", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.BlockUtilizationWindow = make([]uint64, 2)

		err := state.Update(1)
		require.NoError(t, err)

		state.IncrementHeight()

		err = state.Update(2)
		require.NoError(t, err)

		require.Equal(t, uint64(1), state.BlockUtilizationWindow[0])
		require.Equal(t, uint64(2), state.BlockUtilizationWindow[1])
	})

	t.Run("rejects an update that exceeds max block utilization", func(t *testing.T) {
		state := types.DefaultState()
		state.MaxBlockUtilization = 1

		err := state.Update(1)
		require.NoError(t, err)

		require.Equal(t, uint64(1), state.BlockUtilizationWindow[0])

		state.IncrementHeight()

		err = state.Update(2)
		require.Error(t, err)
	})
}

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
		{
			name: "target block utilization is zero",
			state: types.State{
				BaseFee:                math.NewInt(1),
				LearningRate:           math.LegacyMustNewDecFromStr("0.0"),
				BlockUtilizationWindow: []uint64{1},
			},
			expectErr: true,
		},
		{
			name: "max block utilization is zero",
			state: types.State{
				BaseFee:                math.NewInt(1),
				LearningRate:           math.LegacyMustNewDecFromStr("0.0"),
				BlockUtilizationWindow: []uint64{1},
				TargetBlockUtilization: 1,
			},
			expectErr: true,
		},
		{
			name: "target block utilization is greater than max block utilization",
			state: types.State{
				BaseFee:                math.NewInt(1),
				LearningRate:           math.LegacyMustNewDecFromStr("0.0"),
				BlockUtilizationWindow: []uint64{1},
				TargetBlockUtilization: 2,
				MaxBlockUtilization:    1,
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
