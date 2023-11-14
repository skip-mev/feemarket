package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
)

// NewParams instantiates a new EIP-1559 Params object. This params object is utilized
// to implement both the base EIP-1559 fee and AIMD EIP-1559 fee market implementations.
func NewParams(
	alpha math.LegacyDec,
	beta math.LegacyDec,
	theta math.LegacyDec,
	delta math.LegacyDec,
	minBaseFee math.Int,
	minLearingRate math.LegacyDec,
	maxLearningRate math.LegacyDec,
) Params {
	return Params{
		Alpha:           alpha,
		Beta:            beta,
		Theta:           theta,
		Delta:           delta,
		MinBaseFee:      minBaseFee,
		MinLearningRate: minLearingRate,
		MaxLearningRate: maxLearningRate,
	}
}

// ValidateBasic performs basic validation on the parameters.
func (p *Params) ValidateBasic() error {
	if p.Alpha.IsNil() || p.Alpha.IsNegative() {
		return fmt.Errorf("alpha cannot be nil must be between [0, inf)")
	}

	if p.Beta.IsNil() || p.Beta.IsNegative() || p.Beta.GT(math.LegacyOneDec()) {
		return fmt.Errorf("beta cannot be nil and must be between [0, 1]")
	}

	if p.Theta.IsNil() || p.Theta.IsNegative() || p.Theta.GT(math.LegacyOneDec()) {
		return fmt.Errorf("theta cannot be nil and must be between [0, 1]")
	}

	if p.Delta.IsNil() || p.Delta.IsNegative() {
		return fmt.Errorf("delta cannot be nil and must be between [0, inf)")
	}

	if p.MinBaseFee.IsNil() || !p.MinBaseFee.GTE(math.ZeroInt()) {
		return fmt.Errorf("min base fee cannot be nil and must be greater than or equal to zero")
	}

	if p.MaxLearningRate.IsNil() || p.MinLearningRate.IsNegative() {
		return fmt.Errorf("min learning rate cannot be negative or nil")
	}

	if p.MaxLearningRate.IsNil() || p.MaxLearningRate.IsNegative() {
		return fmt.Errorf("max learning rate cannot be negative or nil")
	}

	if p.MinLearningRate.GT(p.MaxLearningRate) {
		return fmt.Errorf("min learning rate cannot be greater than max learning rate")
	}

	return nil
}
