package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
)

// NewState instantiates a new fee market state instance. This is utilized
// to implement both the base EIP-1559 fee market implementation and the
// AIMD EIP-1559 fee market implementation. Note that on init, you initialize
// both the minimum and current base fee to the same value.
func NewState(
	window,
	target, max uint64,
	baseFee math.Int,
	learningRate math.LegacyDec,
) State {
	return State{
		Window:                 make([]uint64, window),
		BaseFee:                baseFee,
		MinBaseFee:             baseFee,
		LearningRate:           learningRate,
		Index:                  0,
		TargetBlockUtilization: target,
		MaxBlockUtilization:    max,
	}
}

// Update updates the block utilization for the current height with the given
// transaction utilization i.e. gas limit.
func (s *State) Update(gas uint64) error {
	update := s.Window[s.Index] + gas
	if update > s.MaxBlockUtilization {
		return fmt.Errorf("block utilization cannot exceed max block utilization")
	}

	s.Window[s.Index] = update
	return nil
}

// IncrementHeight increments the current height of the state.
func (s *State) IncrementHeight() {
	s.Index = (s.Index + 1) % uint64(len(s.Window))
	s.Window[s.Index] = 0
}

// UpdateBaseFee updates the learning rate and base fee based on the AIMD
// learning rate adjustment algorithm. The learning rate is updated
// based on the average utilization of the block window. The base fee is
// update using the new learning rate and the delta adjustment. Please
// see the EIP-1559 specification for more details.
func (s *State) UpdateBaseFee(delta math.LegacyDec) math.Int {
	// Calculate the new base fee with the learning rate adjustment.
	currentBlockSize := math.LegacyNewDecFromInt(math.NewIntFromUint64(s.Window[s.Index]))
	targetBlockSize := math.LegacyNewDecFromInt(math.NewIntFromUint64(s.TargetBlockUtilization))
	utilization := (currentBlockSize.Sub(targetBlockSize)).Quo(targetBlockSize)

	// Truncate the learning rate adjustment to an integer.
	//
	// This is equivalent to
	// 1 + (learningRate * (currentBlockSize - targetBlockSize) / targetBlockSize)
	learningRateAdjustment := math.LegacyOneDec().Add(s.LearningRate.Mul(utilization))

	// Calculate the delta adjustment.
	net := math.LegacyNewDecFromInt(s.GetNetUtilization()).Mul(delta)

	// Update the base fee.
	fee := (math.LegacyNewDecFromInt(s.BaseFee).Mul(learningRateAdjustment)).Add(net).TruncateInt()

	// Ensure the base fee is greater than the minimum base fee.
	if fee.LT(s.MinBaseFee) {
		fee = s.MinBaseFee
	}

	s.BaseFee = fee
	return s.BaseFee
}

// UpdateLearningRate updates the learning rate based on the AIMD
// learning rate adjustment algorithm. The learning rate is updated
// based on the average utilization of the block window. There are
// two cases that can occur:
//
//  1. The average utilization is above the target threshold. In this
//     case, the learning rate is increased by the alpha parameter. This occurs
//     when blocks are nearly full or empty.
//  2. The average utilization is below the target threshold. In this
//     case, the learning rate is decreased by the beta parameter. This occurs
//     when blocks are relatively close to the target block utilization.
//
// For more details, please see the EIP-1559 specification.
func (s *State) UpdateLearningRate(theta, alpha, beta, minLR, maxLR math.LegacyDec) math.LegacyDec {
	// Calculate the average utilization of the block window.
	avg := s.GetAverageUtilization()

	// Determine if the average utilization is above or below the target
	// threshold and adjust the learning rate accordingly.
	var updatedLearningRate math.LegacyDec
	if avg.LTE(theta) || avg.GTE(math.LegacyOneDec().Sub(theta)) {
		updatedLearningRate = alpha.Add(s.LearningRate)
		if updatedLearningRate.GT(maxLR) {
			updatedLearningRate = maxLR
		}
	} else {
		updatedLearningRate = s.LearningRate.Mul(beta)
		if updatedLearningRate.LT(minLR) {
			updatedLearningRate = minLR
		}
	}

	// Update the current learning rate.
	s.LearningRate = updatedLearningRate
	return s.LearningRate
}

// GetNetUtilization returns the net utilization of the block window.
func (s *State) GetNetUtilization() math.Int {
	net := math.NewInt(0)

	targetUtilization := math.NewIntFromUint64(s.TargetBlockUtilization)
	for _, utilization := range s.Window {
		diff := math.NewIntFromUint64(utilization).Sub(targetUtilization)
		net = net.Add(diff)
	}

	return net
}

// GetAverageUtilization returns the average utilization of the block
// window.
func (s *State) GetAverageUtilization() math.LegacyDec {
	var total uint64
	for _, utilization := range s.Window {
		total += utilization
	}

	sum := math.LegacyNewDecFromInt(math.NewIntFromUint64(total))

	multiple := math.LegacyNewDecFromInt(math.NewIntFromUint64(uint64(len(s.Window))))
	divisor := math.LegacyNewDecFromInt(math.NewIntFromUint64(s.MaxBlockUtilization)).Mul(multiple)

	return sum.Quo(divisor)
}

// ValidateBasic performs basic validation on the state.
func (s *State) ValidateBasic() error {
	if s.MaxBlockUtilization == 0 {
		return fmt.Errorf("max utilization cannot be zero")
	}

	if s.TargetBlockUtilization == 0 {
		return fmt.Errorf("target utilization cannot be zero")
	}

	if s.TargetBlockUtilization > s.MaxBlockUtilization {
		return fmt.Errorf("target utilization cannot be greater than max utilization")
	}

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
