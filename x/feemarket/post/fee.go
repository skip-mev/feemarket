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
// Call next PostHandler if fees successfully deducted.
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

// PostHandle deducts the fee from the fee payer based on the min base fee and the gas consumed in the gasmeter.
// If there is a difference between the provided fee and the min-base fee, the difference is paid as a tip.
// Fees are sent to the x/feemarket fee-collector address.
func (dfd FeeMarketDeductDecorator) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate, success bool, next sdk.PostHandler) (sdk.Context, error) {
	// GenTx consume no fee
	if ctx.BlockHeight() == 0 {
		return next(ctx, tx, simulate, success)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	var tip sdk.Coins

	// update fee market params
	params, err := dfd.feemarketKeeper.GetParams(ctx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get fee market params")
	}

	// return if disabled
	if !params.Enabled {
		return next(ctx, tx, simulate, success)
	}

	// update fee market state
	state, err := dfd.feemarketKeeper.GetState(ctx)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to get fee market state")
	}

	baseFee := sdk.NewCoin(params.FeeDenom, state.BaseFee)
	minGasPrices := sdk.NewCoins(baseFee)

	fee := feeTx.GetFee()
	gas := ctx.GasMeter().GasConsumed() // use context gas consumed

	ctx.Logger().Info("fee deduct post handle",
		"min gas prices", minGasPrices,
		"gas consumed", gas,
	)

	if !simulate {
		fee, tip, err = ante.CheckTxFees(ctx, minGasPrices, feeTx, false)
		if err != nil {
			return ctx, err
		}
	}

	ctx.Logger().Info("fee deduct post handle",
		"fee", fee,
		"tip", tip,
	)

	if err := dfd.DeductFeeAndTip(ctx, tx, fee, tip); err != nil {
		return ctx, err
	}

	consensusMaxGas := ctx.ConsensusParams().GetBlock().GetMaxGas()
	maxGas := params.MaxBlockUtilization

	if consensusMaxGas < int64(maxGas) && consensusMaxGas >= 0 {
		maxGas = uint64(consensusMaxGas)
		ctx.Logger().Info("using consensus params max gas",
			"max gas", maxGas,
		)
	}

	err = state.Update(gas, maxGas)
	if err != nil {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrTxTooLarge, err.Error())
	}

	err = dfd.feemarketKeeper.SetState(ctx, state)
	if err != nil {
		return ctx, errorsmod.Wrapf(err, "unable to set fee market state")
	}

	return next(ctx, tx, simulate, success)
}

// DeductFeeAndTip deducts the provided fee and tip from the fee payer.
// If the tx uses a feegranter, the fee granter address will pay the fee instead of the tx signer.
func (dfd FeeMarketDeductDecorator) DeductFeeAndTip(ctx sdk.Context, sdkTx sdk.Tx, fee, tip sdk.Coins) error {
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

	var events sdk.Events

	// deduct the fees and tip
	if !fee.IsZero() {
		err := DeductCoins(dfd.bankKeeper, ctx, deductFeesFromAcc, fee)
		if err != nil {
			return err
		}

		events = append(events, sdk.NewEvent(
			feemarkettypes.EventTypeFeePay,
			sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
		))
	}

	proposer := sdk.AccAddress(ctx.BlockHeader().ProposerAddress)
	if !tip.IsZero() {
		err := SendTip(dfd.bankKeeper, ctx, deductFeesFromAcc.GetAddress(), proposer, tip)
		if err != nil {
			return err
		}

		events = append(events, sdk.NewEvent(
			feemarkettypes.EventTypeTipPay,
			sdk.NewAttribute(feemarkettypes.AttributeKeyTip, tip.String()),
			sdk.NewAttribute(feemarkettypes.AttributeKeyTipPayer, deductFeesFrom.String()),
			sdk.NewAttribute(feemarkettypes.AttributeKeyTipPayee, proposer.String()),
		))
	}

	ctx.EventManager().EmitEvents(events)
	return nil
}

// DeductCoins deducts coins from the given account.  Coins are sent to the feemarket fee collector account.
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

// SendTip sends a tip to the current block proposer.
func SendTip(bankKeeper BankKeeper, ctx sdk.Context, acc, proposer sdk.AccAddress, coins sdk.Coins) error {
	if !coins.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "invalid coin amount: %s", coins)
	}

	err := bankKeeper.SendCoins(ctx, acc, proposer, coins)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
