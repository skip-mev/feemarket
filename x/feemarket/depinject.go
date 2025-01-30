package feemarket

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	modulev1 "github.com/skip-mev/feemarket/api/feemarket/feemarket/module/v1"
	"github.com/skip-mev/feemarket/x/feemarket/keeper"
	"github.com/skip-mev/feemarket/x/feemarket/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	govtypes "cosmossdk.io/x/gov/types"
)

func init() {
	appconfig.Register(
		&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type Inputs struct {
	depinject.In

	Environment   appmodule.Environment
	Config        *modulev1.Module
	Cdc           codec.Codec
	AccountKeeper types.AccountKeeper
}

type Outputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in Inputs) Outputs {
	var (
		authority sdk.AccAddress
		err       error
	)
	if in.Config.Authority != "" {
		authority, err = sdk.AccAddressFromBech32(in.Config.Authority)
		if err != nil {
			panic(err)
		}
	} else {
		authority = authtypes.NewModuleAddress(govtypes.ModuleName)
	}

	Keeper := keeper.NewKeeper(
		in.Cdc,
		in.Environment,
		in.AccountKeeper,
		nil,
		authority.String(),
	)

	m := NewAppModule(in.Cdc, *Keeper)

	return Outputs{Keeper: *Keeper, Module: m}
}
