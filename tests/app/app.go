package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	corestore "cosmossdk.io/core/store"
	lockup "cosmossdk.io/x/accounts/defaults/lockup"
	"cosmossdk.io/x/evidence"

	"cosmossdk.io/x/protocolpool"
	poolkeeper "cosmossdk.io/x/protocolpool/keeper"
	pooltypes "cosmossdk.io/x/protocolpool/types"

	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"

	"cosmossdk.io/x/accounts/accountstd"
	baseaccount "cosmossdk.io/x/accounts/defaults/base"
	txdecode "cosmossdk.io/x/tx/decode"

	accounts "cosmossdk.io/x/accounts"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	testdata_pulsar "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cast"

	"cosmossdk.io/x/accounts/defaults/multisig"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/authz"
	authzkeeper "cosmossdk.io/x/authz/keeper"
	authzmodule "cosmossdk.io/x/authz/module"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/circuit"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	circuittypes "cosmossdk.io/x/circuit/types"
	consensus "cosmossdk.io/x/consensus"
	consensusparamkeeper "cosmossdk.io/x/consensus/keeper"
	consensusparamtypes "cosmossdk.io/x/consensus/types"
	distr "cosmossdk.io/x/distribution"
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	distrtypes "cosmossdk.io/x/distribution/types"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	govkeeper "cosmossdk.io/x/gov/keeper"
	govtypes "cosmossdk.io/x/gov/types"
	govv1beta1 "cosmossdk.io/x/gov/types/v1beta1"
	"cosmossdk.io/x/group"
	groupkeeper "cosmossdk.io/x/group/keeper"
	groupmodule "cosmossdk.io/x/group/module"
	"cosmossdk.io/x/mint"
	mintkeeper "cosmossdk.io/x/mint/keeper"
	minttypes "cosmossdk.io/x/mint/types"
	"cosmossdk.io/x/nft"
	"cosmossdk.io/x/slashing"
	slashingkeeper "cosmossdk.io/x/slashing/keeper"
	slashingtypes "cosmossdk.io/x/slashing/types"
	"cosmossdk.io/x/staking"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/skip-mev/feemarket/x/feemarket"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

const appName = "SimApp"

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:         nil,
		distrtypes.ModuleName:              nil,
		pooltypes.ModuleName:               nil,
		pooltypes.StreamAccount:            nil,
		pooltypes.ProtocolPoolDistrAccount: nil,
		minttypes.ModuleName:               {authtypes.Minter},
		stakingtypes.BondedPoolName:        {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:     {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                {authtypes.Burner},
		feemarkettypes.ModuleName:          {authtypes.Burner},
		feemarkettypes.FeeCollectorName:    {authtypes.Burner},
	}
)

var (
	_ runtime.AppI            = (*SimApp)(nil)
	_ servertypes.Application = (*SimApp)(nil)
)

// SimApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type SimApp struct {
	*baseapp.BaseApp
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// keys to access the substores
	keys map[string]*storetypes.KVStoreKey

	// keepers
	AuthKeeper            authkeeper.AccountKeeper
	AccountKeeper         accounts.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	BankKeeper            bankkeeper.BaseKeeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            *mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	CircuitKeeper         circuitkeeper.Keeper
	PoolKeeper            poolkeeper.Keeper
	FeeMarketKeeper       *feemarketkeeper.Keeper

	// the module manager
	ModuleManager *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".simapp")
}

// NewSimApp returns a reference to an initialized SimApp.
func NewSimApp(
	logger log.Logger,
	db corestore.KVStoreWithBatch,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	signingCtx := interfaceRegistry.SigningContext()
	txDecoder, err := txdecode.NewDecoder(txdecode.Options{
		SigningContext: signingCtx,
		ProtoCodec:     appCodec,
	})
	if err != nil {
		panic(err)
	}
	txConfig := authtx.NewTxConfig(appCodec, signingCtx.AddressCodec(), signingCtx.ValidatorAddressCodec(), authtx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	// Below we could construct and set an application specific mempool and
	// ABCI 1.0 PrepareProposal and ProcessProposal handlers. These defaults are
	// already set in the SDK's BaseApp, this shows an example of how to override
	// them.
	//
	// Example:
	//
	// bApp := baseapp.NewBaseApp(...)
	// nonceMempool := mempool.NewSenderNonceMempool()
	// abciPropHandler := NewDefaultProposalHandler(nonceMempool, bApp)
	//
	// bApp.SetMempool(nonceMempool)
	// bApp.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// bApp.SetProcessProposal(abciPropHandler.ProcessProposalHandler())
	//
	// Alternatively, you can construct BaseApp options, append those to
	// baseAppOptions and pass them to NewBaseApp.
	//
	// Example:
	//
	// prepareOpt = func(app *baseapp.BaseApp) {
	// 	abciPropHandler := baseapp.NewDefaultProposalHandler(nonceMempool, app)
	// 	app.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// }
	// baseAppOptions = append(baseAppOptions, prepareOpt)

	// create and set dummy vote extension handler
	cometService := runtime.NewContextAwareCometInfoService()
	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		evidencetypes.StoreKey, accounts.StoreKey,
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, consensusparamtypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		circuittypes.StoreKey, pooltypes.StoreKey,
		authzkeeper.StoreKey, group.StoreKey, feemarkettypes.StoreKey,
	)

	// register streaming services
	if err := bApp.RegisterStreamingServices(appOpts, keys); err != nil {
		panic(err)
	}

	app := &SimApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
	}

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, runtime.NewEnvironment(runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]), app.Logger()), authtypes.NewModuleAddress(govtypes.ModuleName).String())
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	// add keepers
	govModuleAddr, err := signingCtx.AddressCodec().BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	if err != nil {
		panic(err)
	}

	accountsKeeper, err := accounts.NewKeeper(
		appCodec,
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[accounts.StoreKey]), logger.With(log.ModuleKey, "x/accounts"), runtime.EnvWithMsgRouterService(app.MsgServiceRouter()), runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())),
		signingCtx.AddressCodec(),
		appCodec.InterfaceRegistry(),
		txDecoder,
		// Lockup account
		accountstd.AddAccount(lockup.CONTINUOUS_LOCKING_ACCOUNT, lockup.NewContinuousLockingAccount),
		accountstd.AddAccount(lockup.PERIODIC_LOCKING_ACCOUNT, lockup.NewPeriodicLockingAccount),
		accountstd.AddAccount(lockup.DELAYED_LOCKING_ACCOUNT, lockup.NewDelayedLockingAccount),
		accountstd.AddAccount(lockup.PERMANENT_LOCKING_ACCOUNT, lockup.NewPermanentLockingAccount),
		accountstd.AddAccount("multisig", multisig.NewAccount),
		// PRODUCTION: add
		baseaccount.NewAccount("base", txConfig.SignModeHandler(), baseaccount.WithSecp256K1PubKey()),
	)
	if err != nil {
		panic(err)
	}
	app.AccountKeeper = accountsKeeper

	app.AuthKeeper = authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), logger.With(log.ModuleKey, "x/auth")),
		appCodec,
		authtypes.ProtoBaseAccount,
		accountsKeeper,
		maccPerms,
		signingCtx.AddressCodec(),
		sdk.Bech32MainPrefix,
		govModuleAddr,
	)

	blockedAddrs := BlockedAddresses()
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[banktypes.StoreKey]), logger.With(log.ModuleKey, "x/bank")),
		appCodec,
		app.AuthKeeper,
		blockedAddrs,
		govModuleAddr,
	)

	// optional: enable sign mode textual by overwriting the default tx config (after setting the bank keeper)
	enabledSignModes := append(authtx.DefaultSignModes, apisigning.SignMode_SIGN_MODE_TEXTUAL) //nolint:gocritic
	txConfigOpts := authtx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: txmodule.NewBankKeeperCoinMetadataQueryFn(app.BankKeeper),
		SigningOptions: &signing.Options{
			AddressCodec:          signingCtx.AddressCodec(),
			ValidatorAddressCodec: signingCtx.ValidatorAddressCodec(),
		},
	}
	txConfig, err = authtx.NewTxConfigWithOptions(
		appCodec,
		txConfigOpts,
	)
	if err != nil {
		panic(err)
	}
	app.txConfig = txConfig

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewEnvironment(
			runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
			logger.With(log.ModuleKey, "x/staking"),
			runtime.EnvWithMsgRouterService(app.MsgServiceRouter()),
			runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())),
		app.AuthKeeper,
		app.BankKeeper,
		app.ConsensusParamsKeeper,
		govModuleAddr,
		signingCtx.ValidatorAddressCodec(),
		authcodec.NewBech32Codec(sdk.Bech32PrefixConsAddr),
		cometService,
	)

	app.MintKeeper = mintkeeper.NewKeeper(appCodec, runtime.NewEnvironment(runtime.NewKVStoreService(keys[minttypes.StoreKey]), logger.With(log.ModuleKey, "x/mint")), app.AuthKeeper, app.BankKeeper, authtypes.FeeCollectorName, govModuleAddr)
	if err := app.MintKeeper.SetMintFn(mintkeeper.DefaultMintFn(minttypes.DefaultInflationCalculationFn, app.StakingKeeper, app.MintKeeper)); err != nil {
		panic(err)
	}

	app.PoolKeeper = poolkeeper.NewKeeper(appCodec, runtime.NewEnvironment(runtime.NewKVStoreService(keys[pooltypes.StoreKey]), logger.With(log.ModuleKey, "x/protocolpool")), app.AuthKeeper, app.BankKeeper, govModuleAddr)

	app.DistrKeeper = distrkeeper.NewKeeper(appCodec, runtime.NewEnvironment(runtime.NewKVStoreService(keys[distrtypes.StoreKey]), logger.With(log.ModuleKey, "x/distribution")), app.AuthKeeper, app.BankKeeper, app.StakingKeeper, cometService, authtypes.FeeCollectorName, govModuleAddr)

	app.SlashingKeeper = slashingkeeper.NewKeeper(runtime.NewEnvironment(runtime.NewKVStoreService(keys[slashingtypes.StoreKey]), logger.With(log.ModuleKey, "x/slashing")),
		appCodec, legacyAmino, app.StakingKeeper, govModuleAddr,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(runtime.NewEnvironment(runtime.NewKVStoreService(keys[feegrant.StoreKey]), logger.With(log.ModuleKey, "x/feegrant")), appCodec, app.AuthKeeper)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	app.CircuitKeeper = circuitkeeper.NewKeeper(runtime.NewEnvironment(runtime.NewKVStoreService(keys[circuittypes.StoreKey]), logger.With(log.ModuleKey, "x/circuit")), appCodec, govModuleAddr, app.AuthKeeper.AddressCodec())
	app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper)

	app.AuthzKeeper = authzkeeper.NewKeeper(runtime.NewEnvironment(runtime.NewKVStoreService(keys[authzkeeper.StoreKey]), logger.With(log.ModuleKey, "x/authz"), runtime.EnvWithMsgRouterService(app.MsgServiceRouter()), runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())), appCodec, app.AuthKeeper)

	groupConfig := group.DefaultConfig()
	/*
		Example of setting group params:
		groupConfig.MaxMetadataLen = 1000
	*/
	app.GroupKeeper = groupkeeper.NewKeeper(runtime.NewEnvironment(runtime.NewKVStoreService(keys[group.StoreKey]), logger.With(log.ModuleKey, "x/group"), runtime.EnvWithMsgRouterService(app.MsgServiceRouter()), runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())), appCodec, app.AuthKeeper, groupConfig)

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	// set the governance module account as the authority for conducting upgrades
	app.UpgradeKeeper = upgradekeeper.NewKeeper(runtime.NewEnvironment(runtime.NewKVStoreService(keys[upgradetypes.StoreKey]), logger.With(log.ModuleKey, "x/upgrade"), runtime.EnvWithMsgRouterService(app.MsgServiceRouter()), runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())), skipUpgradeHeights, appCodec, homePath, app.BaseApp, govModuleAddr, app.ConsensusParamsKeeper)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govConfig := govkeeper.DefaultConfig()
	/*
		Example of setting gov params:
		govConfig.MaxMetadataLen = 10000
	*/
	govKeeper := govkeeper.NewKeeper(appCodec, runtime.NewEnvironment(runtime.NewKVStoreService(keys[govtypes.StoreKey]), logger.With(log.ModuleKey, "x/gov"), runtime.EnvWithMsgRouterService(app.MsgServiceRouter()), runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())), app.AuthKeeper, app.BankKeeper, app.StakingKeeper, app.PoolKeeper, govConfig, govModuleAddr)

	// Set legacy router for backwards compatibility with gov v1beta1
	govKeeper.SetLegacyRouter(govRouter)

	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
			// register the governance hooks
		),
	)

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, runtime.NewEnvironment(runtime.NewKVStoreService(keys[evidencetypes.StoreKey]), logger.With(log.ModuleKey, "x/evidence"), runtime.EnvWithMsgRouterService(app.MsgServiceRouter()), runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())), app.StakingKeeper, app.SlashingKeeper, app.ConsensusParamsKeeper, app.AuthKeeper.AddressCodec(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	app.FeeMarketKeeper = feemarketkeeper.NewKeeper(appCodec, runtime.NewEnvironment(runtime.NewKVStoreService(keys[feemarkettypes.StoreKey]), logger), app.AuthKeeper, &feemarkettypes.TestDenomResolver{}, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	/****  Module Options ****/

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(appCodec, app.AuthKeeper, app.StakingKeeper, app, txConfig, genutiltypes.DefaultMessageValidator),
		accounts.NewAppModule(appCodec, app.AccountKeeper),
		auth.NewAppModule(appCodec, app.AuthKeeper, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
		vesting.NewAppModule(app.AuthKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AuthKeeper),
		feegrantmodule.NewAppModule(appCodec, app.FeeGrantKeeper, app.interfaceRegistry),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AuthKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AuthKeeper, app.BankKeeper, app.StakingKeeper, app.interfaceRegistry, cometService),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.StakingKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(appCodec, app.EvidenceKeeper, cometService),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AuthKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		circuit.NewAppModule(appCodec, app.CircuitKeeper),
		protocolpool.NewAppModule(appCodec, app.PoolKeeper, app.AuthKeeper, app.BankKeeper),
		feemarket.NewAppModule(appCodec, *app.FeeMarketKeeper),
	)

	app.ModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.ModuleManager.RegisterInterfaces(interfaceRegistry)

	// NOTE: upgrade module is required to be prioritized
	app.ModuleManager.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)
	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.ModuleManager.SetOrderBeginBlockers(
		minttypes.ModuleName,
		feemarkettypes.ModuleName,
		distrtypes.ModuleName,
		pooltypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
	)
	app.ModuleManager.SetOrderEndBlockers(
		feemarkettypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		pooltypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	genesisModuleOrder := []string{
		authtypes.ModuleName, accounts.ModuleName, banktypes.ModuleName,
		distrtypes.ModuleName, stakingtypes.ModuleName, slashingtypes.ModuleName, govtypes.ModuleName,
		minttypes.ModuleName, genutiltypes.ModuleName, evidencetypes.ModuleName, authz.ModuleName,
		feegrant.ModuleName, nft.ModuleName, group.ModuleName, upgradetypes.ModuleName,
		vestingtypes.ModuleName, consensusparamtypes.ModuleName, circuittypes.ModuleName, pooltypes.ModuleName,
		feemarkettypes.ModuleName,
	}
	app.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	app.ModuleManager.SetOrderExportGenesis(genesisModuleOrder...)

	// Uncomment if you want to set a custom migration order here.
	// app.ModuleManager.SetOrderMigrations(custom order)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err = app.ModuleManager.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// add test gRPC service for testing gRPC queries in isolation
	testdata_pulsar.RegisterQueryServer(app.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AuthKeeper, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	// set denom resolver to test variant.
	app.FeeMarketKeeper.SetDenomResolver(&feemarkettypes.TestDenomResolver{})

	// Create a global ante handler that will be called on each transaction when
	// proposals are being built and verified.
	anteHandlerOptions := ante.HandlerOptions{
		Environment:              runtime.NewEnvironment(nil, logger, runtime.EnvWithMsgRouterService(app.MsgServiceRouter()), runtime.EnvWithQueryRouterService(app.GRPCQueryRouter())), // nil is set as the kvstoreservice to avoid module access
		AccountKeeper:            app.AuthKeeper,
		AccountAbstractionKeeper: app.AccountKeeper,
		BankKeeper:               app.BankKeeper,
		ConsensusKeeper:          app.ConsensusParamsKeeper,
		SignModeHandler:          txConfig.SignModeHandler(),
		FeegrantKeeper:           app.FeeGrantKeeper,
		SigGasConsumer:           ante.DefaultSigVerificationGasConsumer,
	}

	anteOptions := AnteHandlerOptions{
		HandlerOptions:  anteHandlerOptions,
		AccountKeeper:   app.AuthKeeper,
		BankKeeper:      app.BankKeeper,
		FeeMarketKeeper: app.FeeMarketKeeper,
		CircuitKeeper:   &app.CircuitKeeper,
	}
	anteHandler, err := NewAnteHandler(anteOptions)
	if err != nil {
		panic(err)
	}

	postHandlerOptions := PostHandlerOptions{
		AccountKeeper:   app.AuthKeeper,
		BankKeeper:      app.BankKeeper,
		FeeMarketKeeper: app.FeeMarketKeeper,
	}
	postHandler, err := NewPostHandler(postHandlerOptions)
	if err != nil {
		panic(err)
	}

	// set ante and post handlers
	app.SetAnteHandler(anteHandler)
	app.SetPostHandler(postHandler)

	// At startup, after all modules have been registered, check that all prot
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		panic(err)
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		fmt.Fprintln(os.Stderr, err.Error())
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return app
}

// Name returns the name of the App
func (app *SimApp) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *SimApp) PreBlocker(ctx sdk.Context, _ *abci.FinalizeBlockRequest) error {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *SimApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *SimApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

func (app *SimApp) Configurator() module.Configurator {
	return app.configurator
}

// InitChainer application update at chain initialization
func (app *SimApp) InitChainer(ctx sdk.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	var genesisState GenesisState
	err := json.Unmarshal(req.AppStateBytes, &genesisState)
	if err != nil {
		return nil, err
	}
	err = app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	if err != nil {
		return nil, err
	}
	return app.ModuleManager.InitGenesis(ctx, genesisState)
}

// LoadHeight loads a particular height
func (app *SimApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *SimApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns SimApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *SimApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns SimApp's InterfaceRegistry
func (app *SimApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns SimApp's TxConfig
func (app *SimApp) TxConfig() client.TxConfig {
	return app.txConfig
}

// AutoCliOpts returns the autocli options for the app.
func (app *SimApp) AutoCliOpts() autocli.AppOptions {
	return autocli.AppOptions{
		Modules:               app.ModuleManager.Modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

// DefaultGenesis returns a default genesis from the registered AppModule's.
func (app *SimApp) DefaultGenesis() map[string]json.RawMessage {
	return app.ModuleManager.DefaultGenesis()
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *SimApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetStoreKeys returns all the stored store keys.
func (app *SimApp) GetStoreKeys() []storetypes.StoreKey {
	keys := make([]storetypes.StoreKey, 0, len(app.keys))
	for _, key := range app.keys {
		keys = append(keys, key)
	}

	return keys
}

// SimulationManager implements the SimulationApp interface
func (app *SimApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *SimApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new CometBFT queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.ModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *SimApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *SimApp) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		cmtApp.Query,
	)
}

// ValidatorKeyProvider returns a function that generates a validator key
// Supported key types are those supported by Comet: ed25519, secp256k1, bls12-381
func (app *SimApp) ValidatorKeyProvider() runtime.KeyGenF {
	return func() (cmtcrypto.PrivKey, error) {
		return cmted25519.GenPrivKey(), nil
	}
}

func (app *SimApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// GetMaccPerms returns a copy of the module account permissions
//
// NOTE: This is solely to be used for testing purposes.
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}

	return dupMaccPerms
}

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}
