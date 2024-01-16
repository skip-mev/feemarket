package types

import (
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper (noalias)
//
//go:generate mockery --name AccountKeeper --filename mock_account_keeper.go
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI
}

// ConsensusKeeper defines the expected consensus keeper (noalias)
//
//go:generate mockery --name ConsensusKeeper --filename mock_consensus_keeper.go
type ConsensusKeeper interface {
	Get(ctx sdk.Context) (*tmproto.ConsensusParams, error)
}
