package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DenomResolver is an interface to convert a given token to the feemarket's base token.
type DenomResolver interface {
	// ConvertToBaseToken converts feeCoin into the equivalent amount of the token denominated in baseDenom.
	ConvertToBaseToken(ctx sdk.Context, feeCoin sdk.Coin, baseDenom string) (sdk.Coin, error)
}

// TestDenomResolver is a test implementation of the DenomResolver interface.  It returns "1atom" for all inputs.
// NOTE: DO NOT USE THIS IN PRODUCTION
type TestDenomResolver struct{}

// ConvertToBaseToken returns "1atom" for all inputs.
func (r *TestDenomResolver) ConvertToBaseToken(ctx sdk.Context, feeCoin sdk.Coin, baseDenom string) (sdk.Coin, error) {
	return sdk.NewCoin("atom", sdk.OneInt()), nil
}

// ErrorDenomResolver is a test implementation of the DenomResolver interface.  It returns an error for all inputs.
// NOTE: DO NOT USE THIS IN PRODUCTION
type ErrorDenomResolver struct{}

// ConvertToBaseToken returns an error for all inputs.
func (r *ErrorDenomResolver) ConvertToBaseToken(ctx sdk.Context, feeCoin sdk.Coin, baseDenom string) (sdk.Coin, error) {
	return sdk.Coin{}, fmt.Errorf("error resolving denom")
}
