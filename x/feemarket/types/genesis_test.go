package types_test

import (
	"github.com/skip-mev/feemarket/x/feemarket/plugins/mock"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func TestGenesis(t *testing.T) {
	t.Run("can create a new default genesis state", func(t *testing.T) {
		gs := types.NewDefaultGenesisState()
		require.NoError(t, gs.ValidateBasic())
	})

	t.Run("can accept a valid genesis state with a valid FeeMarket type", func(t *testing.T) {
		plugin := types.MustNewPlugin(mock.NewFeeMarket())
		gs := types.NewGenesisState(plugin, types.Params{Enabled: false})
		require.NoError(t, gs.ValidateBasic())
	})

	t.Run("can accept a valid genesis state with multiple incentive types", func(t *testing.T) {
		badPrice := types.NewIncentives("badprice", [][]byte{[]byte("incentive1")})
		goodPrice := types.NewIncentives("goodprice", [][]byte{[]byte("incentive1")})

		gs := types.NewGenesisState([]types.IncentivesByType{badPrice, goodPrice})

		require.NoError(t, gs.ValidateBasic())
	})

	t.Run("can reject a genesis state with duplicate incentive types", func(t *testing.T) {
		badPrice := types.NewIncentives("badprice", [][]byte{[]byte("incentive1")})
		goodPrice := types.NewIncentives("badprice", [][]byte{[]byte("incentive1")})

		gs := types.NewGenesisState([]types.IncentivesByType{badPrice, goodPrice})

		require.Error(t, gs.ValidateBasic())
	})

	t.Run("can reject a genesis state with an empty incentive type", func(t *testing.T) {
		badPrice := types.NewIncentives("", [][]byte{[]byte("incentive1")})

		gs := types.NewGenesisState([]types.IncentivesByType{badPrice})

		require.Error(t, gs.ValidateBasic())
	})

	t.Run("can reject a genesis state with an empty incentive", func(t *testing.T) {
		badPrice := types.NewIncentives("badprice", [][]byte{[]byte("")})

		gs := types.NewGenesisState([]types.IncentivesByType{badPrice})

		require.Error(t, gs.ValidateBasic())
	})
}
