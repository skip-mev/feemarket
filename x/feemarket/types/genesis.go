package types

import (
	"encoding/json"

	"github.com/skip-mev/feemarket/x/feemarket/interfaces"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/skip-mev/feemarket/x/feemarket/plugins/mock"
)

// NewDefaultGenesisState returns a default genesis state for the module.
func NewDefaultGenesisState() *GenesisState {
	return &GenesisState{
		Plugin: MustNewPlugin(mock.NewFeeMarket()), // TODO replace with another impl
		Params: DefaultParams(),
	}
}

// NewGenesisState returns a new genesis state for the module.
func NewGenesisState(plugin interfaces.FeeMarket, params Params) *GenesisState {
	return &GenesisState{
		Plugin: plugin,
		Params: params,
	}
}

// ValidateBasic performs basic validation of the genesis state data returning an
// error for any failed validation criteria.
func (gs *GenesisState) ValidateBasic() error {
	if err := gs.Plugin.ValidateBasic(); err != nil {
		return err
	}

	return gs.Params.ValidateBasic()
}

// GetGenesisStateFromAppState returns x/sla GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var gs GenesisState
	cdc.MustUnmarshalJSON(appState[ModuleName], &gs)
	return gs
}
