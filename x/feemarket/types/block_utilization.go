package types

import (
	fmt "fmt"
)

// NewBlockUtilization instantiates a new block utilization instance. This
// struct is utilized to track how full blocks are over a sliding
// window.
func NewBlockUtilization(window uint64) BlockUtilization {
	return BlockUtilization{
		Window: make([]uint64, window),
		Index:  0,
	}
}

// ValidateBasic performs basic validation on the state.
func (s *BlockUtilization) ValidateBasic() error {
	if s.Window == nil || len(s.Window) == 0 {
		return fmt.Errorf("block utilization window cannot be nil or empty")
	}

	return nil
}
