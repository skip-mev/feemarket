package fuzz_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/skip-mev/feemarket/x/feemarket/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// TestLearningRate ensure's that the learning rate is always
// constant for the default EIP-1559 implementation.
func TestLearningRate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		// Randomly generate alpha and beta.
		a := rapid.Uint64Range(1, 50).Draw(t, "alpha")
		alpha := math.LegacyNewDec(int64(a)).Quo(math.LegacyNewDec(100))

		b := rapid.Uint64Range(50, 99).Draw(t, "beta")
		beta := math.LegacyNewDec(int64(b)).Quo(math.LegacyNewDec(100))

		params.Alpha = alpha
		params.Beta = beta

		prevLearningRate := state.LearningRate

		// Randomly generate the block utilization.
		blockUtilization := rapid.Uint64Range(0, params.MaxBlockUtilization).Draw(t, "gas")

		// Update the fee market.
		if err := state.Update(blockUtilization, params); err != nil {
			t.Fatalf("block update errors: %v", err)
		}

		// Update the learning rate.
		lr := state.UpdateLearningRate(params)
		require.Equal(t, types.DefaultMinLearningRate, lr)
		require.Equal(t, prevLearningRate, state.LearningRate)
	})
}

// TestBaseFee ensure's that the base fee moves in the correct
// direction for the default EIP-1559 implementation.
func TestBaseFee(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		// Update the current base fee to be 10% higher than the minimum base fee.
		prevBaseFee := state.BaseFee.Mul(math.NewInt(11)).Quo(math.NewInt(10))
		state.BaseFee = prevBaseFee

		// Randomly generate the block utilization.
		blockUtilization := rapid.Uint64Range(0, params.MaxBlockUtilization).Draw(t, "gas")

		// Update the fee market.
		if err := state.Update(blockUtilization, params); err != nil {
			t.Fatalf("block update errors: %v", err)
		}

		// Update the learning rate.
		state.UpdateLearningRate(params)
		// Update the base fee.
		state.UpdateBaseFee(params)

		// Ensure that the minimum base fee is always less than the base fee.
		require.True(t, params.MinBaseFee.LTE(state.BaseFee))

		switch {
		case blockUtilization > params.TargetBlockUtilization:
			require.True(t, state.BaseFee.GTE(prevBaseFee))
		case blockUtilization < params.TargetBlockUtilization:
			require.True(t, state.BaseFee.LTE(prevBaseFee))
		default:
			require.Equal(t, state.BaseFee, prevBaseFee)
		}
	})
}
