package encoding

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	appparams "github.com/skip-mev/feemarket/tests/app/params"
	"github.com/skip-mev/feemarket/testutils/sample"
)

// MakeTestEncodingConfig creates a test EncodingConfig for a  test configuration.
func MakeTestEncodingConfig() appparams.EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := sample.InterfaceRegistry()
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
