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

// UpdateBaseFee updates the learning rate and base fee based on the AIMD
// learning rate adjustment algorithm. The learning rate is updated
// based on the average utilization of the block window. The base fee is
// update using the new learning rate and the delta adjustment. Please
// see the EIP-1559 specification for more details.
func (s *State) UpdateBaseFee(params Params) math.Int {
	// Update the learning rate.
	newLR := s.UpdateLearningRate(params)

	// Calculate the new base fee with the learning rate adjustment.
	currentBlockSize := math.LegacyNewDecFromInt(math.NewIntFromUint64(s.Window[s.Index]))
	targetBlockSize := math.LegacyNewDecFromInt(math.NewIntFromUint64(params.TargetBlockUtilization))
	utilization := (currentBlockSize.Sub(targetBlockSize)).Quo(targetBlockSize)

	// Truncate the learning rate adjustment to an integer.
	//
	// This is equivalent to
	// 1 + (learningRate * (currentBlockSize - targetBlockSize) / targetBlockSize)
	learningRateAdjustment := math.LegacyOneDec().Add(newLR.Mul(utilization))

	// Calculate the delta adjustment.
	net := s.GetNetUtilization(params.TargetBlockUtilization)
	delta := params.Delta.Mul(math.LegacyNewDecFromInt(net))

	// Update the base fee.
	s.BaseFee = (math.LegacyNewDecFromInt(s.BaseFee).Mul(learningRateAdjustment)).Add(delta).TruncateInt()
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
func (s *State) UpdateLearningRate(params Params) math.LegacyDec {
	// Calculate the average utilization of the block window.
	avg := s.GetAverageUtilization(params.MaxBlockUtilization)

	// Determine if the average utilization is above or below the target
	// threshold and adjust the learning rate accordingly.
	var updatedLearningRate math.LegacyDec
	if avg.LTE(params.Theta) || avg.GTE(math.LegacyOneDec().Sub(params.Theta)) {
		updatedLearningRate = params.Alpha.Add(s.LearningRate)
		if updatedLearningRate.GT(params.MaxLearningRate) {
			updatedLearningRate = params.MaxLearningRate
		}
	} else {
		updatedLearningRate = s.LearningRate.Mul(params.Beta)
		if updatedLearningRate.LT(params.MinLearningRate) {
			updatedLearningRate = params.MinLearningRate
		}
	}

	// Update the current learning rate.
	s.LearningRate = updatedLearningRate
	return s.LearningRate
}

// GetNetUtilization returns the net utilization of the block window.
func (s *State) GetNetUtilization(target uint64) math.Int {
	net := math.NewInt(0)

	targetUtilization := math.NewIntFromUint64(target)
	for _, utilization := range s.Window {
		diff := math.NewIntFromUint64(utilization).Sub(targetUtilization)
		net = net.Add(diff)
	}

	return net
}

// GetAverageUtilization returns the average utilization of the block
// window.
func (s *State) GetAverageUtilization(max uint64) math.LegacyDec {
	var total uint64
	for _, utilization := range s.Window {
		total += utilization
	}

	sum := math.LegacyNewDecFromInt(math.NewIntFromUint64(total))

	multiple := math.LegacyNewDecFromInt(math.NewIntFromUint64(uint64(len(s.Window))))
	divisor := math.LegacyNewDecFromInt(math.NewIntFromUint64(max)).Mul(multiple)

	return sum.Quo(divisor)
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
