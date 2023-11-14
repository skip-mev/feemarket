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
	window,
	targetBlockUtilization, maxBlockUtilization uint64,
) State {
	return State{
		BaseFee:                baseFee,
		LearningRate:           learningRate,
		BlockUtilizationWindow: make([]uint64, window),
		TargetBlockUtilization: targetBlockUtilization,
		MaxBlockUtilization:    maxBlockUtilization,
	}
}

// GetNetUtilization returns the net utilization of the block window. This
// is utilized to update base fee.
func (s *State) GetNetUtilization() math.Int {
	net := math.NewInt(0)

	targetUtilization := math.NewIntFromUint64(s.TargetBlockUtilization)
	for _, utilization := range s.BlockUtilizationWindow {
		diff := math.NewIntFromUint64(utilization).Sub(targetUtilization)
		net = net.Add(diff)
	}

	return net
}

// GetAverageUtilization returns the average utilization of the block
// window. This is utilization to both update the learning rate and base fee.
func (s *State) GetAverageUtilization() math.LegacyDec {
	var total uint64
	for _, utilization := range s.BlockUtilizationWindow {
		total += utilization
	}

	sum := math.LegacyNewDecFromInt(math.NewIntFromUint64(total))

	multiple := math.LegacyNewDecFromInt(math.NewIntFromUint64(uint64(len(s.BlockUtilizationWindow))))
	target := math.LegacyNewDecFromInt(math.NewIntFromUint64(s.TargetBlockUtilization)).Mul(multiple)

	return sum.Quo(target)
}

// Update will update the state based on the transaction's gas wanted.
func (s *State) Update(gasWanted uint64) error {
	blockUtilization := s.BlockUtilizationWindow[s.Index]

	updatedUtilization := blockUtilization + gasWanted
	if updatedUtilization > s.MaxBlockUtilization {
		return fmt.Errorf("block utilization %d cannot exceed max block utilization %d", updatedUtilization, s.MaxBlockUtilization)
	}

	s.BlockUtilizationWindow[s.Index] = updatedUtilization
	return nil
}

// IncrementHeight increments the height of state. This is used to
// start a new block entry in the block utilization window.
func (s *State) IncrementHeight() {
	s.Index = (s.Index + 1) % uint64(len(s.BlockUtilizationWindow))
	s.BlockUtilizationWindow[s.Index] = 0
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

	if s.TargetBlockUtilization == 0 {
		return fmt.Errorf("target block utilization cannot be zero")
	}

	if s.TargetBlockUtilization > s.MaxBlockUtilization {
		return fmt.Errorf("target block utilization cannot be greater than max block size")
	}

	return nil
}
