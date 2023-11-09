package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/feemarket/x/feemarket/plugins/defaultmarket"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

func TestMustNewPlugin(t *testing.T) {
	t.Run("create valid plugin", func(t *testing.T) {
		require.NotPanics(t, func() {
			types.MustNewPlugin(defaultmarket.NewDefaultFeeMarket())
		})
	})
}
