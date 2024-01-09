package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrTooManyFeeCoins = sdkerrors.New(ModuleName, 1, "too many fee coins provided.  Only one fee coin may be provided")
	ErrResolverNotSet  = sdkerrors.New(ModuleName, 2, "denom resolver interface not set")
)
