package interfaces

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// FeeMarketImplementation represents the interface of various FeeMarket types implemented
// by other modules or packages.
type FeeMarketImplementation interface {
	proto.Message

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() error

	// Marshal Marshall the feemarket into bytes.
	Marshal() ([]byte, error)

	// Unmarshal the feemarket from bytes.
	Unmarshal([]byte) error

	// ------------------- Fee Market Parameters ------------------- //

	// Init which initializes the fee market (in InitGenesis)
	Init(ctx sdk.Context) error

	// Export which exports the fee market (in ExportGenesis)
	Export(ctx sdk.Context) (json.RawMessage, error)

	// BeginBlockUpdateHandler allows the fee market to be updated
	// after every block. This will be added to the BeginBlock chain.
	BeginBlockUpdateHandler(ctx sdk.Context) UpdateHandler

	// EndBlockUpdateHandler allows the fee market to be updated
	// after every block. This will be added to the EndBlock chain.
	EndBlockUpdateHandler(ctx sdk.Context) UpdateHandler

	// ------------------- Fee Market Queries ------------------- //

	// GetFeeMarketInfo retrieves the fee market's information about
	// how to pay for a transaction (min gas price, min tip,
	// where the fees are being distributed, etc.).
	GetFeeMarketInfo(ctx sdk.Context) map[string]string

	// ------------------- Fee Market Extraction ------------------- //

	// FeeAnteHandler will be called in the module AnteHandler,
	// this is where the fee market would extract and distribute
	// fees from a given transaction
	FeeAnteHandler(
		ctx sdk.Context,
		tx sdk.Tx,
		simulate bool,
		next sdk.AnteHandler,
	) sdk.AnteHandler

	// FeePostHandler will be called in the module PostHandler
	// if PostHandlers are implemented. This is another place
	// the fee market might refund users
	FeePostHandler(
		ctx sdk.Context,
		tx sdk.Tx,
		simulate,
		success bool,
		next sdk.PostHandler,
	) sdk.PostHandler
}

// UpdateHandler is responsible for updating the parameters of the
// fee market plugin. Fees can optionally also be extracted here.
type UpdateHandler func(ctx sdk.Context) error
