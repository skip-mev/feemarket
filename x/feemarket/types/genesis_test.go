package types_test

import (
	"testing"

	"github.com/skip-mev/feemarket/x/feemarket/interfaces"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/feemarket/x/feemarket/plugins/mock"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func TestGenesis(t *testing.T) {
	t.Run("can create a new default genesis state", func(t *testing.T) {
		gs := types.NewDefaultGenesisState()
		require.NoError(t, gs.ValidateBasic())
	})

	t.Run("can accept a valid genesis state with a valid FeeMarket type", func(t *testing.T) {
		plugin := types.MustNewPlugin(mock.NewFeeMarket())
		gs := types.NewGenesisState(plugin, types.NewParams(false))
		require.NoError(t, gs.ValidateBasic())
	})

	t.Run("can reject a genesis with empty implementation", func(t *testing.T) {
		plugin := interfaces.FeeMarket{Implementation: make([]byte, 0)}

		gs := types.NewGenesisState(plugin, types.NewParams(false))
		require.Error(t, gs.ValidateBasic())
	})
}
