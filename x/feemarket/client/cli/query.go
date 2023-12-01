package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

// GetQueryCmd returns the parent command for all x/feemarket cli query commands.
func GetQueryCmd() *cobra.Command {
	// create base command
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// add sub-commands
	cmd.AddCommand(
		GetParamsCmd(),
		GetStateCmd(),
<<<<<<< HEAD
=======
		GetBaseFeeCmd(),
>>>>>>> main
	)

	return cmd
}

// GetParamsCmd returns the cli-command that queries the current feemarket parameters.
func GetParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
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

			return clientCtx.PrintProto(&resp.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetStateCmd returns the cli-command that queries the current feemarket state.
func GetStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Query for the current feemarket state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			resp, err := queryClient.State(cmd.Context(), &types.StateRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&resp.State)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
<<<<<<< HEAD
=======

// GetBaseFeeCmd returns the cli-command that queries the current feemarket base fee.
func GetBaseFeeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base-fee",
		Short: "Query for the current feemarket base fee",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			resp, err := queryClient.BaseFee(cmd.Context(), &types.BaseFeeRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintString(resp.Fees.String())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
>>>>>>> main
