package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

// GetQueryCmd returns the parent command for all x/feemarket cli query commands.
func GetQueryCmd() *cobra.Command {
	// create base command
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		RunE:  client.ValidateCmd,
	}

	// add sub-commands
	cmd.AddCommand(
		GetParamsCmd(),
	)

	return cmd
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
