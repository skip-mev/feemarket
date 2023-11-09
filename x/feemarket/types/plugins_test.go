package types_test

import (
	"github.com/skip-mev/feemarket/x/feemarket/plugins/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMustNewPlugin(t *testing.T) {
	t.Run("create valid plugin", func(t *testing.T) {
		require.NotPanics(t, func() {
			MustNewPlugin(mock.NewFeeMarket())
		})
	})
}
