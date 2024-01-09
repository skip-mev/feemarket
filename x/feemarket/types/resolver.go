package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// DenomResolver is an interface to convert a given token to the feemarket's base token.
type DenomResolver interface {
	// ConvertToBaseToken converts feeCoin into the equivalent amount of the token denominated in baseDenom.
	ConvertToBaseToken(ctx sdk.Context, feeCoin sdk.Coin, baseDenom string)
}
