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

// TestDenomResolver is a test implementation of the DenomResolver interface.  It returns "feeCoin.Amount baseDenom" for all coins that are not the baseDenom.
// NOTE: DO NOT USE THIS IN PRODUCTION
type TestDenomResolver struct{}

// ConvertToBaseToken returns "feeCoin.Amount baseDenom" for all coins that are not the baseDenom.
func (r *TestDenomResolver) ConvertToBaseToken(_ sdk.Context, feeCoin sdk.Coin, baseDenom string) (sdk.Coin, error) {
	if feeCoin.Denom == baseDenom {
		return feeCoin, nil
	}

	return sdk.NewCoin(baseDenom, feeCoin.Amount), nil
}

// ErrorDenomResolver is a test implementation of the DenomResolver interface.  It returns an error for all coins that are not the baseDenom.
// NOTE: DO NOT USE THIS IN PRODUCTION
type ErrorDenomResolver struct{}

// ConvertToBaseToken returns an error for all coins that are not the baseDenom.
func (r *ErrorDenomResolver) ConvertToBaseToken(_ sdk.Context, feeCoin sdk.Coin, baseDenom string) (sdk.Coin, error) {
	if feeCoin.Denom == baseDenom {
		return feeCoin, nil
	}

	return sdk.Coin{}, fmt.Errorf("error resolving denom")
}
