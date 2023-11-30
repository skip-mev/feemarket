package integration

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	initBalance = 10000000000000
)

// TestSuite runs the feemarket integration test-suite against a given interchaintest specification
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

func NewIntegrationTestSuiteFromSpec(spec *interchaintest.ChainSpec) *TestSuite {
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
	s.user1 = s.GetAndFundTestUsers(ctx, s.T().Name(), initBalance, cc)[0]
	s.user2 = s.GetAndFundTestUsers(ctx, s.T().Name(), initBalance, cc)[0]
	s.user3 = s.GetAndFundTestUsers(ctx, s.T().Name(), initBalance, cc)[0]

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

	params := s.QueryParams()
	state := s.QueryState()

	s.T().Log("new test case at block height", height+1)
	s.T().Log("params:", params.String())
	s.T().Log("state:", state.String())
}

func (s *TestSuite) TestQueryParams() {
	s.SetupSubTest()

	// query params
	params := s.QueryParams()

	// expect validate to pass
	require.NoError(s.T(), params.ValidateBasic(), params)
}

func (s *TestSuite) TestQueryState() {
	s.SetupSubTest()

	// query params
	state := s.QueryState()

	// expect validate to pass
	require.NoError(s.T(), state.ValidateBasic(), state)
}

func (s *TestSuite) TestSendTxUpdating() {
	s.SetupSubTest()

	ctx := context.Background()

	// cast chain to cosmos-chain
	cosmosChain, ok := s.chain.(*cosmos.CosmosChain)
	s.Require().True(ok)
	// get nodes
	nodes := cosmosChain.Nodes()
	s.Require().True(len(nodes) > 0)

	state := s.QueryState()
	params := s.QueryParams()

	gas := int64(1000000)
	minBaseFee := sdk.NewCoins(sdk.NewCoin(params.FeeDenom, state.MinBaseFee.MulRaw(gas)))

	// send with the exact expected fee
	_, err := s.SendCoins(
		ctx,
		cosmosChain,
		s.user1.KeyName(),
		s.user1.FormattedAddress(),
		s.user2.FormattedAddress(),
		sdk.NewCoins(sdk.NewCoin(cosmosChain.Config().Denom, sdk.NewInt(10000))),
		minBaseFee,
		gas,
	)
	s.Require().NoError(err, s.user1.FormattedAddress(), s.user2.FormattedAddress())
}
