package ante

import (
	"fmt"
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

// TxFeeChecker check if the provided fee is enough and returns the effective fee and tx priority,
// the effective fee should be deducted later, and the priority should be returned in abci response.
type TxFeeChecker func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error)

// FeeMarketDecorator deducts fees from the fee payer based off of the current state of the feemarket.
// The fee payer is the fee granter (if specified) or first signer of the tx.
// If the fee payer does not have the funds to pay for the fees, return an InsufficientFunds error.
// Call next AnteHandler if fees successfully deducted.
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type FeeMarketDecorator struct {
	accountKeeper   AccountKeeper
	bankKeeper      BankKeeper
	feegrantKeeper  FeeGrantKeeper
	feemarketKeeper FeeMarketKeeper
}

func NewFeeMarketDecorator(ak AccountKeeper, bk BankKeeper, fk FeeGrantKeeper, fmk FeeMarketKeeper) FeeMarketDecorator {
	return FeeMarketDecorator{
		accountKeeper:   ak,
		bankKeeper:      bk,
		feegrantKeeper:  fk,
		feemarketKeeper: fmk,
	}
}

func (dfd FeeMarketDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	var (
		priority int64
		tip      sdk.Coins
	)

	minGasPrices, err := dfd.feemarketKeeper.GetMinGasPrices(ctx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get fee market state")
	}

	minGasPricesDec := sdk.NewDecCoinsFromCoins(minGasPrices...)

	fee := feeTx.GetFee()
	if !simulate {
		fee, tip, priority, err = checkTxFees(ctx, minGasPricesDec, tx)
		if err != nil {
			return ctx, err
		}
	}

	if err := dfd.checkDeductFeeAndTip(ctx, tx, fee, tip); err != nil {
		return ctx, err
	}

	newCtx := ctx.WithPriority(priority).WithMinGasPrices(minGasPricesDec)
	return next(newCtx, tx, simulate)
}

func (dfd FeeMarketDecorator) checkDeductFeeAndTip(ctx sdk.Context, sdkTx sdk.Tx, fee, tip sdk.Coins) error {
	feeTx, ok := sdkTx.(sdk.FeeTx)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.accountKeeper.GetModuleAddress(feemarkettypes.FeeCollectorName); addr == nil {
		return fmt.Errorf("fee collector module account (%s) has not been set", feemarkettypes.FeeCollectorName)
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return sdkerrors.ErrInvalidRequest.Wrap("fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, sdkTx.GetMsgs())
			if err != nil {
				return errorsmod.Wrapf(err, "%s does not allow to pay fees for %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.accountKeeper.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return sdkerrors.ErrUnknownAddress.Wrapf("fee payer address: %s does not exist", deductFeesFrom)
	}

	// deduct the fees
	if !fee.IsZero() {
		err := DeductCoins(dfd.bankKeeper, ctx, deductFeesFromAcc, fee)
		if err != nil {
			return err
		}
	}

	// deduct the tip
	if !tip.IsZero() {
		err := DeductCoins(dfd.bankKeeper, ctx, deductFeesFromAcc, tip)
		if err != nil {
			return err
		}
	}

	events := sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
			sdk.NewAttribute(feemarkettypes.AttributeKeyTip, tip.String()),
			sdk.NewAttribute(feemarkettypes.AttributeKeyTipPayer, sdk.AccAddress(ctx.BlockHeader().ProposerAddress).String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}

// DeductCoins deducts coins from the given account.  Coins are sent to the feemarket module account.
func DeductCoins(bankKeeper BankKeeper, ctx sdk.Context, acc authtypes.AccountI, coins sdk.Coins) error {
	if !coins.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "invalid coin amount: %s", coins)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), feemarkettypes.FeeCollectorName, coins)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}

// checkTxFees implements the logic for the fee market to check if a Tx has provided suffucient
// fees given the current state of the fee market. Returns an error if insufficient fees.
func checkTxFees(_ sdk.Context, fees sdk.DecCoins, tx sdk.Tx) (feeCoins sdk.Coins, tip sdk.Coins, priority int64, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, nil, 0, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins = feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet the minimum
	minGasPrices := fees
	if !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdkmath.LegacyNewDec(int64(gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !feeCoins.IsAnyGTE(requiredFees) {
			return nil, nil, 0, errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
		}

		tip = feeCoins.Sub(requiredFees...) // tip is whatever is left
		feeCoins = requiredFees             //  set free coins to be ONLY the required amount
	}

	priority = getTxPriority(feeCoins, int64(gas))
	return feeCoins, tip, priority, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
// NOTE: This implementation should be used with a great consideration as it opens potential attack vectors
// where txs with multiple coins could not be prioritize as expected.
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
