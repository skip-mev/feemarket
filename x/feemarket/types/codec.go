package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	"github.com/skip-mev/feemarket/x/feemarket/interfaces"
	"github.com/skip-mev/feemarket/x/feemarket/plugins/defaultmarket"
)

// RegisterLegacyAminoCodec registers the necessary x/feemarket interfaces (messages) on the
// provided LegacyAmino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgParams{}, "feemarket/MsgParams")

	cdc.RegisterInterface((*interfaces.FeeMarketImplementation)(nil), nil)
	cdc.RegisterConcrete(&defaultmarket.DefaultMarket{}, "feemarket/DefaultMarket", nil)
}

// RegisterInterfaces registers the x/feemarket interfaces (messages + msg server) on the
// provided InterfaceRegistry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgParams{},
	)

	registry.RegisterInterface(
		"feemarket.feemarket.v1.FeeMarketImplementation",
		(*interfaces.FeeMarketImplementation)(nil),
		&defaultmarket.DefaultMarket{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
