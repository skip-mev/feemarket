package types_test

import (
	"math/rand"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

var OneHundred = math.LegacyNewDecFromInt(math.NewInt(100))

func TestState_Update(t *testing.T) {
	t.Run("can add to window", func(t *testing.T) {
		state := types.DefaultState()

		err := state.Update(100)
		require.NoError(t, err)
		require.Equal(t, uint64(100), state.Window[0])
	})

	t.Run("can add several txs to window", func(t *testing.T) {
		state := types.DefaultState()

		err := state.Update(100)
		require.NoError(t, err)
		require.Equal(t, uint64(100), state.Window[0])

		err = state.Update(200)
		require.NoError(t, err)
		require.Equal(t, uint64(300), state.Window[0])
	})

	t.Run("errors when it exceeds max block utilization", func(t *testing.T) {
		state := types.DefaultState()

		err := state.Update(state.MaxBlockUtilization + 1)
		require.Error(t, err)
	})

	t.Run("can update with several blocks in default eip-1559", func(t *testing.T) {
		state := types.DefaultState()

		err := state.Update(100)
		require.NoError(t, err)
		require.Equal(t, uint64(100), state.Window[0])

		state.IncrementHeight()

		err = state.Update(200)
		require.NoError(t, err)
		require.Equal(t, uint64(200), state.Window[0])

		err = state.Update(300)
		require.NoError(t, err)
		require.Equal(t, uint64(500), state.Window[0])
	})

	t.Run("can update with several blocks in default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		err := state.Update(100)
		require.NoError(t, err)
		require.Equal(t, uint64(100), state.Window[0])

		state.IncrementHeight()

		err = state.Update(200)
		require.NoError(t, err)
		require.Equal(t, uint64(200), state.Window[1])

		state.IncrementHeight()

		err = state.Update(300)
		require.NoError(t, err)
		require.Equal(t, uint64(300), state.Window[2])

		state.IncrementHeight()

		err = state.Update(400)
		require.NoError(t, err)
		require.Equal(t, uint64(400), state.Window[3])
	})

	t.Run("correctly wraps around with aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 3)

		err := state.Update(100)
		require.NoError(t, err)
		require.Equal(t, uint64(100), state.Window[0])

		state.IncrementHeight()

		err = state.Update(200)
		require.NoError(t, err)
		require.Equal(t, uint64(200), state.Window[1])

		state.IncrementHeight()

		err = state.Update(300)
		require.NoError(t, err)
		require.Equal(t, uint64(300), state.Window[2])

		state.IncrementHeight()
		require.Equal(t, uint64(0), state.Window[0])

		err = state.Update(400)
		require.NoError(t, err)
		require.Equal(t, uint64(400), state.Window[0])
		require.Equal(t, uint64(200), state.Window[1])
		require.Equal(t, uint64(300), state.Window[2])
	})
}

func TestState_UpdateBaseFee(t *testing.T) {
	t.Run("empty block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		state.BaseFee = math.NewInt(1000)
		state.MinBaseFee = math.NewInt(125)

		params := types.DefaultParams()

		newBaseFee := state.UpdateBaseFee(params.Delta)
		expectedBaseFee := math.NewInt(875)
		require.True(t, expectedBaseFee.Equal(newBaseFee))
	})

	t.Run("target block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		state.BaseFee = math.NewInt(1000)
		state.MinBaseFee = math.NewInt(125)

		params := types.DefaultParams()

		state.Window[0] = params.TargetBlockUtilization

		newBaseFee := state.UpdateBaseFee(params.Delta)
		expectedBaseFee := math.NewInt(1000)
		require.True(t, expectedBaseFee.Equal(newBaseFee))
	})

	t.Run("full block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		state.BaseFee = math.NewInt(1000)
		state.MinBaseFee = math.NewInt(125)

		params := types.DefaultParams()

		state.Window[0] = params.MaxBlockUtilization

		newBaseFee := state.UpdateBaseFee(params.Delta)
		expectedBaseFee := math.NewInt(1125)
		require.True(t, expectedBaseFee.Equal(newBaseFee))
	})

	t.Run("empty block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.BaseFee = math.NewInt(1000)
		state.MinBaseFee = math.NewInt(125)
		state.LearningRate = math.LegacyMustNewDecFromStr("0.125")

		params := types.DefaultAIMDParams()

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		newBaseFee := state.UpdateBaseFee(params.Delta)

		expectedBaseFee := math.NewInt(850)
		require.True(t, expectedBaseFee.Equal(newBaseFee))
	})

	t.Run("target block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.BaseFee = math.NewInt(1000)
		state.MinBaseFee = math.NewInt(125)
		state.LearningRate = math.LegacyMustNewDecFromStr("0.125")

		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.TargetBlockUtilization
		}

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		newBaseFee := state.UpdateBaseFee(params.Delta)

		expectedBaseFee := math.NewInt(1000)
		require.True(t, expectedBaseFee.Equal(newBaseFee))
	})

	t.Run("full blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.BaseFee = math.NewInt(1000)
		state.MinBaseFee = math.NewInt(125)
		state.LearningRate = math.LegacyMustNewDecFromStr("0.125")

		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.MaxBlockUtilization
		}

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		newBaseFee := state.UpdateBaseFee(params.Delta)

		expectedBaseFee := math.NewInt(1150)
		require.True(t, expectedBaseFee.Equal(newBaseFee))
	})

	t.Run("never goes below min base fee with default eip1599", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		// Empty block
		newBaseFee := state.UpdateBaseFee(params.Delta)
		expectedBaseFee := params.MinBaseFee
		require.True(t, expectedBaseFee.Equal(newBaseFee))
	})

	t.Run("never goes below min base fee with default aimd eip1599", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		// Empty blocks
		newBaseFee := state.UpdateBaseFee(params.Delta)
		expectedBaseFee := params.MinBaseFee
		require.True(t, expectedBaseFee.Equal(newBaseFee))
	})
}

func TestState_UpdateLearningRate(t *testing.T) {
	t.Run("empty block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("target block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = state.TargetBlockUtilization

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("full block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = state.MaxBlockUtilization

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("between 0 and target with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = 50000

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("between target and max with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = state.TargetBlockUtilization + 50000

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("random value with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		randomValue := rand.Int63n(1000000000)
		state.Window[0] = uint64(randomValue)

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("empty block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := params.MinLearningRate.Add(params.Alpha)
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("target block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		defaultLR := math.LegacyMustNewDecFromStr("0.125")
		state.LearningRate = defaultLR

		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = state.TargetBlockUtilization
		}

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := defaultLR.Mul(params.Beta)
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("full blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		defaultLR := math.LegacyMustNewDecFromStr("0.125")
		state.LearningRate = defaultLR

		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = state.MaxBlockUtilization
		}

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := defaultLR.Add(params.Alpha)
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("varying blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		defaultLR := math.LegacyMustNewDecFromStr("0.125")
		state.LearningRate = defaultLR

		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = state.MaxBlockUtilization
			} else {
				state.Window[i] = 0
			}
		}

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := defaultLR.Mul(params.Beta)
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("exceeds threshold with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		defaultLR := math.LegacyMustNewDecFromStr("0.125")
		state.LearningRate = defaultLR

		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = state.MaxBlockUtilization
			} else {
				state.Window[i] = state.TargetBlockUtilization + 1
			}
		}

		state.UpdateLearningRate(params.Theta, params.Alpha, params.Beta, params.MinLearningRate, params.MaxLearningRate)
		expectedLearningRate := defaultLR.Add(params.Alpha)
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})
}

func TestState_GetNetUtilization(t *testing.T) {
	t.Run("empty block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()

		netUtilization := state.GetNetUtilization()
		expectedUtilization := math.NewInt(0).Sub(math.NewIntFromUint64(state.TargetBlockUtilization))
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("target block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()

		state.Window[0] = state.TargetBlockUtilization

		netUtilization := state.GetNetUtilization()
		expectedUtilization := math.NewInt(0)
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("full block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()

		state.Window[0] = state.MaxBlockUtilization

		netUtilization := state.GetNetUtilization()
		expectedUtilization := math.NewIntFromUint64(state.MaxBlockUtilization - state.TargetBlockUtilization)
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("empty block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		netUtilization := state.GetNetUtilization()

		multiple := math.NewIntFromUint64(uint64(len(state.Window)))
		expectedUtilization := math.NewInt(0).Sub(math.NewIntFromUint64(state.TargetBlockUtilization)).Mul(multiple)
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("full blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = state.MaxBlockUtilization
		}

		netUtilization := state.GetNetUtilization()

		multiple := math.NewIntFromUint64(uint64(len(state.Window)))
		expectedUtilization := math.NewIntFromUint64(state.MaxBlockUtilization).Sub(math.NewIntFromUint64(state.TargetBlockUtilization)).Mul(multiple)
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("varying blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = state.MaxBlockUtilization
			} else {
				state.Window[i] = 0
			}
		}

		netUtilization := state.GetNetUtilization()
		expectedUtilization := math.ZeroInt()
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("exceeds target rate with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = params.MaxBlockUtilization
			} else {
				state.Window[i] = params.TargetBlockUtilization
			}
		}

		netUtilization := state.GetNetUtilization()
		first := math.NewIntFromUint64(params.MaxBlockUtilization).Mul(math.NewIntFromUint64(params.Window / 2))
		second := math.NewIntFromUint64(params.TargetBlockUtilization).Mul(math.NewIntFromUint64(params.Window / 2))
		expectedUtilization := first.Add(second).Sub(math.NewIntFromUint64(params.TargetBlockUtilization).Mul(math.NewIntFromUint64(params.Window)))
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("state with 4 entries in window with different updates", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 4)
		state.TargetBlockUtilization = 100
		state.MaxBlockUtilization = 200

		state.Window[0] = 100
		state.Window[1] = 200
		state.Window[2] = 0
		state.Window[3] = 50

		netUtilization := state.GetNetUtilization()
		expectedUtilization := math.NewIntFromUint64(50).Mul(math.NewInt(-1))
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("state with 4 entries in window with monotonically increasing updates", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 4)
		state.TargetBlockUtilization = 100
		state.MaxBlockUtilization = 200

		state.Window[0] = 0
		state.Window[1] = 25
		state.Window[2] = 50
		state.Window[3] = 75

		netUtilization := state.GetNetUtilization()
		expectedUtilization := math.NewIntFromUint64(250).Mul(math.NewInt(-1))
		require.True(t, expectedUtilization.Equal(netUtilization))
	})
}

func TestState_GetAverageUtilization(t *testing.T) {
	t.Run("empty block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()

		avgUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyZeroDec()
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("target block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()

		state.Window[0] = state.TargetBlockUtilization

		avgUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("0.5")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("full block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()

		state.Window[0] = state.MaxBlockUtilization

		avgUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("1.0")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("empty block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		avgUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyZeroDec()
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("target block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = state.TargetBlockUtilization
		}

		avgUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("0.5")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("full blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = state.MaxBlockUtilization
		}

		avgUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("1.0")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("varying blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = state.MaxBlockUtilization
			} else {
				state.Window[i] = 0
			}
		}

		avgUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("0.5")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("exceeds target rate with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = state.MaxBlockUtilization
			} else {
				state.Window[i] = state.TargetBlockUtilization
			}
		}

		avgUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("0.75")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("state with 4 entries in window with different updates", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 4)
		state.TargetBlockUtilization = 100
		state.MaxBlockUtilization = 200

		state.Window[0] = 100
		state.Window[1] = 200
		state.Window[2] = 0
		state.Window[3] = 50

		avgUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("0.4375")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("state with 4 entries in window with monotonically increasing updates", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 4)

		params := types.DefaultAIMDParams()
		params.Window = 4
		state.TargetBlockUtilization = 100
		state.MaxBlockUtilization = 200

		state.Window[0] = 0
		state.Window[1] = 25
		state.Window[2] = 50
		state.Window[3] = 75

		avgUtilization := state.GetAverageUtilization()
		expectedUtilization := math.LegacyMustNewDecFromStr("0.1875")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})
}

func TestState_ValidateBasic(t *testing.T) {
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
			name:      "invalid window",
			state:     types.State{},
			expectErr: true,
		},
		{
			name: "invalid negative base fee",
			state: types.State{
				Window:  make([]uint64, 1),
				BaseFee: math.NewInt(-1),
			},
			expectErr: true,
		},
		{
			name: "invalid learning rate",
			state: types.State{
				Window:       make([]uint64, 1),
				BaseFee:      math.NewInt(1),
				LearningRate: math.LegacyMustNewDecFromStr("-1.0"),
			},
			expectErr: true,
		},
		{
			name: "valid other state",
			state: types.State{
				Window:                 make([]uint64, 1),
				BaseFee:                math.NewInt(1),
				LearningRate:           math.LegacyMustNewDecFromStr("0.5"),
				MaxBlockUtilization:    10,
				TargetBlockUtilization: 5,
			},
			expectErr: false,
		},
		{
			name: "invalid zero base fee",
			state: types.State{
				Window:       make([]uint64, 1),
				BaseFee:      math.ZeroInt(),
				LearningRate: math.LegacyMustNewDecFromStr("0.5"),
			},
			expectErr: true,
		},
		{
			name: "invalid zero learning rate",
			state: types.State{
				Window:       make([]uint64, 1),
				BaseFee:      math.NewInt(1),
				LearningRate: math.LegacyZeroDec(),
			},
			expectErr: true,
		},
		{
			name: "invalid target size",
			state: types.State{
				Window:                 make([]uint64, 1),
				BaseFee:                math.NewInt(1),
				LearningRate:           math.LegacyMustNewDecFromStr("0.5"),
				TargetBlockUtilization: 0,
				MaxBlockUtilization:    100,
			},
			expectErr: true,
		},
		{
			name: "invalid max size",
			state: types.State{
				Window:                 make([]uint64, 1),
				BaseFee:                math.NewInt(1),
				LearningRate:           math.LegacyMustNewDecFromStr("0.5"),
				TargetBlockUtilization: 10,
				MaxBlockUtilization:    0,
			},
			expectErr: true,
		},
		{
			name: "target is larger than max",
			state: types.State{
				Window:                 make([]uint64, 1),
				BaseFee:                math.NewInt(1),
				LearningRate:           math.LegacyMustNewDecFromStr("0.5"),
				TargetBlockUtilization: 10,
				MaxBlockUtilization:    9,
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
