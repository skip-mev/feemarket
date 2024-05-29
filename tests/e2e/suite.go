package e2e

import (
	"context"
	"math/rand"
	"sync"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	initBalance = 30000000000000
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
	// interchain
	ic *interchaintest.Interchain
	// users
	user1, user2, user3 ibc.Wallet
	// denom
	denom string

	// overrides for key-ring configuration of the broadcaster
	broadcasterOverrides *KeyringOverride

	// bc is the RPC interface to the ITS network
	bc *cosmos.Broadcaster

	cdc codec.Codec
}

func NewE2ETestSuiteFromSpec(spec *interchaintest.ChainSpec) *TestSuite {
	return &TestSuite{
		spec:  spec,
		denom: "stake",
	}
}

func (s *TestSuite) WithDenom(denom string) *TestSuite {
	s.denom = denom

	// update the bech32 prefixes
	sdk.GetConfig().SetBech32PrefixForAccount(s.denom, s.denom+sdk.PrefixPublic)
	sdk.GetConfig().SetBech32PrefixForValidator(s.denom+sdk.PrefixValidator, s.denom+sdk.PrefixValidator+sdk.PrefixPublic)
	sdk.GetConfig().Seal()
	return s
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
	s.chain = ChainBuilderFromChainSpec(s.T(), s.spec)

	// build the interchain
	s.T().Log("building interchain")
	ctx := context.Background()
	s.ic = BuildInterchain(s.T(), ctx, s.chain)

	cc, ok := s.chain.(*cosmos.CosmosChain)
	if !ok {
		panic("unable to assert ibc.Chain as CosmosChain")
	}

	// create the broadcaster
	s.T().Log("creating broadcaster")
	s.setupBroadcaster()

	s.cdc = s.chain.Config().EncodingConfig.Codec

	// get the users
	s.user1 = s.GetAndFundTestUsers(ctx, s.T().Name(), initBalance, cc)
	s.user2 = s.GetAndFundTestUsers(ctx, s.T().Name(), initBalance, cc)
	s.user3 = s.GetAndFundTestUsers(ctx, s.T().Name(), initBalance, cc)

	// create the broadcaster
	s.T().Log("creating broadcaster")
	s.setupBroadcaster()
}

func (s *TestSuite) TearDownSuite() {
	// close the interchain
	s.Require().NoError(s.ic.Close())
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
	sendAmt := int64(100000)

	s.Run("expect fee market state to decrease", func() {
		s.T().Log("performing sends...")
		for {
			// send with the exact expected defaultGasPrice

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
			gasPrice := s.QueryDefaultGasPrice()
			s.T().Log("base defaultGasPrice", gasPrice.String())

			if gasPrice.Amount.Equal(params.MinBaseGasPrice) {
				break
			}
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

	baseGasPrice := s.QueryDefaultGasPrice()
	gas := int64(20000100)
	sendAmt := int64(100)

	s.Run("expect fee market gas price to increase", func() {
		s.T().Log("performing sends...")
		for {
			// send with the exact expected baseGasPrice
			baseGasPrice = s.QueryDefaultGasPrice()
			minBaseFee := sdk.NewDecCoinFromDec(baseGasPrice.Denom, baseGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
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
			baseGasPrice = s.QueryDefaultGasPrice()
			s.T().Log("gas price", baseGasPrice.String())

			if baseGasPrice.Amount.GT(math.LegacyNewDec(1000000)) {
				break
			}
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
