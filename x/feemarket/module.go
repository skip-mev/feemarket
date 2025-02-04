package feemarket

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"

	"github.com/skip-mev/feemarket/x/feemarket/client/cli"
	"github.com/skip-mev/feemarket/x/feemarket/keeper"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

// ConsensusVersion is the x/feemarket module's consensus version identifier.
const ConsensusVersion = 1

var (
	_ module.HasGenesis          = AppModule{}
	_ module.HasGRPCGateway      = AppModule{}
	_ module.HasRegisterServices = AppModule{}

	_ appmodule.AppModule             = AppModule{}
	_ appmodule.HasBeginBlocker       = AppModule{}
	_ appmodule.HasEndBlocker         = AppModule{}
	_ appmodule.HasAminoCodec         = AppModule{}
	_ appmodule.HasRegisterInterfaces = AppModule{}
)

// Name returns the name of x/feemarket module.
func (am AppModule) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the necessary types from the x/feemarket module for amino
// serialization.
func (am AppModule) RegisterLegacyAminoCodec(cdc registry.AminoRegistrar) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the necessary implementations / interfaces in the x/feemarket
// module w/ the interface-registry.
func (am AppModule) RegisterInterfaces(ir registry.InterfaceRegistrar) {
	types.RegisterInterfaces(ir)
}

// RegisterGRPCGatewayRoutes registers the necessary REST routes for the GRPC-gateway to
// the x/feemarket module QueryService on mux. This method panics on failure.
func (am AppModule) RegisterGRPCGatewayRoutes(cliCtx client.Context, mux *runtime.ServeMux) {
	// Register the gate-way routes w/ the provided mux.
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(cliCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd is a no-op, as no txs are registered for submission (apart from messages that
// can only be executed by governance).
func (am AppModule) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns the x/feemarket module base query cli-command.
func (am AppModule) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule represents an application module for the x/feemarket module.
type AppModule struct {
	cdc codec.Codec

	k keeper.Keeper
}

func (am AppModule) BeginBlock(_ context.Context) error {
	return nil
}

// NewAppModule returns an application module for the x/feemarket module.
func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		cdc: cdc,
		k:   k,
	}
}

// EndBlock returns an endblocker for the x/feemarket module.
func (am AppModule) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return am.k.EndBlock(sdkCtx)
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

func (am AppModule) RegisterServices(r grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(r, keeper.NewMsgServer(&am.k))
	types.RegisterQueryServer(r, keeper.NewQueryServer(am.k))
	return nil
}

// DefaultGenesis returns default genesis state as raw bytes for the feemarket
// module.
func (am AppModule) DefaultGenesis() json.RawMessage {
	return am.cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the feemarket module.
func (am AppModule) ValidateGenesis(data json.RawMessage) error {
	var gs types.GenesisState
	if err := am.cdc.UnmarshalJSON(data, &gs); err != nil {
		return err
	}

	return gs.ValidateBasic()
}

// InitGenesis performs the genesis initialization for the x/feemarket module. This method returns
// no validator set updates. This method panics on any errors.
func (am AppModule) InitGenesis(ctx context.Context, data json.RawMessage) error {
	var gs types.GenesisState
	if err := am.cdc.UnmarshalJSON(data, &gs); err != nil {
		return err
	}
	am.k.InitGenesis(ctx, gs)
	return nil
}

// ExportGenesis returns the feemarket module's exported genesis state as raw
// JSON bytes. This method panics on any error.
func (am AppModule) ExportGenesis(ctx context.Context) (json.RawMessage, error) {
	gs := am.k.ExportGenesis(ctx)
	return am.cdc.MarshalJSON(gs)
}
