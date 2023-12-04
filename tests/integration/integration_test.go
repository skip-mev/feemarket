package integration_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testkeeper "github.com/skip-mev/feemarket/testutils/keeper"
)

func TestKeepers(t *testing.T) {
	var (
		ctx, tk, _ = testkeeper.NewTestSetup(t)
		wCtx       = sdk.WrapSDKContext(ctx)
	)

	numIterations := 100

	for i := 0; i < numIterations; i++ {
		_ = tk
		_ = wCtx
	}
}
