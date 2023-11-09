package types

// MustNewPlugin creates a new instance of a FeeMarket plugin by marshalling a
// FeeMarketImplementation to bytes.  Will panic() if marshalling fails.
func MustNewPlugin(implementation FeeMarketImplementation) FeeMarket {
	implBz, err := implementation.Marshal()
	if err != nil {
		panic(err)
	}

	return FeeMarket{Implementation: implBz}
}
