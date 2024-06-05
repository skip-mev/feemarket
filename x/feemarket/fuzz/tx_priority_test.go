package fuzz_test

import (
	"math"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/skip-mev/feemarket/x/feemarket/ante"
)

type input struct {
	payFee          sdk.Coin
	gasLimit        int64
	currentGasPrice sdk.DecCoin
}

// TestGetTxPriority ensures that tx priority is properly bounded
func TestGetTxPriority(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		inputs := CreateRandomInput(t)

		priority := ante.GetTxPriority(inputs.payFee, inputs.gasLimit, inputs.currentGasPrice)
		require.GreaterOrEqual(t, priority, 0)
		require.LessOrEqual(t, priority, math.MaxInt64)
	})
}

// CreateRandomInput returns a random inputs to the priority function.
func CreateRandomInput(t *rapid.T) input {
	denom := "skip"

	price := rapid.Int64Range(1, math.MaxInt64).Draw(t, "gas price")
	gasLimit := rapid.Int64Range(1, math.MaxInt64).Draw(t, "gas limit")
	priceDec := sdkmath.LegacyNewDecWithPrec(price, 1000)

	payFeeAmt := rapid.Int64Range(priceDec.MulInt64(gasLimit).TruncateInt64(), math.MaxInt64).Draw(t, "fee amount")

	return input{
		payFee:          sdk.NewCoin(denom, sdkmath.NewInt(payFeeAmt)),
		gasLimit:        gasLimit,
		currentGasPrice: sdk.NewDecCoinFromDec(denom, priceDec),
	}
}
