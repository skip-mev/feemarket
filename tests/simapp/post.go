package simapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PostHandlerOptions are the options required for constructing a default SDK PostHandler.
type PostHandlerOptions struct{}

// NewPostHandler returns an empty PostHandler chain.
func NewPostHandler(_ PostHandlerOptions) (sdk.PostHandler, error) {
	var postDecorators []sdk.PostDecorator

	return sdk.ChainPostDecorators(postDecorators...), nil
}
