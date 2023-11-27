package simapp

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"
)

// PostHandlerOptions are the options required for constructing a default SDK PostHandler.
type PostHandlerOptions struct {
	AccountKeeper   feemarketante.AccountKeeper
	BankKeeper      feemarketante.BankKeeper
	FeeMarketKeeper feemarketante.FeeMarketKeeper
}

// NewPostHandler returns an empty PostHandler chain.
func NewPostHandler(options PostHandlerOptions) (sdk.PostHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for post builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for post builder")
	}

	if options.FeeMarketKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "feemarket keeper is required for post builder")
	}

	var postDecorators []sdk.PostDecorator

	return sdk.ChainPostDecorators(postDecorators...), nil
}
