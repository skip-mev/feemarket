package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Note: We use the same default values as Ethereum for the EIP-1559
// fee market implementation. These parameters do not implement the
// AIMD learning rate adjustment algorithm.

var (
	// DefaultWindow is the default window size for the sliding window
	// used to calculate the base fee. In the base EIP-1559 implementation,
	// only the previous block is considered.
	DefaultWindow uint64 = 1

	// DefaultAlpha is not used in the base EIP-1559 implementation.
	DefaultAlpha = math.LegacyMustNewDecFromStr("0.0")

	// DefaultBeta is not used in the base EIP-1559 implementation.
	DefaultBeta = math.LegacyMustNewDecFromStr("1.0")

	// DefaultTheta is not used in the base EIP-1559 implementation.
	DefaultTheta = math.LegacyMustNewDecFromStr("0.0")

	// DefaultDelta is not used in the base EIP-1559 implementation.
	DefaultDelta = math.LegacyMustNewDecFromStr("0.0")

	// DefaultTargetBlockSize is the default target block utilization. This is the default
	// on Ethereum. This denominated in units of gas consumed in a block.
	DefaultTargetBlockSize uint64 = 15_000_000

	// DefaultMaxBlockSize is the default maximum block utilization. This is the default
	// on Ethereum. This denominated in units of gas consumed in a block.
	DefaultMaxBlockSize uint64 = 30_000_000

	// DefaultMinBaseFee is the default minimum base fee. This is the default
	// on Ethereum. Note that Ethereum is denominated in 1e18 wei. Cosmos chains will
	// likely want to change this to 1e6.
	DefaultMinBaseFee = math.NewInt(1_000_000_000)

	// DefaultMinLearningRate is not used in the base EIP-1559 implementation.
	DefaultMinLearningRate = math.LegacyMustNewDecFromStr("0.125")

	// DefaultMaxLearningRate is not used in the base EIP-1559 implementation.
	DefaultMaxLearningRate = math.LegacyMustNewDecFromStr("0.125")

	// DefaultFeeDenom is the Cosmos SDK default bond denom.
	DefaultFeeDenom = sdk.DefaultBondDenom
)

// DefaultParams returns a default set of parameters that implements
// the EIP-1559 fee market implementation without the AIMD learning
// rate adjustment algorithm.
func DefaultParams() Params {
	return NewParams(
		DefaultWindow,
		DefaultAlpha,
		DefaultBeta,
		DefaultTheta,
		DefaultDelta,
		DefaultTargetBlockSize,
		DefaultMaxBlockSize,
		DefaultMinBaseFee,
		DefaultMinLearningRate,
		DefaultMaxLearningRate,
		DefaultFeeDenom,
		true,
	)
}

// DefaultState returns the default state for the EIP-1559 fee market
// implementation without the AIMD learning rate adjustment algorithm.
func DefaultState() State {
	return NewState(
		DefaultWindow,
		DefaultTargetBlockSize,
		DefaultMaxBlockSize,
		DefaultMinBaseFee,
		DefaultMinLearningRate,
	)
}

// DefaultGenesisState returns a default genesis state that implements
// the EIP-1559 fee market implementation without the AIMD learning
// rate adjustment algorithm.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), DefaultState())
}
