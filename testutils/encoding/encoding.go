package encoding

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/skip-mev/chaintestutil/sample"

	appparams "github.com/skip-mev/feemarket/tests/app/params"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

// MakeTestEncodingConfig creates a test EncodingConfig for a test configuration.
func MakeTestEncodingConfig() appparams.EncodingConfig {
	amino := codec.NewLegacyAmino()

	addFeeMarket := func(ir codectypes.InterfaceRegistry) {
		feemarkettypes.RegisterInterfaces(ir)
	}

	interfaceRegistry := sample.InterfaceRegistry(addFeeMarket)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(cdc, tx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(amino)
	std.RegisterInterfaces(interfaceRegistry)

	return appparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             cdc,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}
