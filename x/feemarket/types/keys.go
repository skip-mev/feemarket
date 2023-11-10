package types

const (
	// ModuleName is the name of the feemarket module.
	ModuleName = "feemarket"
	// StoreKey is the store key string for the feemarket module.
	StoreKey = ModuleName
)

const (
	prefixParams = iota + 1
	prefixData
)

var (
	// KeyParams is the store key for the feemarket module's parameters.
	KeyParams = []byte{prefixParams}

	// KeyData is the store key for the feemarket module's data.
	KeyData = []byte{prefixData}
)
