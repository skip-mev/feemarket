package app

import (
	"github.com/cosmos/cosmos-sdk/x/auth/ante/unorderedtx"

	circuitante "cosmossdk.io/x/circuit/ante"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"
)

// AnteHandlerOptions are the options required for constructing an SDK AnteHandler with the fee market injected.
type AnteHandlerOptions struct {
	ante.HandlerOptions
	BankKeeper      feemarketante.BankKeeper
	AccountKeeper   feemarketante.AccountKeeper
	FeeMarketKeeper feemarketante.FeeMarketKeeper
	CircuitKeeper   circuitante.CircuitBreaker
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options AnteHandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.CircuitKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "circuit keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "base options bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	if options.FeeMarketKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "feemarket keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper keeper is required for ante builder")
	}

	if options.FeegrantKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "feegrant keeper is required for ante builder")
	}

	feemarketDecorator := feemarketante.NewFeeMarketCheckDecorator( // fee market check replaces fee deduct decorator
		options.AccountKeeper,
		options.BankKeeper,
		options.FeegrantKeeper,
		options.FeeMarketKeeper,
		ante.NewDeductFeeDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			options.FeegrantKeeper,
			options.TxFeeChecker,
		),
	) // fees are deducted in the fee market deduct post handler
	_ = feemarketDecorator
	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(options.Environment, options.ConsensusKeeper), // outermost AnteDecorator. SetUpContext must be called first
		circuitante.NewCircuitBreakerDecorator(options.CircuitKeeper),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(options.Environment),
		ante.NewTxTimeoutHeightDecorator(options.Environment),
		ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, options.UnorderedTxManager, options.Environment, ante.DefaultSha256Cost),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler, options.SigGasConsumer, options.AccountAbstractionKeeper),
		feemarketDecorator,
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
