package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DenomResolver is an interface to convert a given token to the feemarket's base token.
type DenomResolver interface {
	// ConvertToDenom converts coin into the equivalent amount of the token denominated in denom.
	ConvertToDenom(ctx sdk.Context, coin sdk.Coin, denom string) (sdk.Coin, error)
}

// TestDenomResolver is a test implementation of the DenomResolver interface.  It returns "feeCoin.Amount baseDenom" for all coins that are not the baseDenom.
// NOTE: DO NOT USE THIS IN PRODUCTION
type TestDenomResolver struct{}

// ConvertToDenom returns "coin.Amount denom" for all coins that are not the denom.
func (r *TestDenomResolver) ConvertToDenom(_ sdk.Context, coin sdk.Coin, denom string) (sdk.Coin, error) {
	if coin.Denom == denom {
		return coin, nil
	}

	return sdk.NewCoin(denom, coin.Amount), nil
}

// ErrorDenomResolver is a test implementation of the DenomResolver interface.  It returns an error for all coins that are not the baseDenom.
// NOTE: DO NOT USE THIS IN PRODUCTION
type ErrorDenomResolver struct{}

// ConvertToDenom returns an error for all coins that are not the denom.
func (r *ErrorDenomResolver) ConvertToDenom(_ sdk.Context, coin sdk.Coin, denom string) (sdk.Coin, error) {
	if coin.Denom == denom {
		return coin, nil
	}

	return sdk.Coin{}, fmt.Errorf("error resolving denom")
}
