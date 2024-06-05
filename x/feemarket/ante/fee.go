package ante

import (
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

type feeMarketCheckDecorator struct {
	feemarketKeeper FeeMarketKeeper
}

func newFeeMarketCheckDecorator(fmk FeeMarketKeeper) feeMarketCheckDecorator {
	return feeMarketCheckDecorator{
		feemarketKeeper: fmk,
	}
}

// FeeMarketCheckDecorator checks sufficient fees from the fee payer based off of the current
// state of the feemarket.
// If the fee payer does not have the funds to pay for the fees, return an InsufficientFunds error.
// Call next AnteHandler if fees successfully checked.
//
// If x/feemarket is disabled (params.Enabled == false), the handler will fall back to the default
// Cosmos SDK fee deduction antehandler.
//
// CONTRACT: Tx must implement FeeTx interface
type FeeMarketCheckDecorator struct {
	feemarketKeeper    FeeMarketKeeper
	feemarketDecorator feeMarketCheckDecorator
	cosmosDecorator    ante.DeductFeeDecorator
}

func NewFeeMarketCheckDecorator(fmk FeeMarketKeeper, ak AccountKeeper, bk BankKeeper, fgk FeeGrantKeeper, txfc ante.TxFeeChecker) FeeMarketCheckDecorator {
	return FeeMarketCheckDecorator{
		feemarketKeeper: fmk,
		feemarketDecorator: newFeeMarketCheckDecorator(
			fmk,
		),
		cosmosDecorator: ante.NewDeductFeeDecorator(ak, bk, fgk, txfc),
	}
}

// AnteHandle calls the feemarket internal antehandler if the keeper is enabled.  If disabled, the Cosmos SDK
// fee antehandler is fallen back to.
func (d FeeMarketCheckDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	params, err := d.feemarketKeeper.GetParams(ctx)
	if err != nil {
		return ctx, err
	}
	if params.Enabled {
		return d.feemarketDecorator.anteHandle(ctx, tx, simulate, next)
	}
	return d.cosmosDecorator.AnteHandle(ctx, tx, simulate, next)
}

// anteHandle checks if the tx provides sufficient fee to cover the required fee from the fee market.
func (dfd feeMarketCheckDecorator) anteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
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

	params, err := dfd.feemarketKeeper.GetParams(ctx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get fee market params")
	}

	// return if disabled
	if !params.Enabled {
		return next(ctx, tx, simulate)
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas() // use provided gas limit

	if len(feeCoins) != 1 {
		if len(feeCoins) == 0 {
			return ctx, errorsmod.Wrapf(feemarkettypes.ErrNoFeeCoins, "got length %d", len(feeCoins))
		}
		return ctx, errorsmod.Wrapf(feemarkettypes.ErrTooManyFeeCoins, "got length %d", len(feeCoins))
	}

	feeCoin := feeCoins[0]
	feeGas := int64(feeTx.GetGas())

	minGasPrice, err := dfd.feemarketKeeper.GetMinGasPrice(ctx, feeCoin.GetDenom())
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get min gas price for denom %s", feeCoin.GetDenom())
	}

	ctx.Logger().Info("fee deduct ante handle",
		"min gas prices", minGasPrice,
		"fee", feeCoins,
		"gas limit", gas,
	)

	ctx = ctx.WithMinGasPrices(sdk.NewDecCoins(minGasPrice))

	if !simulate {
		_, _, err := CheckTxFee(ctx, minGasPrice, feeCoin, feeGas, true)
		if err != nil {
			return ctx, errorsmod.Wrapf(err, "error checking fee")
		}

		priorityFee, err := dfd.resolveTxPriorityCoins(ctx, feeCoin, params.FeeDenom)
		if err != nil {
			return ctx, errorsmod.Wrapf(err, "error resolving fee priority")
		}

		baseGasPrice, err := dfd.feemarketKeeper.GetMinGasPrice(ctx, params.FeeDenom)
		if err != nil {
			return ctx, err
		}

		ctx = ctx.WithPriority(GetTxPriority(priorityFee, int64(gas), baseGasPrice))
		return next(ctx, tx, simulate)
	}
	return next(ctx, tx, simulate)
}

// resolveTxPriorityCoins converts the coins to the proper denom used for tx prioritization calculation.
func (dfd feeMarketCheckDecorator) resolveTxPriorityCoins(ctx sdk.Context, fee sdk.Coin, baseDenom string) (sdk.Coin, error) {
	if fee.Denom == baseDenom {
		return fee, nil
	}

	feeDec := sdk.NewDecCoinFromCoin(fee)
	convertedDec, err := dfd.feemarketKeeper.GetDenomResolver().ConvertToDenom(ctx, feeDec, baseDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// truncate down
	return sdk.NewCoin(baseDenom, convertedDec.Amount.TruncateInt()), nil
}

// CheckTxFee implements the logic for the fee market to check if a Tx has provided sufficient
// fees given the current state of the fee market. Returns an error if insufficient fees.
func CheckTxFee(ctx sdk.Context, gasPrice sdk.DecCoin, feeCoin sdk.Coin, feeGas int64, isAnte bool) (payCoin sdk.Coin, tip sdk.Coin, err error) {
	payCoin = feeCoin

	// Ensure that the provided fees meet the minimum
	if !gasPrice.IsZero() {
		var (
			requiredFee sdk.Coin
			consumedFee sdk.Coin
		)

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas, where fee = ceil(minGasPrice * gas).
		gasConsumed := int64(ctx.GasMeter().GasConsumed())
		gcDec := sdkmath.LegacyNewDec(gasConsumed)
		glDec := sdkmath.LegacyNewDec(feeGas)

		consumedFeeAmount := gasPrice.Amount.Mul(gcDec)
		limitFee := gasPrice.Amount.Mul(glDec)

		consumedFee = sdk.NewCoin(gasPrice.Denom, consumedFeeAmount.Ceil().RoundInt())
		requiredFee = sdk.NewCoin(gasPrice.Denom, limitFee.Ceil().RoundInt())

		if !payCoin.IsGTE(requiredFee) {
			return sdk.Coin{}, sdk.Coin{}, sdkerrors.ErrInsufficientFee.Wrapf(
				"got: %s required: %s, minGasPrice: %s, gas: %d",
				payCoin,
				requiredFee,
				gasPrice,
				gasConsumed,
			)
		}

		if isAnte {
			tip = payCoin.Sub(requiredFee)
			payCoin = requiredFee
		} else {
			tip = payCoin.Sub(consumedFee)
			payCoin = consumedFee
		}
	}

	return payCoin, tip, nil
}

const (
	// gasPricePrecision is the amount of digit precision to scale the gas prices to.
	gasPricePrecision = 6
)

// GetTxPriority returns a naive tx priority based on the amount of gas price provided in a transaction.
//
// The fee amount is divided by the gasLimit to calculate "Effective Gas Price".
// This value is then normalized and scaled into an integer, so it can be used as a priority.
//
//	effectiveGasPrice = feeAmount / gas limit (denominated in fee per gas)
//	normalizedGasPrice = effectiveGasPrice / currentGasPrice (floor is 1.  The minimum effective gas price can ever be is current gas price)
//	scaledGasPrice = normalizedGasPrice * 10 ^ gasPricePrecision (amount of decimal places in the normalized gas price to consider when converting to int64).
func GetTxPriority(fee sdk.Coin, gasLimit int64, currentGasPrice sdk.DecCoin) int64 {
	effectiveGasPrice := fee.Amount.ToLegacyDec().QuoInt64(gasLimit)
	normalizedGasPrice := effectiveGasPrice.Quo(currentGasPrice.Amount)
	scaledGasPrice := normalizedGasPrice.MulInt64(int64(math.Pow10(gasPricePrecision)))

	// overflow panic protection
	if scaledGasPrice.GTE(sdkmath.LegacyNewDec(math.MaxInt64)) {
		return math.MaxInt64
	} else if scaledGasPrice.LTE(sdkmath.LegacyOneDec()) {
		return 0
	}

	return scaledGasPrice.TruncateInt64()
}
