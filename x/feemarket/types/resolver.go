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

// TestDenomResolver is a test implementation of the DenomResolver interface.  It returns "1atom" for all coins that are not the baseDenom.
// NOTE: DO NOT USE THIS IN PRODUCTION
type TestDenomResolver struct{}

// ConvertToBaseToken returns "1000000baseDenom" for all coins that are not the baseDenom.
func (r *TestDenomResolver) ConvertToBaseToken(ctx sdk.Context, feeCoin sdk.Coin, baseDenom string) (sdk.Coin, error) {
	if feeCoin.Denom == baseDenom {
		return feeCoin, nil
	}

	return sdk.NewCoin(baseDenom, sdk.NewInt(1_000_000)), nil
}

// ErrorDenomResolver is a test implementation of the DenomResolver interface.  It returns an error for all coins that are not the baseDenom.
// NOTE: DO NOT USE THIS IN PRODUCTION
type ErrorDenomResolver struct{}

// ConvertToBaseToken returns an error for all coins that are not the baseDenom.
func (r *ErrorDenomResolver) ConvertToBaseToken(ctx sdk.Context, feeCoin sdk.Coin, baseDenom string) (sdk.Coin, error) {
	if feeCoin.Denom == baseDenom {
		return feeCoin, nil
	}

	return sdk.Coin{}, fmt.Errorf("error resolving denom")
}
