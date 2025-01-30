package keeper

import (
	"context"
	"testing"

	"github.com/skip-mev/feemarket/x/feemarket"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
	"gotest.tools/v3/assert"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/accounts"
	"cosmossdk.io/x/accounts/accountstd"
	baseaccount "cosmossdk.io/x/accounts/defaults/base"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/consensus"
	consensuskeeper "cosmossdk.io/x/consensus/keeper"
	consensustypes "cosmossdk.io/x/consensus/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	minttypes "cosmossdk.io/x/mint/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type fixture struct {
	app *integration.App

	cdc codec.Codec

	authKeeper      authkeeper.AccountKeeper
	accountsKeeper  accounts.Keeper
	bankKeeper      bankkeeper.Keeper
	feemarketKeeper *feemarketkeeper.Keeper
	consensusKeeper consensuskeeper.Keeper
	feegrantKeeper  feegrantkeeper.Keeper
}

func (f fixture) mustAddr(address []byte) string {
	s, _ := f.authKeeper.AddressCodec().BytesToString(address)
	return s
}

func initFixture(t *testing.T, extraAccs map[string]accountstd.Interface) *fixture {
	t.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		accounts.StoreKey,
		feemarkettypes.StoreKey,
		feegrant.StoreKey,
	)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(
		codectestutil.CodecOptions{},
		auth.AppModule{},
		bank.AppModule{},
		accounts.AppModule{},
		feemarket.AppModule{},
		feegrantmodule.AppModule{},
	)
	cdc := encodingCfg.Codec

	logger := log.NewTestLogger(t)

	router := baseapp.NewMsgServiceRouter()
	queryRouter := baseapp.NewGRPCQueryRouter()

	handler := directHandler{}
	account := baseaccount.NewAccount("base", signing.NewHandlerMap(handler), baseaccount.WithSecp256K1PubKey())

	var accs []accountstd.AccountCreatorFunc
	for name, acc := range extraAccs {
		f := accountstd.AddAccount(name, func(_ accountstd.Dependencies) (accountstd.Interface, error) {
			return acc, nil
		})
		accs = append(accs, f)
	}

	accountsKeeper, err := accounts.NewKeeper(
		cdc,
		runtime.NewEnvironment(
			runtime.NewKVStoreService(keys[accounts.StoreKey]), log.NewNopLogger(), runtime.EnvWithQueryRouterService(queryRouter), runtime.EnvWithMsgRouterService(router)),
		addresscodec.NewBech32Codec("cosmos"),
		cdc.InterfaceRegistry(),
		nil,
		append(accs, account)...,
	)
	assert.NilError(t, err)
	accountsv1.RegisterQueryServer(queryRouter, accounts.NewQueryServer(accountsKeeper))

	authority := authtypes.NewModuleAddress("gov")

	authKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
		cdc,
		authtypes.ProtoBaseAccount,
		accountsKeeper,
		map[string][]string{
			minttypes.ModuleName:            {authtypes.Minter},
			feemarkettypes.ModuleName:       nil,
			feemarkettypes.FeeCollectorName: {authtypes.Burner},
		},
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	fgKeeper := feegrantkeeper.NewKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[accounts.StoreKey]), log.NewNopLogger()),
		cdc,
		authKeeper,
	)

	blockedAddresses := map[string]bool{
		authKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[banktypes.StoreKey]), log.NewNopLogger()),
		cdc,
		authKeeper,
		blockedAddresses,
		authority.String(),
	)

	feemarketKeeper := feemarketkeeper.NewKeeper(
		cdc,
		keys[feemarkettypes.StoreKey],
		authKeeper,
		&feemarkettypes.TestDenomResolver{},
		authority.String(),
	)

	consensusKeeper := consensuskeeper.NewKeeper(
		cdc,
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[consensustypes.StoreKey]), log.NewNopLogger()),
		authority.String(),
	)

	accountsModule := accounts.NewAppModule(cdc, accountsKeeper)
	authModule := auth.NewAppModule(cdc, authKeeper, accountsKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, authKeeper)
	feemarketModule := feemarket.NewAppModule(cdc, *feemarketKeeper)
	consensusModule := consensus.NewAppModule(cdc, consensusKeeper)
	fgModule := feegrantmodule.NewAppModule(cdc, fgKeeper, cdc.InterfaceRegistry())

	integrationApp := integration.NewIntegrationApp(
		logger,
		keys,
		cdc,
		encodingCfg.InterfaceRegistry.SigningContext().AddressCodec(),
		encodingCfg.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
		map[string]appmodule.AppModule{
			accounts.ModuleName:       accountsModule,
			authtypes.ModuleName:      authModule,
			banktypes.ModuleName:      bankModule,
			feemarkettypes.ModuleName: feemarketModule,
			consensustypes.ModuleName: consensusModule,
			feegrant.ModuleName:       fgModule,
		},
		router,
		queryRouter,
	)

	authtypes.RegisterInterfaces(cdc.InterfaceRegistry())
	banktypes.RegisterInterfaces(cdc.InterfaceRegistry())
	feemarkettypes.RegisterInterfaces(cdc.InterfaceRegistry())
	consensustypes.RegisterInterfaces(cdc.InterfaceRegistry())

	authtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), authkeeper.NewMsgServerImpl(authKeeper))
	authtypes.RegisterQueryServer(integrationApp.QueryHelper(), authkeeper.NewQueryServer(authKeeper))

	banktypes.RegisterMsgServer(router, bankkeeper.NewMsgServerImpl(bankKeeper))

	feemarkettypes.RegisterMsgServer(router, feemarketkeeper.NewMsgServer(feemarketKeeper))
	feemarkettypes.RegisterQueryServer(router, feemarketkeeper.NewQueryServer(*feemarketKeeper))

	consensustypes.RegisterQueryServer(router, consensusKeeper)
	consensustypes.RegisterMsgServer(router, consensusKeeper)

	feegrant.RegisterQueryServer(router, fgKeeper)
	feegrant.RegisterMsgServer(router, feegrantkeeper.NewMsgServerImpl(fgKeeper))

	return &fixture{
		app:             integrationApp,
		cdc:             cdc,
		accountsKeeper:  accountsKeeper,
		authKeeper:      authKeeper,
		bankKeeper:      bankKeeper,
		feemarketKeeper: feemarketKeeper,
		consensusKeeper: consensusKeeper,
		feegrantKeeper:  fgKeeper,
	}
}

type directHandler struct{}

func (s directHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT
}

func (s directHandler) GetSignBytes(_ context.Context, _ signing.SignerData, _ signing.TxData) ([]byte, error) {
	panic("not implemented")
}
