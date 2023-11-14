package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
)

// NewState instantiates a new fee market state instance. This is utilized
// to implement both the base EIP-1559 fee market implementation and the
// AIMD EIP-1559 fee market implementation.
func NewState(window uint64, baseFee math.Int, learningRate math.LegacyDec) State {
	return State{
		Window:       make([]uint64, window),
		BaseFee:      baseFee,
		LearningRate: learningRate,
		Index:        0,
	}
}

// ValidateBasic performs basic validation on the state.
func (s *State) ValidateBasic() error {
	if s.Window == nil || len(s.Window) == 0 {
		return fmt.Errorf("block utilization window cannot be nil or empty")
	}

	if s.BaseFee.IsNil() || s.BaseFee.LTE(math.ZeroInt()) {
		return fmt.Errorf("base fee must be positive")
	}

	if s.LearningRate.IsNil() || s.LearningRate.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("learning rate must be positive")
	}

	return nil
}
