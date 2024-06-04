package app

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"
)

// AnteHandlerOptions are the options required for constructing an SDK AnteHandler with the fee market injected.
type AnteHandlerOptions struct {
	BaseOptions     authante.HandlerOptions
	AccountKeeper   feemarketante.AccountKeeper
	FeeMarketKeeper feemarketante.FeeMarketKeeper
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options AnteHandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BaseOptions.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.BaseOptions.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	if options.FeeMarketKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "feemarket keeper is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		authante.NewExtensionOptionsDecorator(options.BaseOptions.ExtensionOptionChecker),
		authante.NewValidateBasicDecorator(),
		authante.NewTxTimeoutHeightDecorator(),
		authante.NewValidateMemoDecorator(options.AccountKeeper),
		authante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		feemarketante.NewFeeMarketCheckDecorator( // fee market check replaces fee deduct decorator
			options.FeeMarketKeeper,
			options.AccountKeeper,
			options.BaseOptions.BankKeeper,
			options.BaseOptions.FeegrantKeeper,
			options.BaseOptions.TxFeeChecker,
		), // fees are deducted in the fee market deduct post handler
		authante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		authante.NewValidateSigCountDecorator(options.AccountKeeper),
		authante.NewSigGasConsumeDecorator(options.AccountKeeper, options.BaseOptions.SigGasConsumer),
		authante.NewSigVerificationDecorator(options.AccountKeeper, options.BaseOptions.SignModeHandler),
		authante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
