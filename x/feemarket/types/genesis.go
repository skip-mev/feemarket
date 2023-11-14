package types

import (
	"encoding/json"
	fmt "fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
)

// NewGenesisState returns a new genesis state for the module.
func NewGenesisState(
	params Params,
	baseFee math.Int,
	learningRate math.LegacyDec,
	utilization BlockUtilization,
) *GenesisState {
	return &GenesisState{
		Params:       params,
		BaseFee:      baseFee,
		LearningRate: learningRate,
		Utilization:  utilization,
	}
}

// ValidateBasic performs basic validation of the genesis state data returning an
// error for any failed validation criteria.
func (gs *GenesisState) ValidateBasic() error {
	if err := gs.Params.ValidateBasic(); err != nil {
		return err
	}

	if err := gs.Utilization.ValidateBasic(); err != nil {
		return err
	}

	if gs.BaseFee.IsNil() || gs.BaseFee.IsNegative() {
		return fmt.Errorf("base fee cannot be nil or negative")
	}

	if gs.LearningRate.IsNil() || gs.LearningRate.IsNegative() {
		return fmt.Errorf("learning rate cannot be nil or negative")
	}

	return nil
}

// GetGenesisStateFromAppState returns x/feemarket GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var gs GenesisState
	cdc.MustUnmarshalJSON(appState[ModuleName], &gs)
	return gs
}
