package types

const (
	// ModuleName is the name of the feemarket module.
	ModuleName = "feemarket"
	// StoreKey is the store key string for the feemarket module.
	StoreKey = ModuleName
)

const (
	prefixParams = iota + 1
	prefixState
)

var (
	// KeyParams is the store key for the feemarket module's parameters.
	KeyParams = []byte{prefixParams}

	// KeyState is the store key for the feemarket module's data.
	KeyState = []byte{prefixState}
)
