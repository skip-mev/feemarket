package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func FuzzDefaultFeeMarket(f *testing.F) {
	testCases := []uint64{
		0,
		1_000,
		10_000,
		100_000,
		1_000_000,
		10_000_000,
		100_000_000,
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	defaultLR := math.LegacyMustNewDecFromStr("0.125")
	params := types.DefaultParams()

	// Default fee market.
	f.Fuzz(func(t *testing.T, blockGasUsed uint64) {
		state := types.DefaultState()
		state.MinBaseFee = math.NewIntFromUint64(100)
		state.BaseFee = math.NewIntFromUint64(200)
		err := state.Update(blockGasUsed)

		if blockGasUsed > state.MaxBlockUtilization {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)
		require.Equal(t, blockGasUsed, state.Window[state.Index])

		// Ensure the learning rate is always the default learning rate.
		lr := state.UpdateLearningRate(
			params.Theta,
			params.Alpha,
			params.Beta,
			params.MinLearningRate,
			params.MaxLearningRate,
		)
		require.Equal(t, defaultLR, lr)

		oldFee := state.BaseFee
		newFee := state.UpdateBaseFee(params.Delta)

		if blockGasUsed > state.TargetBlockUtilization {
			require.True(t, newFee.GT(oldFee))
		} else {
			require.True(t, newFee.LT(oldFee))
		}
	})
}

func FuzzAIMDFeeMarket(f *testing.F) {
	testCases := []uint64{
		0,
		1_000,
		10_000,
		100_000,
		1_000_000,
		10_000_000,
		100_000_000,
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	params := types.DefaultAIMDParams()

	// Fee market with adjustable learning rate.
	f.Fuzz(func(t *testing.T, blockGasUsed uint64) {
		state := types.DefaultAIMDState()
		state.MinBaseFee = math.NewIntFromUint64(100)
		state.BaseFee = math.NewIntFromUint64(200)
		state.Window = make([]uint64, 1)
		err := state.Update(blockGasUsed)

		if blockGasUsed > state.MaxBlockUtilization {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)
		require.Equal(t, blockGasUsed, state.Window[state.Index])

		oldFee := state.BaseFee
		newFee := state.UpdateBaseFee(params.Delta)

		if blockGasUsed > state.TargetBlockUtilization {
			require.True(t, newFee.GT(oldFee))
		} else {
			require.True(t, newFee.LT(oldFee))
		}
	})
}
