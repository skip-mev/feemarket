package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

var _ types.QueryServer = (*QueryServer)(nil)

// QueryServer defines the gRPC server for the x/feemarket module.
type QueryServer struct {
	k Keeper
}

// NewQueryServer creates a new instance of the x/feemarket QueryServer type.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return &QueryServer{k: keeper}
}

// Params defines a method that returns the current feemarket parameters.
func (q QueryServer) Params(goCtx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := q.k.GetParams(ctx)
	return &types.ParamsResponse{Params: params}, err
}

// State defines a method that returns the current feemarket state.
func (q QueryServer) State(goCtx context.Context, _ *types.StateRequest) (*types.StateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	state, err := q.k.GetState(ctx)
	return &types.StateResponse{State: state}, err
}

// BaseFee defines a method that returns the current feemarket base fee.
func (q QueryServer) BaseFee(goCtx context.Context, _ *types.BaseFeeRequest) (*types.BaseFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	fees, err := q.k.GetMinGasPrices(ctx)
	return &types.BaseFeeResponse{Fees: fees}, err
}
