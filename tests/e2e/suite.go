package e2e

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	envKeepAlive = "FEEMARKET_INTEGRATION_KEEPALIVE"
	initBalance  = 30000000000000
	defaultDenom = "stake"
)

var r *rand.Rand

// initialize random generator with fixed seed for reproducibility
func init() {
	s := rand.NewSource(1)
	r = rand.New(s)
}

<<<<<<< HEAD
=======
func DefaultOracleSidecar(image ibc.DockerImage) ibc.SidecarConfig {
	return ibc.SidecarConfig{
		ProcessName: "oracle",
		Image:       image,
		HomeDir:     "/oracle",
		Ports:       []string{"8080", "8081"},
		StartCmd: []string{
			"slinky",
			"--oracle-config", "/oracle/oracle.json",
		},
		ValidatorProcess: true,
		PreStart:         true,
	}
}

func DefaultOracleConfig(url string) oracleconfig.OracleConfig {
	cfg := marketmap.DefaultAPIConfig
	cfg.Endpoints = []oracleconfig.Endpoint{
		{
			URL: url,
		},
	}

	// Create the oracle config
	oracleConfig := oracleconfig.OracleConfig{
		UpdateInterval: 500 * time.Millisecond,
		MaxPriceAge:    1 * time.Minute,
		Host:           "0.0.0.0",
		Port:           "8080",
		Providers: map[string]oracleconfig.ProviderConfig{
			marketmap.Name: {
				Name: marketmap.Name,
				API:  cfg,
				Type: "market_map_provider",
			},
		},
	}

	return oracleConfig
}

func DefaultMarketMap() mmtypes.MarketMap {
	return mmtypes.MarketMap{}
}

func GetOracleSideCar(node *cosmos.ChainNode) *cosmos.SidecarProcess {
	if len(node.Sidecars) == 0 {
		panic("no sidecars found")
	}
	return node.Sidecars[0]
}

type TestTxConfig struct {
	SmallSendsNum          int
	LargeSendsNum          int
	TargetIncreaseGasPrice math.LegacyDec
}

func (tx *TestTxConfig) Validate() error {
	if tx.SmallSendsNum < 1 || tx.LargeSendsNum < 1 {
		return fmt.Errorf("sends num should be greater than 1")
	}

	if tx.TargetIncreaseGasPrice.IsNil() {
		return fmt.Errorf("target increase gas price is nil")
	}

	if tx.TargetIncreaseGasPrice.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("target increase gas price is less than or equal to 0")
	}

	return nil
}

>>>>>>> daca8c8 (test: make extendable (#112))
// TestSuite runs the feemarket e2e test-suite against a given interchaintest specification
type TestSuite struct {
	suite.Suite
	// spec
	spec *interchaintest.ChainSpec
	// chain
	chain ibc.Chain
	// users
	user1, user2, user3 ibc.Wallet

	// overrides for key-ring configuration of the broadcaster
	broadcasterOverrides *KeyringOverride

	// bc is the RPC interface to the ITS network
	bc *cosmos.Broadcaster

	cdc codec.Codec

	// default token denom
	denom string

	// authority address
	authority sdk.AccAddress

	// block time
	blockTime time.Duration

	// interchain constructor
	icc InterchainConstructor

	// interchain
	ic Interchain

	// chain constructor
	cc ChainConstructor

	txConfig TestTxConfig
}

// Option is a function that modifies the TestSuite
type Option func(*TestSuite)

// WithDenom sets the token denom
func WithDenom(denom string) Option {
	return func(s *TestSuite) {
		s.denom = denom
	}
}

// WithAuthority sets the authority address
func WithAuthority(addr sdk.AccAddress) Option {
	return func(s *TestSuite) {
		s.authority = addr
	}
}

// WithBlockTime sets the block time
func WithBlockTime(t time.Duration) Option {
	return func(s *TestSuite) {
		s.blockTime = t
	}
}

// WithInterchainConstructor sets the interchain constructor
func WithInterchainConstructor(ic InterchainConstructor) Option {
	return func(s *TestSuite) {
		s.icc = ic
	}
}

// WithChainConstructor sets the chain constructor
func WithChainConstructor(cc ChainConstructor) Option {
	return func(s *TestSuite) {
		s.cc = cc
	}
}

<<<<<<< HEAD
func NewIntegrationSuite(spec *interchaintest.ChainSpec, opts ...Option) *TestSuite {
	suite := &TestSuite{
		spec:      spec,
		denom:     defaultDenom,
		authority: authtypes.NewModuleAddress(govtypes.ModuleName),
		icc:       DefaultInterchainConstructor,
		cc:        DefaultChainConstructor,
=======
func NewIntegrationSuite(spec *interchaintest.ChainSpec, oracleImage ibc.DockerImage, txCfg TestTxConfig, opts ...Option) *TestSuite {
	if err := txCfg.Validate(); err != nil {
		panic(err)
	}

	suite := &TestSuite{
		spec:         spec,
		oracleConfig: DefaultOracleSidecar(oracleImage),
		denom:        defaultDenom,
		gasPrices:    "",
		authority:    authtypes.NewModuleAddress(govtypes.ModuleName),
		icc:          DefaultInterchainConstructor,
		cc:           DefaultChainConstructor,
		txConfig:     txCfg,
>>>>>>> daca8c8 (test: make extendable (#112))
	}

	for _, opt := range opts {
		opt(suite)
	}

	return suite
}

func (s *TestSuite) WithKeyringOptions(cdc codec.Codec, opts keyring.Option) {
	s.broadcasterOverrides = &KeyringOverride{
		cdc:            cdc,
		keyringOptions: opts,
	}
}

func (s *TestSuite) SetupSuite() {
	// build the chain
	s.T().Log("building chain with spec", s.spec)
	chains := s.cc(s.T(), s.spec)

	// build the interchain
	s.T().Log("building interchain")
	ctx := context.Background()
	// start the chain
	s.ic = s.icc(context.Background(), s.T(), chains)

	s.chain = chains[0]

	// create the broadcaster
	s.T().Log("creating broadcaster")
	s.setupBroadcaster()

	s.cdc = s.chain.Config().EncodingConfig.Codec

	// get the users
	s.user1 = s.GetAndFundTestUsers(ctx, s.T().Name(), initBalance, chains[0])
	s.user2 = s.GetAndFundTestUsers(ctx, s.T().Name(), initBalance, chains[0])
	s.user3 = s.GetAndFundTestUsers(ctx, s.T().Name(), initBalance, chains[0])

	// create the broadcaster
	s.T().Log("creating broadcaster")
	s.setupBroadcaster()
}

func (s *TestSuite) TearDownSuite() {
	defer s.Teardown()
	// get the oracle integration-test suite keep alive env
	if ok := os.Getenv(envKeepAlive); ok == "" {
		return
	}

	// await on a signal to keep the chain running
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s.T().Log("Keeping the chain running")
	<-sig
}

func (s *TestSuite) Teardown() {
	// stop all nodes + sidecars in the chain
	ctx := context.Background()
	if s.chain == nil {
		return
	}

	cc, ok := s.chain.(*cosmos.CosmosChain)
	if !ok {
		panic("unable to assert ibc.Chain as CosmosChain")
	}

	_ = cc.StopAllNodes(ctx)
	_ = cc.StopAllSidecars(ctx)
}

func (s *TestSuite) SetupSubTest() {
	// wait for 1 block height
	// query height
	height, err := s.chain.(*cosmos.CosmosChain).Height(context.Background())
	s.Require().NoError(err)
	s.WaitForHeight(s.chain.(*cosmos.CosmosChain), height+1)

	state := s.QueryState()
	s.T().Log("state at block height", height+1, ":", state.String())
	gasPrice := s.QueryDefaultGasPrice()
	s.T().Log("gas price at block height", height+1, ":", gasPrice.String())
}

func (s *TestSuite) TestQueryParams() {
	s.Run("query params", func() {
		// query params
		params := s.QueryParams()

		// expect validate to pass
		require.NoError(s.T(), params.ValidateBasic(), params)
	})
}

func (s *TestSuite) TestQueryState() {
	s.Run("query state", func() {
		// query state
		state := s.QueryState()

		// expect validate to pass
		require.NoError(s.T(), state.ValidateBasic(), state)
	})
}

func (s *TestSuite) TestQueryGasPrice() {
	s.Run("query gas price", func() {
		// query gas price
		gasPrice := s.QueryDefaultGasPrice()

		// expect validate to pass
		require.NoError(s.T(), gasPrice.Validate(), gasPrice)
	})
}

// TestSendTxDecrease tests that the feemarket will decrease until it hits the min gas price
// when gas utilization is below the target block utilization.
func (s *TestSuite) TestSendTxDecrease() {
	// cast chain to cosmos-chain
	cosmosChain, ok := s.chain.(*cosmos.CosmosChain)
	s.Require().True(ok)
	// get nodes
	nodes := cosmosChain.Nodes()
	s.Require().True(len(nodes) > 0)

	params := s.QueryParams()

	defaultGasPrice := s.QueryDefaultGasPrice()
	gas := int64(200000)
	minBaseFee := sdk.NewDecCoinFromDec(defaultGasPrice.Denom, defaultGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
	minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.TruncateInt()))
	sendAmt := int64(100)

	s.Run("expect fee market state to decrease", func() {
		s.T().Log("performing sends...")
<<<<<<< HEAD
		for {
			// send with the exact expected fee
			height, err := s.chain.(*cosmos.CosmosChain).Height(context.Background())
			s.Require().NoError(err)
			// send with the exact expected defaultGasPrice
=======
		sig := make(chan struct{})
		quit := make(chan struct{})
		defer close(quit)

		checkPrice := func(c, quit chan struct{}) {
			select {
			case <-time.After(500 * time.Millisecond):
				gasPrice := s.QueryDefaultGasPrice()
				s.T().Log("gas price", gasPrice.String())

				if gasPrice.Amount.Equal(params.MinBaseGasPrice) {
					c <- struct{}{}
				}
			case <-quit:
				return
			}
		}
		go checkPrice(sig, quit)

		select {
		case <-sig:
			break

		case <-time.After(100 * time.Millisecond):
>>>>>>> daca8c8 (test: make extendable (#112))
			wg := sync.WaitGroup{}
			wg.Add(3)

			go func() {
				defer wg.Done()
				txResp, err := s.SendCoinsMultiBroadcast(
					context.Background(),
					s.user1,
					s.user2,
					sdk.NewCoins(sdk.NewCoin(cosmosChain.Config().Denom, math.NewInt(sendAmt))),
					minBaseFeeCoins,
					gas,
					s.txConfig.SmallSendsNum,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.DeliverTx.Code, txResp.DeliverTx)
			}()

			go func() {
				defer wg.Done()
				txResp, err := s.SendCoinsMultiBroadcast(
					context.Background(),
					s.user3,
					s.user2,
					sdk.NewCoins(sdk.NewCoin(cosmosChain.Config().Denom, math.NewInt(sendAmt))),
					minBaseFeeCoins,
					gas,
					s.txConfig.SmallSendsNum,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.DeliverTx.Code, txResp.DeliverTx)
			}()

			go func() {
				defer wg.Done()
				txResp, err := s.SendCoinsMultiBroadcast(
					context.Background(),
					s.user2,
					s.user3,
					sdk.NewCoins(sdk.NewCoin(cosmosChain.Config().Denom, math.NewInt(sendAmt))),
					minBaseFeeCoins,
					gas,
					s.txConfig.SmallSendsNum,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.DeliverTx.Code, txResp.DeliverTx)
			}()
			wg.Wait()
<<<<<<< HEAD
			s.WaitForHeight(s.chain.(*cosmos.CosmosChain), height+1)

			gasPrice := s.QueryDefaultGasPrice()
			s.T().Log("base defaultGasPrice", gasPrice.String())

			if gasPrice.Amount.Equal(params.MinBaseGasPrice) {
				break
			}
=======
>>>>>>> daca8c8 (test: make extendable (#112))
		}

		// wait for 5 blocks
		// query height
		height, err := s.chain.(*cosmos.CosmosChain).Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.chain.(*cosmos.CosmosChain), height+5)

		gasPrice := s.QueryDefaultGasPrice()
		s.T().Log("gas price", gasPrice.String())

		amt, err := s.chain.GetBalance(context.Background(), s.user1.FormattedAddress(), minBaseFee.Denom)
		s.Require().NoError(err)
		s.Require().True(amt.LT(math.NewInt(initBalance)), amt)
		s.T().Log("balance:", amt.String())
	})
}

// TestSendTxIncrease tests that the feemarket will increase
// when gas utilization is above the target block utilization.
func (s *TestSuite) TestSendTxIncrease() {
	// cast chain to cosmos-chain
	cosmosChain, ok := s.chain.(*cosmos.CosmosChain)
	s.Require().True(ok)
	// get nodes
	nodes := cosmosChain.Nodes()
	s.Require().True(len(nodes) > 0)

<<<<<<< HEAD
	baseGasPrice := s.QueryDefaultGasPrice()
	gas := int64(20000100)
	sendAmt := int64(100)

=======
>>>>>>> daca8c8 (test: make extendable (#112))
	params := s.QueryParams()

	gas := int64(params.MaxBlockUtilization)
	sendAmt := int64(100)

	s.Run("expect fee market gas price to increase", func() {
		s.T().Log("performing sends...")
		sig := make(chan struct{})
		quit := make(chan struct{})
		defer close(quit)

		checkPrice := func(c, quit chan struct{}) {
			select {
			case <-time.After(500 * time.Millisecond):
				gasPrice := s.QueryDefaultGasPrice()
				s.T().Log("gas price", gasPrice.String())

				if gasPrice.Amount.GT(s.txConfig.TargetIncreaseGasPrice) {
					c <- struct{}{}
				}
			case <-quit:
				return
			}
		}
		go checkPrice(sig, quit)

		select {
		case <-sig:
			break

		case <-time.After(100 * time.Millisecond):
			// send with the exact expected baseGasPrice
			baseGasPrice := s.QueryDefaultGasPrice()
			minBaseFee := sdk.NewDecCoinFromDec(baseGasPrice.Denom, baseGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
			// add headroom
			minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.Add(math.LegacyNewDec(10)).TruncateInt()))

			height, err := s.chain.(*cosmos.CosmosChain).Height(context.Background())
			s.Require().NoError(err)
			wg := sync.WaitGroup{}
			wg.Add(3)

			go func() {
				defer wg.Done()
				txResp, err := s.SendCoinsMultiBroadcast(
					context.Background(),
					s.user1,
					s.user2,
					sdk.NewCoins(sdk.NewCoin(cosmosChain.Config().Denom, math.NewInt(sendAmt))),
					minBaseFeeCoins,
					gas,
					s.txConfig.LargeSendsNum,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.DeliverTx.Code, txResp.DeliverTx)
			}()

			go func() {
				defer wg.Done()
				txResp, err := s.SendCoinsMultiBroadcast(
					context.Background(),
					s.user3,
					s.user2,
					sdk.NewCoins(sdk.NewCoin(cosmosChain.Config().Denom, math.NewInt(sendAmt))),
					minBaseFeeCoins,
					gas,
					s.txConfig.LargeSendsNum,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.DeliverTx.Code, txResp.DeliverTx)
			}()

			go func() {
				defer wg.Done()
				txResp, err := s.SendCoinsMultiBroadcast(
					context.Background(),
					s.user2,
					s.user1,
					sdk.NewCoins(sdk.NewCoin(cosmosChain.Config().Denom, math.NewInt(sendAmt))),
					minBaseFeeCoins,
					gas,
					s.txConfig.LargeSendsNum,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.DeliverTx.Code, txResp.DeliverTx)
			}()
			wg.Wait()
<<<<<<< HEAD
			s.WaitForHeight(s.chain.(*cosmos.CosmosChain), height+1)

			baseGasPrice = s.QueryDefaultGasPrice()
			s.T().Log("gas price", baseGasPrice.String())

			if baseGasPrice.Amount.GT(params.MinBaseGasPrice.Mul(math.LegacyNewDec(10))) {
				break
			}
=======
>>>>>>> daca8c8 (test: make extendable (#112))
		}

		// wait for 5 blocks
		// query height
		height, err := s.chain.(*cosmos.CosmosChain).Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.chain.(*cosmos.CosmosChain), height+5)

		gasPrice := s.QueryDefaultGasPrice()
		s.T().Log("gas price", gasPrice.String())

		amt, err := s.chain.GetBalance(context.Background(), s.user1.FormattedAddress(), gasPrice.Denom)
		s.Require().NoError(err)
		s.Require().True(amt.LT(math.NewInt(initBalance)), amt)
		s.T().Log("balance:", amt.String())
	})
}
