package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
)

// NewState instantiates a new EIP-1559 State object. This can
// be used to implement both the base EIP-1559 fee and AIMD EIP-1559
// fee market implementations.
func NewState(
	baseFee math.Int,
	learningRate math.LegacyDec,
	window uint64,
) State {
	return State{
		BaseFee:                baseFee,
		LearningRate:           learningRate,
		BlockUtilizationWindow: make([]uint64, window),
	}
}

// ValidateBasic performs basic validation on the state.
func (s *State) ValidateBasic() error {
	if s.BaseFee.IsNil() || s.BaseFee.LT(math.ZeroInt()) {
		return fmt.Errorf("current base fee cannot be nil, negative or zero")
	}

	if s.LearningRate.IsNil() || s.LearningRate.LT(math.LegacyZeroDec()) {
		return fmt.Errorf("current learning rate cannot be nil or negative")
	}

	if s.BlockUtilizationWindow == nil || len(s.BlockUtilizationWindow) == 0 {
		return fmt.Errorf("block utilization window cannot be nil or empty")
	}

	return nil
}
