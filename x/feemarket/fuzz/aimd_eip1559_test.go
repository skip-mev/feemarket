package fuzz_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/skip-mev/feemarket/x/feemarket/types"
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
		window := rapid.Int64Range(1, 50).Draw(t, "window")
		state.Window = make([]uint64, window)

		params := CreateRandomAIMDParams(t)

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
		window := rapid.Int64Range(1, 50).Draw(t, "window")
		state.Window = make([]uint64, window)

		params := CreateRandomAIMDParams(t)

		// Randomly generate the block utilization.
		numBlocks := rapid.Uint64Range(0, uint64(window)*10).Draw(t, "num_blocks")
		gasGen := rapid.Uint64Range(0, params.MaxBlockUtilization)

		// Update the fee market.
		for i := uint64(0); i < numBlocks; i++ {
			blockUtilization := gasGen.Draw(t, "gas")
			prevBaseGasPrice := state.BaseGasPrice

			if err := state.Update(blockUtilization, params); err != nil {
				t.Fatalf("block update errors: %v", err)
			}

			// Update the learning rate.
			state.UpdateLearningRate(params)
			// Update the base gas price.
			state.UpdateBaseGasPrice(params)

			// Ensure that the minimum base fee is always less than the base fee.
			require.True(t, params.MinBaseGasPrice.LTE(state.BaseGasPrice))

			switch {
			case blockUtilization > params.TargetBlockUtilization:
				require.True(t, state.BaseGasPrice.GTE(prevBaseGasPrice))
			case blockUtilization < params.TargetBlockUtilization:
				require.True(t, state.BaseGasPrice.LTE(prevBaseGasPrice))
			default:

				// Account for the delta adjustment.
				net := state.GetNetUtilization(params)
				switch {
				case net.GT(math.ZeroInt()):
					require.True(t, state.BaseGasPrice.GTE(prevBaseGasPrice))
				case net.LT(math.ZeroInt()):
					require.True(t, state.BaseGasPrice.LTE(prevBaseGasPrice))
				default:
					require.True(t, state.BaseGasPrice.Equal(prevBaseGasPrice))
				}
			}

			// Update the current height.
			state.IncrementHeight()
		}
	})
}

// CreateRandomAIMDParams returns a random set of parameters for the AIMD
// EIP-1559 fee market implementation.
func CreateRandomAIMDParams(t *rapid.T) types.Params {
	a := rapid.Uint64Range(1, 1000).Draw(t, "alpha")
	alpha := math.LegacyNewDec(int64(a)).Quo(math.LegacyNewDec(1000))

	b := rapid.Uint64Range(50, 99).Draw(t, "beta")
	beta := math.LegacyNewDec(int64(b)).Quo(math.LegacyNewDec(100))

	th := rapid.Uint64Range(10, 90).Draw(t, "theta")
	theta := math.LegacyNewDec(int64(th)).Quo(math.LegacyNewDec(100))

	d := rapid.Uint64Range(1, 1000).Draw(t, "delta")
	delta := math.LegacyNewDec(int64(d)).Quo(math.LegacyNewDec(1000))

	targetBlockUtilization := rapid.Uint64Range(1, 30_000_000).Draw(t, "target_block_utilization")
	maxBlockUtilization := rapid.Uint64Range(targetBlockUtilization, targetBlockUtilization*5).Draw(t, "max_block_utilization")

	distributeFees := rapid.Bool().Draw(t, "distribute_fees")

	params := types.DefaultAIMDParams()
	params.Alpha = alpha
	params.Beta = beta
	params.Theta = theta
	params.Delta = delta
	params.MaxBlockUtilization = maxBlockUtilization
	params.TargetBlockUtilization = targetBlockUtilization
	params.DistributeFees = distributeFees

	return params
}
