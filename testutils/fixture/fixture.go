package fixture

import (
	"context"
	"math"
	"testing"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/skip-mev/feemarket/x/feemarket"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/accounts"
	baseaccount "cosmossdk.io/x/accounts/defaults/base"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/consensus"
	consensuskeeper "cosmossdk.io/x/consensus/keeper"
	consensustypes "cosmossdk.io/x/consensus/types"
	distrtypes "cosmossdk.io/x/distribution/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	minttypes "cosmossdk.io/x/mint/types"
	stakingtypes "cosmossdk.io/x/staking/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type TestFixture struct {
	Ctx sdk.Context

	Cdc codec.Codec

	EncodingConfig moduletestutil.TestEncodingConfig

	AuthKeeper      authkeeper.AccountKeeper
	AccountsKeeper  accounts.Keeper
	BankKeeper      bankkeeper.Keeper
	FeeMarketKeeper *feemarketkeeper.Keeper
	ConsensusKeeper consensuskeeper.Keeper
	FeeGrantKeeper  feegrantkeeper.Keeper
}

func setupDBs() (store.CommitMultiStore, *dbm.MemDB) {
	logger := log.NewNopLogger()
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())

	// Create a context using a custom timestamp
	return stateStore, db
}

func NewTestFixture(t *testing.T) *TestFixture {
	cms, db := setupDBs()
	t.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		accounts.StoreKey,
		feemarkettypes.StoreKey,
		feegrant.StoreKey,
		consensustypes.StoreKey,
	)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(
		codectestutil.CodecOptions{},
		auth.AppModule{},
		bank.AppModule{},
		accounts.AppModule{},
		feemarket.AppModule{},
		feegrantmodule.AppModule{},
		consensus.AppModule{},
	)
	cdc := encodingCfg.Codec

	queryRouter := baseapp.NewGRPCQueryRouter()

	handler := directHandler{}
	account := baseaccount.NewAccount("base", signing.NewHandlerMap(handler), baseaccount.WithSecp256K1PubKey())

	accKey := keys[accounts.StoreKey]
	cms.MountStoreWithDB(accKey, storetypes.StoreTypeIAVL, db)
	accStore := runtime.NewKVStoreService(accKey)
	accountsKeeper, err := accounts.NewKeeper(
		cdc,
		runtime.NewEnvironment(accStore, log.NewNopLogger()),
		addresscodec.NewBech32Codec("cosmos"),
		cdc.InterfaceRegistry(),
		nil,
		account,
	)
	assert.NilError(t, err)
	accountsv1.RegisterQueryServer(queryRouter, accounts.NewQueryServer(accountsKeeper))

	authority := authtypes.NewModuleAddress("gov")

	authKey := keys[authtypes.StoreKey]
	cms.MountStoreWithDB(authKey, storetypes.StoreTypeIAVL, db)
	authKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(authKey), log.NewNopLogger()),
		cdc,
		authtypes.ProtoBaseAccount,
		accountsKeeper,
		map[string][]string{
			authtypes.FeeCollectorName:      nil,
			distrtypes.ModuleName:           nil,
			minttypes.ModuleName:            {authtypes.Minter},
			stakingtypes.BondedPoolName:     {authtypes.Burner, authtypes.Staking},
			stakingtypes.NotBondedPoolName:  {authtypes.Burner, authtypes.Staking},
			feemarkettypes.ModuleName:       nil,
			feemarkettypes.FeeCollectorName: {authtypes.Burner},
		},
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	fgKey := keys[feegrant.StoreKey]
	cms.MountStoreWithDB(fgKey, storetypes.StoreTypeIAVL, db)
	fgKeeper := feegrantkeeper.NewKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(fgKey), log.NewNopLogger()),
		cdc,
		authKeeper,
	)

	blockedAddresses := map[string]bool{
		authKeeper.GetAuthority(): false,
	}
	bankKey := keys[banktypes.StoreKey]
	cms.MountStoreWithDB(bankKey, storetypes.StoreTypeIAVL, db)
	bankKeeper := bankkeeper.NewBaseKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(bankKey), log.NewNopLogger()),
		cdc,
		authKeeper,
		blockedAddresses,
		authority.String(),
	)

	fmKey := keys[feemarkettypes.StoreKey]
	cms.MountStoreWithDB(fmKey, storetypes.StoreTypeIAVL, db)
	feemarketKeeper := feemarketkeeper.NewKeeper(
		cdc,
		runtime.NewEnvironment(runtime.NewKVStoreService(fmKey), log.NewNopLogger()),
		authKeeper,
		&feemarkettypes.TestDenomResolver{},
		authority.String(),
	)

	consensusKey := keys[consensustypes.StoreKey]
	cms.MountStoreWithDB(consensusKey, storetypes.StoreTypeIAVL, db)
	consensusKeeper := consensuskeeper.NewKeeper(
		cdc,
		runtime.NewEnvironment(runtime.NewKVStoreService(consensusKey), log.NewNopLogger()),
		authority.String(),
	)

	require.NoError(t, cms.LoadLatestVersion())
	ctx := sdk.NewContext(cms, false, log.NewNopLogger())

	err = feemarketKeeper.SetState(ctx, feemarkettypes.DefaultState())
	require.NoError(t, err)
	err = feemarketKeeper.SetParams(ctx, feemarkettypes.DefaultParams())
	require.NoError(t, err)

	err = consensusKeeper.ParamsStore.Set(ctx, cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: math.MaxInt64, MaxBytes: math.MaxInt64}})
	require.NoError(t, err)

	authtypes.RegisterInterfaces(cdc.InterfaceRegistry())
	banktypes.RegisterInterfaces(cdc.InterfaceRegistry())
	feemarkettypes.RegisterInterfaces(cdc.InterfaceRegistry())
	consensustypes.RegisterInterfaces(cdc.InterfaceRegistry())
	feegrant.RegisterInterfaces(cdc.InterfaceRegistry())

	return &TestFixture{
		Ctx:             ctx,
		Cdc:             cdc,
		EncodingConfig:  encodingCfg,
		AuthKeeper:      authKeeper,
		AccountsKeeper:  accountsKeeper,
		BankKeeper:      bankKeeper,
		FeeMarketKeeper: feemarketKeeper,
		ConsensusKeeper: consensusKeeper,
		FeeGrantKeeper:  fgKeeper,
	}
}

var _ signing.SignModeHandler = &directHandler{}

type directHandler struct{}

func (s directHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT
}

func (s directHandler) GetSignBytes(_ context.Context, _ signing.SignerData, _ signing.TxData) ([]byte, error) {
	panic("not implemented")
}
