package types

const (
	// ModuleName is the name of the feemarket module.
	ModuleName = "feemarket"
	// StoreKey is the store key string for the feemarket module.
	StoreKey = ModuleName
)

const (
	prefixParams = iota + 1
)

// KeyParams is the store key for the feemarket module's parameters.
var KeyParams = []byte{prefixParams}
