package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

var _ types.QueryServer = (*QueryServer)(nil)

// QueryServer defines the gRPC server for the x/sla module.
type QueryServer struct {
	k Keeper
}

// Params defines a method that returns the current feemarket parameters.
func (q QueryServer) Params(goCtx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := q.k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &types.ParamsResponse{Params: params}, nil
}

// FeeMarketInfo defines a method that returns the current feemarket state info.
func (q QueryServer) FeeMarketInfo(goCtx context.Context, _ *types.FeeMarketInfoRequest) (*types.FeeMarketInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	info := q.k.GetFeeMarketInfo(ctx)

	return &types.FeeMarketInfoResponse{Info: info}, nil
}
