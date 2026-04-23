package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
)

// MintToAccount mints the specified coins into the account balance.
func (tk *TestKeepers) MintToAccount(ctx sdk.Context, address string, coins sdk.Coins) {
	sdkAddr, err := sdk.AccAddressFromBech32(address)
	require.NoError(tk.T, err)
	require.NoError(tk.T, tk.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(tk.T, tk.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, sdkAddr, coins))
}

// MintToModule mints the specified coins into the module account balance.
func (tk *TestKeepers) MintToModule(ctx sdk.Context, moduleAcc string, coins sdk.Coins) {
	require.NoError(tk.T, tk.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(tk.T, tk.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, moduleAcc, coins))
}
