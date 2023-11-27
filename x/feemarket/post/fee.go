package post

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/skip-mev/feemarket/x/feemarket/ante"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

// FeeMarketDeductDecorator deducts fees from the fee payer based off of the current state of the feemarket.
// The fee payer is the fee granter (if specified) or first signer of the tx.
// If the fee payer does not have the funds to pay for the fees, return an InsufficientFunds error.
// If there is an excess between the given fee and the on-chain min base fee is given as a tip.
// Call next AnteHandler if fees successfully deducted.
// CONTRACT: Tx must implement FeeTx interface
type FeeMarketDeductDecorator struct {
	accountKeeper   AccountKeeper
	bankKeeper      BankKeeper
	feegrantKeeper  FeeGrantKeeper
	feemarketKeeper FeeMarketKeeper
}

func NewFeeMarketDeductDecorator(ak AccountKeeper, bk BankKeeper, fk FeeGrantKeeper, fmk FeeMarketKeeper) FeeMarketDeductDecorator {
	return FeeMarketDeductDecorator{
		accountKeeper:   ak,
		bankKeeper:      bk,
		feegrantKeeper:  fk,
		feemarketKeeper: fmk,
	}
}

func (dfd FeeMarketDeductDecorator) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate, success bool, next sdk.PostHandler) (sdk.Context, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	var tip sdk.Coins

	minGasPrices, err := dfd.feemarketKeeper.GetMinGasPrices(ctx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get fee market state")
	}

	fee := feeTx.GetFee()
	gas := ctx.GasMeter().GasConsumed() // use context gas consumed

	if !simulate {
		fee, tip, err = ante.CheckTxFees(minGasPrices, tx, gas)
		if err != nil {
			return ctx, err
		}
	}

	if err := dfd.checkDeductFeeAndTip(ctx, tx, fee, tip); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate, success)
}

func (dfd FeeMarketDeductDecorator) checkDeductFeeAndTip(ctx sdk.Context, sdkTx sdk.Tx, fee, tip sdk.Coins) error {
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
