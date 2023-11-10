package types

// NewAIMDEIP1559 instantiates a new AIMD EIP-1559 object.
func NewAIMDEIP1559(
	state State,
	params Params,
) AIMDEIP1559 {
	return AIMDEIP1559{
		State:  state,
		Params: params,
	}
}
