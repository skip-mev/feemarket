package fuzz_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/skip-mev/feemarket/x/feemarket/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// TestAIMDLearningRate ensure's that the additive increase
// multiplicative decrease learning rate algorithm correctly
// adjusts the learning rate. In particular, if the block
// utilization is greater than theta or less than 1 - theta, then
// the learning rate is increased by the additive increase
// parameter. Otherwise, the learning rate is decreased by
// the multiplicative decrease parameter.
func TestAIMDLearningRate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		// Randomly generate the block utilization.
		numBlocks := rapid.Uint64Range(0, 1000).Draw(t, "num_blocks")
		gasGen := rapid.Uint64Range(0, params.MaxBlockUtilization)

		// Update the fee market.
		for i := uint64(0); i < numBlocks; i++ {
			blockUtilization := gasGen.Draw(t, "gas")
			prevLearningRate := state.LearningRate

			// Update the fee market.
			if err := state.Update(blockUtilization, params); err != nil {
				t.Fatalf("block update errors: %v", err)
			}

			// Update the learning rate.
			lr := state.UpdateLearningRate(params)
			utilization := state.GetAverageUtilization(params)

			// Ensure that the learning rate is always bounded.
			require.True(t, lr.GTE(params.MinLearningRate))
			require.True(t, lr.LTE(params.MaxLearningRate))

			if utilization.LTE(params.Theta) || utilization.GTE(math.LegacyOneDec().Sub(params.Theta)) {
				require.True(t, lr.GTE(prevLearningRate))
			} else {
				require.True(t, lr.LTE(prevLearningRate))
			}

			// Update the current height.
			state.IncrementHeight()
		}
	})
}

// TestAIMDBaseFee ensure's that the additive increase multiplicative
// decrease base fee adjustment algorithm correctly adjusts the base
// fee. In particular, the base fee should function the same as the
// default EIP-1559 base fee adjustment algorithm.
func TestAIMDBaseFee(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		state := types.DefaultAIMDState()
		state.BaseFee = state.BaseFee.Mul(math.NewInt(10))
		params := types.DefaultAIMDParams()

		// Randomly generate the block utilization.
		numBlocks := rapid.Uint64Range(0, 1000).Draw(t, "num_blocks")
		gasGen := rapid.Uint64Range(0, params.MaxBlockUtilization)

		// Update the fee market.
		for i := uint64(0); i < numBlocks; i++ {
			blockUtilization := gasGen.Draw(t, "gas")
			prevBaseFee := state.BaseFee

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

			// Update the current height.
			state.IncrementHeight()
		}
	})
}
