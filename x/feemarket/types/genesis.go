package types

import (
	"encoding/json"

	"github.com/skip-mev/feemarket/x/feemarket/plugins/defaultmarket"

	"github.com/skip-mev/feemarket/x/feemarket/interfaces"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NewDefaultGenesisState returns a default genesis state for the module.
func NewDefaultGenesisState() *GenesisState {
	return &GenesisState{
		Plugin: MustNewPlugin(defaultmarket.NewDefaultFeeMarket()), // TODO replace with another impl
		Params: DefaultParams(),
	}
}

// NewGenesisState returns a new genesis state for the module.  Panics if it cannot marshal plugin.
func NewGenesisState(plugin interfaces.FeeMarketImplementation, params Params) *GenesisState {
	bz := MustNewPlugin(plugin)

	return &GenesisState{
		Plugin: bz,
		Params: params,
	}
}

// ValidateBasic performs basic validation of the genesis state data returning an
// error for any failed validation criteria.
func (gs *GenesisState) ValidateBasic() error {
	return gs.Params.ValidateBasic()
}

// GetGenesisStateFromAppState returns x/feemarket GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var gs GenesisState
	cdc.MustUnmarshalJSON(appState[ModuleName], &gs)
	return gs
}
