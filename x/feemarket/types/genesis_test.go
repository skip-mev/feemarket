package types_test

import (
	"testing"

	"github.com/skip-mev/feemarket/x/feemarket/plugins/defaultmarket"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func TestGenesis(t *testing.T) {
	t.Run("can create a new default genesis state", func(t *testing.T) {
		gs := types.NewDefaultGenesisState()
		require.NoError(t, gs.ValidateBasic())
	})

	t.Run("can accept a valid genesis state with a valid FeeMarket type", func(t *testing.T) {
		gs := types.NewGenesisState(defaultmarket.NewDefaultFeeMarket(), types.NewParams(false))
		require.NoError(t, gs.ValidateBasic())
	})
}
