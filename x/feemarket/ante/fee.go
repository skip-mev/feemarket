package ante

import (
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// FeeMarketCheckDecorator checks sufficient fees from the fee payer based off of the current
// state of the feemarket.
// If the fee payer does not have the funds to pay for the fees, return an InsufficientFunds error.
// Call next AnteHandler if fees successfully checked.
// CONTRACT: Tx must implement FeeTx interface
type FeeMarketCheckDecorator struct {
	feemarketKeeper FeeMarketKeeper
}

func NewFeeMarketCheckDecorator(fmk FeeMarketKeeper) FeeMarketCheckDecorator {
	return FeeMarketCheckDecorator{
		feemarketKeeper: fmk,
	}
}

// AnteHandle checks if the tx provides sufficient fee to cover the required fee from the fee market.
func (dfd FeeMarketCheckDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// GenTx consume no fee
	if ctx.BlockHeight() == 0 {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, sdkerrors.ErrInvalidGasLimit.Wrapf("must provide positive gas")
	}

	minGasPrices, err := dfd.feemarketKeeper.GetMinGasPrices(ctx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get fee market state")
	}

	fee := feeTx.GetFee()
	gas := feeTx.GetGas() // use provided gas limit

	ctx.Logger().Info("fee deduct ante handle",
		"min gas prices", minGasPrices,
		"fee", fee,
		"gas limit", gas,
	)

	if !simulate {
		fee, _, err = CheckTxFees(ctx, minGasPrices, feeTx, true)
		if err != nil {
			return ctx, errorsmod.Wrapf(err, "error checking fee")
		}
		priorityCtx := ctx.WithPriority(getTxPriority(fee, int64(gas))).WithMinGasPrices(minGasPrices)
		return next(priorityCtx, tx, simulate)
	}

	ctx = ctx.WithMinGasPrices(minGasPrices)
	return next(ctx, tx, simulate)
}

// CheckTxFees implements the logic for the fee market to check if a Tx has provided sufficient
// fees given the current state of the fee market. Returns an error if insufficient fees.
func CheckTxFees(ctx sdk.Context, minFeesDecCoins sdk.DecCoins, feeTx sdk.FeeTx, isCheck bool) (feeCoins sdk.Coins, tip sdk.Coins, err error) {
	feeCoins = feeTx.GetFee()

	// Ensure that the provided fees meet the minimum
	minGasPrices := minFeesDecCoins
	if !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))
		consumedFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas, where fee = ceil(minGasPrice * gas).
		gasConsumed := int64(ctx.GasMeter().GasConsumed())
		gcDec := sdkmath.LegacyNewDec(gasConsumed)
		glDec := sdkmath.LegacyNewDec(int64(feeTx.GetGas()))

		for i, gp := range minGasPrices {
			consumedFee := gp.Amount.Mul(gcDec)
			limitFee := gp.Amount.Mul(glDec)
			consumedFees[i] = sdk.NewCoin(gp.Denom, consumedFee.Ceil().RoundInt())
			requiredFees[i] = sdk.NewCoin(gp.Denom, limitFee.Ceil().RoundInt())
		}

		if !feeCoins.IsAnyGTE(requiredFees) {
			return nil, nil, sdkerrors.ErrInsufficientFee.Wrapf(
				"got: %s required: %s, minGasPrices: %s, gas: %d",
				feeCoins,
				requiredFees,
				minGasPrices,
				gasConsumed,
			)
		}

		if isCheck {
			//  set fee coins to be required amount if checking
			feeCoins = requiredFees
		} else {
			// tip is the difference between feeCoins and the required fees
			tip = feeCoins.Sub(requiredFees...)
			// set fee coins to be ONLY the consumed amount if we are calculated consumed fee to deduct
			feeCoins = consumedFees
		}
	}

	return feeCoins, tip, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
// NOTE: This implementation should be used with a great consideration as it opens potential attack vectors
// where txs with multiple coins could not be prioritized as expected.
func getTxPriority(fee sdk.Coins, gas int64) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		gasPrice := c.Amount.QuoRaw(gas)
		if gasPrice.IsInt64() {
			p = gasPrice.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
