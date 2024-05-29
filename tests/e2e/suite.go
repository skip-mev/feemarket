package e2e

import (
	"context"
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
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

const (
	envKeepAlive          = "FEEMARKET_INTEGRATION_KEEPALIVE"
	initBalance           = 30000000000000
	genesisAmount         = 1000000000
	defaultDenom          = "stake"
	validatorKey          = "validator"
	yes                   = "yes"
	deposit               = 1000000
	userMnemonic          = "foster poverty abstract scorpion short shrimp tilt edge romance adapt only benefit moral another where host egg echo ability wisdom lizard lazy pool roast"
	userAccountAddressHex = "877E307618AB73E009A978AC32E0264791F6D40A"
)

var r *rand.Rand

// initialize random generator with fixed seed for reproducibility
func init() {
	s := rand.NewSource(1)
	r = rand.New(s)
}

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

func NewIntegrationSuite(spec *interchaintest.ChainSpec, opts ...Option) *TestSuite {
	suite := &TestSuite{
		spec:      spec,
		denom:     defaultDenom,
		authority: authtypes.NewModuleAddress(govtypes.ModuleName),
		icc:       DefaultInterchainConstructor,
		cc:        DefaultChainConstructor,
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
	fee := s.QueryBaseFee()
	s.T().Log("fee at block height", height+1, ":", fee.String())
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

func (s *TestSuite) TestQueryBaseFee() {
	s.Run("query base fee", func() {
		// query base fee
		fees := s.QueryBaseFee()

		// expect validate to pass
		require.NoError(s.T(), fees.Validate(), fees)
	})
}

// TestSendTxDecrease tests that the feemarket will decrease until it hits the min base fee
// when gas utilization is below the target block utilization.
func (s *TestSuite) TestSendTxDecrease() {
	// cast chain to cosmos-chain
	cosmosChain, ok := s.chain.(*cosmos.CosmosChain)
	s.Require().True(ok)
	// get nodes
	nodes := cosmosChain.Nodes()
	s.Require().True(len(nodes) > 0)

	params := s.QueryParams()

	baseFee := s.QueryBaseFee()
	gas := int64(200000)
	minBaseFee := baseFee.MulDec(math.LegacyNewDec(gas))[0]
	minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.TruncateInt()))
	sendAmt := int64(100000)

	s.Run("expect fee market state to decrease", func() {
		s.T().Log("performing sends...")
		for {
			// send with the exact expected fee

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
					1,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.TxResult.Code, txResp.TxResult)
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
					1,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.TxResult.Code, txResp.TxResult)
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
					1,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.TxResult.Code, txResp.TxResult)
			}()

			wg.Wait()
			fee := s.QueryBaseFee()
			s.T().Log("base fee", fee.String())

			if fee.AmountOf(feemarkettypes.DefaultFeeDenom).Equal(params.MinBaseFee) {
				break
			}
		}

		// wait for 5 blocks
		// query height
		height, err := s.chain.(*cosmos.CosmosChain).Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.chain.(*cosmos.CosmosChain), height+5)

		fee := s.QueryBaseFee()
		s.T().Log("base fee", fee.String())

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

	baseFee := s.QueryBaseFee()
	gas := int64(20000100)
	sendAmt := int64(100)

	s.Run("expect fee market fee to increase", func() {
		s.T().Log("performing sends...")
		for {
			// send with the exact expected fee
			baseFee = s.QueryBaseFee()
			minBaseFee := baseFee.MulDec(math.LegacyNewDec(gas))[0]
			// add headroom
			minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.Add(math.LegacyNewDec(10)).TruncateInt()))

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
					400,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.TxResult.Code, txResp.TxResult)
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
					400,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.TxResult.Code, txResp.TxResult)
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
					400,
				)
				s.Require().NoError(err, txResp)
				s.Require().Equal(uint32(0), txResp.CheckTx.Code, txResp.CheckTx)
				s.Require().Equal(uint32(0), txResp.TxResult.Code, txResp.TxResult)
			}()

			wg.Wait()
			baseFee = s.QueryBaseFee()
			s.T().Log("base fee", baseFee.String())

			if baseFee.AmountOf(feemarkettypes.DefaultFeeDenom).GT(math.LegacyNewDec(1000000)) {
				break
			}
		}

		// wait for 5 blocks
		// query height
		height, err := s.chain.(*cosmos.CosmosChain).Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.chain.(*cosmos.CosmosChain), height+5)

		fee := s.QueryBaseFee()
		s.T().Log("base fee", fee.String())

		amt, err := s.chain.GetBalance(context.Background(), s.user1.FormattedAddress(), baseFee[0].Denom)
		s.Require().NoError(err)
		s.Require().True(amt.LT(math.NewInt(initBalance)), amt)
		s.T().Log("balance:", amt.String())
	})
}
