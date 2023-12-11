package types

const (
	// ModuleName is the name of the feemarket module.
	ModuleName = "feemarket"
	// StoreKey is the store key string for the feemarket module.
	StoreKey = ModuleName

	// FeeCollectorName the root string for the fee market fee collector account address.
	FeeCollectorName = "feemarket-fee-collector"
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

	EventTypeFeePay      = "fee_pay"
	EventTypeTipPay      = "tip_pay"
	AttributeKeyTip      = "tip"
	AttributeKeyTipPayer = "tip_payer"
	AttributeKeyTipPayee = "tip_payee"
)
