package feemarket

import sdk "github.com/cosmos/cosmos-sdk/types"

type (
	// FeeMarket defines the expected interface each fee market plugin must
	// implement to be utilized by the fee market module.
	FeeMarket interface {
		// ------------------- Fee Market Parameters ------------------- //

		// Init which initializes the fee market (in InitGenesis)
		Init(ctx sdk.Context) error

		// EndBlockUpdateHandler allows the fee market to be updated
		// after every block. This will be added to the EndBlock chain.
		EndBlockUpdateHandler(ctx sdk.Context) UpdateHandler

		// EpochUpdateHandler allows the fee market to be updated
		// after every given epoch identifier. This maps the epoch
		// identifier to the UpdateHandler that should be executed.
		EpochUpdateHandler(ctx sdk.Context) map[string]UpdateHandler

		// ------------------- Fee Market Queries ------------------- //

		// GetMinGasPrice retrieves the minimum gas price(s) needed
		// to be included in the block for the given transaction
		GetMinGasPrice(ctx sdk.Context, tx sdk.Tx) sdk.Coins

		// GetFeeMarketInfo retrieves the fee market's information about
		// how to pay for a transaction (min gas price, min tip,
		// where the fees are being distributed, etc.).
		GetFeeMarketInfo(ctx sdk.Context) map[string]string

		// GetID returns the identifier of the fee market
		GetID() string

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
	UpdateHandler func(ctx sdk.Context) error
)
