package types

import "cosmossdk.io/math"

// NewState instantiates a new EIP-1559 State object.
func NewState(
	currentBaseFee math.Int,
	window uint64,
) State {
	return State{
		CurrentBaseFee:         currentBaseFee,
		BlockUtilizationWindow: make([]uint64, window),
	}
}
