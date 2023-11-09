package types

import "github.com/skip-mev/feemarket/x/feemarket/interfaces"

// MustNewPlugin creates a new instance of a FeeMarket plugin by marshalling a
// FeeMarketImplementation to bytes.  Will panic() if marshalling fails.
func MustNewPlugin(implementation interfaces.FeeMarketImplementation) []byte {
	implBz, err := implementation.Marshal()
	if err != nil {
		panic(err)
	}

	return implBz
}
