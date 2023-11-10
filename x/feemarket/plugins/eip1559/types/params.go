package types

import "cosmossdk.io/math"

var (
	// DefaultWindow is the default window size for the sliding window
	// used to calculate the base fee.
	DefaultWindow uint64 = 8

	// DefaultAlpha is the default alpha value for the learning
	// rate calculation. This value determines how much we want to additively
	// increase the learning rate when the target block size is exceeded.
	DefaultAlpha math.LegacyDec = math.LegacyMustNewDecFromStr("0.025")

	// DefaultBeta is the default beta value for the learning rate
	// calculation. This value determines how much we want to multiplicatively
	// decrease the learning rate when the target utilization is not met.
	DefaultBeta math.LegacyDec = math.LegacyMustNewDecFromStr("0.95")

	// DefaultTheta is the default threshold for determining whether
	// to increase or decrease the learning rate. In this case, we increase
	// the learning rate if the block utilization within the window is greater
	// than 0.75 or less than 0.25. Otherwise, we multiplicatively decrease
	// the learning rate.
	DefaultTheta math.LegacyDec = math.LegacyMustNewDecFromStr("0.25")

	// DefaultDelta is the default delta value for how much we additively increase
	// or decrease the base fee when the net block utilization within the window
	// is not equal to the target utilization.
	DefaultDelta math.LegacyDec = math.LegacyMustNewDecFromStr("0.0")

	// DefaultTargetBlockSize is the default target block size. This is the default
	// on Ethereum.
	DefaultTargetBlockSize uint64 = 15_000_000

	// DefaultMaxBlockSize is the default maximum block size. This is the default
	// on Ethereum.
	DefaultMaxBlockSize uint64 = 30_000_000

	// DefaultMinBaseFee is the default minimum base fee. This is the default
	// on Ethereum.
	DefaultMinBaseFee math.Int = math.NewInt(1_000_000_000)

	// DefaultMinLearningRate is the default minimum learning rate.
	DefaultMinLearningRate math.LegacyDec = math.LegacyMustNewDecFromStr("0.01")

	// DefaultMaxLearningRate is the default maximum learning rate.
	DefaultMaxLearningRate math.LegacyDec = math.LegacyMustNewDecFromStr("0.50")
)

// NewParams instantiates a new EIP-1559 Params object.
func NewParams(
	window uint64,
	alpha math.LegacyDec,
	beta math.LegacyDec,
	theta math.LegacyDec,
	delta math.LegacyDec,
	targetBlockSize uint64,
	maxBlockSize uint64,
	minBaseFee math.Int,
	minLearingRate math.LegacyDec,
	maxLearningRate math.LegacyDec,
) Params {
	return Params{
		Window:          window,
		Alpha:           alpha,
		Beta:            beta,
		Theta:           theta,
		Delta:           delta,
		TargetBlockSize: targetBlockSize,
		MaxBlockSize:    maxBlockSize,
		MinBaseFee:      minBaseFee,
		MinLearningRate: minLearingRate,
		MaxLearningRate: maxLearningRate,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		Window:          DefaultWindow,
		Alpha:           DefaultAlpha,
		Beta:            DefaultBeta,
		Theta:           DefaultTheta,
		Delta:           DefaultDelta,
		TargetBlockSize: DefaultTargetBlockSize,
		MaxBlockSize:    DefaultMaxBlockSize,
		MinBaseFee:      DefaultMinBaseFee,
		MinLearningRate: DefaultMinLearningRate,
		MaxLearningRate: DefaultMaxLearningRate,
	}
}
