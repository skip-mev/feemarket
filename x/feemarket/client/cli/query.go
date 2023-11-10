package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

// v returns the parent command for all x/feemarket cli query commands.
func GetQueryCmd() *cobra.Command {
	// create base command
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		RunE:  client.ValidateCmd,
	}

	// add sub-commands
	cmd.AddCommand(
		GetFeeMarketInfo(),
		GetParamsCmd(),
	)

	return cmd
}

// GetFeeMarketInfo returns the cli-command that queries all feemarket state info.
func GetFeeMarketInfo() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Query for feemarket state info in the store",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			resp, err := queryClient.FeeMarketInfo(cmd.Context(), &types.FeeMarketInfoRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}
}

// GetParamsCmd returns the cli-command that queries the current feemarket parameters.
func GetParamsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Query for the current feemarket parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			resp, err := queryClient.Params(cmd.Context(), &types.ParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}
}
