package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

var _ types.MsgServer = (*MsgServer)(nil)

// MsgServer is the server API for x/feemarket Msg service.
type MsgServer struct {
	k Keeper
}

// NewMsgServer returns the MsgServer implementation.
func NewMsgServer(k Keeper) types.MsgServer {
	return &MsgServer{k}
}

// Params defines a method that updates the module's parameters. The signer of the message must
// be the module authority.
func (ms MsgServer) Params(goCtx context.Context, msg *types.MsgParams) (*types.MsgParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != ms.k.GetAuthority() {
		return nil, fmt.Errorf("invalid authority to execute message")
	}

	maxGas, err := ms.k.GetMaxGasUtilization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get consensus params: %w", err)
	}

	maxUtilization := uint64(maxGas)
	if msg.Params.TargetBlockUtilization > maxUtilization {
		return nil, fmt.Errorf("target block size of %d cannot be greater than max block size of %d", msg.Params.TargetBlockUtilization, maxUtilization)
	}

	if maxUtilization/msg.Params.TargetBlockUtilization > types.MaxBlockUtilizationRatio {
		return nil, fmt.Errorf(fmt.Sprintf("max block size cannot be greater than target block size times %d", types.MaxBlockUtilizationRatio))
	}

	params := msg.Params
	if err := ms.k.SetParams(ctx, params); err != nil {
		return nil, fmt.Errorf("error setting params: %w", err)
	}

	newState := types.NewState(params.WindowSize, params.MinBaseFee, params.MinLearningRate)
	if err := ms.k.SetState(ctx, newState); err != nil {
		return nil, fmt.Errorf("error setting state: %w", err)
	}

	return &types.MsgParamsResponse{}, nil
}
