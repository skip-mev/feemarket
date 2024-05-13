package ante

import (
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
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

	minFee := minGasPrices[0]

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas() // use provided gas limit

	if len(feeCoins) != 1 {
		return ctx, feemarkettypes.ErrTooManyFeeCoins
	}

	ctx.Logger().Info("fee deduct ante handle",
		"min gas prices", minFee,
		"fee", feeCoins,
		"gas limit", gas,
	)

	if !simulate {
		fee, _, err := CheckTxFee(ctx, minGasPrices, feeTx, true, dfd.feemarketKeeper.GetDenomResolver())
		if err != nil {
			return ctx, errorsmod.Wrapf(err, "error checking fee")
		}
		priorityCtx := ctx.WithPriority(getTxPriority(fee, int64(gas))).WithMinGasPrices(minGasPrices)
		return next(priorityCtx, tx, simulate)
	}

	ctx = ctx.WithMinGasPrices(minGasPrices)
	return next(ctx, tx, simulate)
}

// CheckTxFee implements the logic for the fee market to check if a Tx has provided sufficient
// fees given the current state of the fee market. Returns an error if insufficient fees.
func CheckTxFee(ctx sdk.Context, minFeesDecCoins sdk.DecCoins, feeTx sdk.FeeTx, isCheck bool, resolver feemarkettypes.DenomResolver) (payCoin sdk.Coin, tip sdk.Coin, err error) {
	if len(feeTx.GetFee()) != 1 {
		return sdk.Coin{}, sdk.Coin{}, feemarkettypes.ErrTooManyFeeCoins
	}

	feeDenom := minFeesDecCoins.GetDenomByIndex(0)
	feeCoin := feeTx.GetFee()[0]

	coinWithBaseDenom := feeCoin
	if feeCoin.Denom != feeDenom {
		coinWithBaseDenom, err = resolver.ConvertToDenom(ctx, feeCoin, feeDenom)
		if err != nil {
			return sdk.Coin{}, sdk.Coin{}, err
		}
	}

	// Ensure that the provided fees meet the minimum
	minGasPrice := minFeesDecCoins[0]
	if !minGasPrice.IsZero() {
		var (
			requiredFee sdk.Coin
			consumedFee sdk.Coin
		)

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas, where fee = ceil(minGasPrice * gas).
		gasConsumed := int64(ctx.GasMeter().GasConsumed())
		gcDec := sdkmath.LegacyNewDec(gasConsumed)
		glDec := sdkmath.LegacyNewDec(int64(feeTx.GetGas()))

		consumedFeeAmount := minGasPrice.Amount.Mul(gcDec)
		limitFee := minGasPrice.Amount.Mul(glDec)
		consumedFee = sdk.NewCoin(minGasPrice.Denom, consumedFeeAmount.Ceil().RoundInt())
		requiredFee = sdk.NewCoin(minGasPrice.Denom, limitFee.Ceil().RoundInt())

		if coinWithBaseDenom.Denom != requiredFee.Denom || !coinWithBaseDenom.IsGTE(requiredFee) {
			return sdk.Coin{}, sdk.Coin{}, sdkerrors.ErrInsufficientFee.Wrapf(
				"got: %s required: %s, minGasPrice: %s, gas: %d",
				coinWithBaseDenom,
				requiredFee,
				minGasPrice,
				gasConsumed,
			)
		}

		// convert back to given denom is not base denom
		if feeCoin.Denom != requiredFee.Denom {
			requiredFee, err = resolver.ConvertToDenom(ctx, requiredFee, feeCoin.Denom)
			if err != nil {
				return sdk.Coin{}, sdk.Coin{}, err
			}

			if !isCheck {
				consumedFee, err = resolver.ConvertToDenom(ctx, consumedFee, feeCoin.Denom)
				if err != nil {
					return sdk.Coin{}, sdk.Coin{}, err
				}
			}
		}

		if isCheck {
			//  set fee coins to be required amount if checking
			feeCoin = requiredFee
		} else {
			// tip is the difference between feeCoin and the required fee
			tip = feeCoin.Sub(requiredFee)
			// set fee coin to be ONLY the consumed amount if we are calculated consumed fee to deduct
			feeCoin = consumedFee
		}
	}

	return feeCoin, tip, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
// NOTE: This implementation should be used with a great consideration as it opens potential attack vectors
// where txs with multiple coins could not be prioritized as expected.
func getTxPriority(fee sdk.Coin, gas int64) int64 {
	var priority int64
	p := int64(math.MaxInt64)
	gasPrice := fee.Amount.QuoRaw(gas)
	if gasPrice.IsInt64() {
		p = gasPrice.Int64()
	}
	if priority == 0 || p < priority {
		priority = p
	}

	return priority
}
