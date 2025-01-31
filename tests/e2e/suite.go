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

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v9"
	"github.com/strangelove-ventures/interchaintest/v9/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v9/ibc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	govtypes "cosmossdk.io/x/gov/types"
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

// TestSuite runs the feemarket e2e test-suite against a given interchaintest specification
type TestSuite struct {
	suite.Suite
	// spec
	spec *interchaintest.ChainSpec
	// add more fields here as necessary
	chain *cosmos.CosmosChain
	// users
	user1, user2, user3 ibc.Wallet

	// overrides for key-ring configuration of the broadcaster
	broadcasterOverrides *KeyringOverride

	// bc is the RPC interface to the ITS network
	bc *cosmos.Broadcaster

	cdc codec.Codec

	// default token denom
	denom string

	gasPrices string

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

// WithGasPrices sets gas prices.
func WithGasPrices(gasPrices string) Option {
	return func(s *TestSuite) {
		s.gasPrices = gasPrices
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

func NewIntegrationSuite(spec *interchaintest.ChainSpec, txCfg TestTxConfig, opts ...Option) *TestSuite {
	if err := txCfg.Validate(); err != nil {
		panic(err)
	}

	suite := &TestSuite{
		spec:      spec,
		denom:     defaultDenom,
		gasPrices: "",
		authority: authtypes.NewModuleAddress(govtypes.ModuleName),
		icc:       DefaultInterchainConstructor,
		cc:        DefaultChainConstructor,
		txConfig:  txCfg,
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
	chains := s.cc(s.T(), s.spec, s.gasPrices)

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

	if len(chains) < 1 {
		panic("no chains created")
	}

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

	_ = s.chain.StopAllNodes(ctx)
	_ = s.chain.StopAllSidecars(ctx)
}

func (s *TestSuite) SetupSubTest() {
	// wait for 1 block height
	// query height
	height, err := s.chain.Height(context.Background())
	s.Require().NoError(err)
	s.WaitForHeight(s.chain, height+1)

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
	// get nodes
	nodes := s.chain.Nodes()
	s.Require().True(len(nodes) > 0)

	params := s.QueryParams()

	defaultGasPrice := s.QueryDefaultGasPrice()
	gas := int64(200000)
	minBaseFee := sdk.NewDecCoinFromDec(defaultGasPrice.Denom, defaultGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
	minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.TruncateInt()))
	sendAmt := int64(100)

	s.Run("expect fee market state to decrease", func() {
		s.T().Log("performing sends...")
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
			wg := &sync.WaitGroup{}
			wg.Add(3)

			smallSend := func(wg *sync.WaitGroup, userA, userB ibc.Wallet) {
				defer wg.Done()
				coins := sdk.NewCoins(sdk.NewCoin(s.chain.Config().Denom, math.NewInt(sendAmt)))
				bal := s.QueryBalance(userA)
				s.T().Logf("user %q has balance: %s", userA.FormattedAddress(), bal.String())
				bal = s.QueryBalance(userB)
				s.T().Logf("user %q has balance: %s", userB.FormattedAddress(), bal.String())
				s.T().Logf("sending %s from %q to %q (gas: %d, fee: %s)", coins.String(), userA.FormattedAddress(), userB.FormattedAddress(), gas, minBaseFee.String())

				txResp, err := s.SendCoinsMultiBroadcast(
					context.Background(),
					userA,
					userB,
					coins,
					minBaseFeeCoins,
					gas,
					s.txConfig.SmallSendsNum,
				)
				s.T().Logf("txResp: %v", txResp)
				s.T().Logf("error?: %v", err)
				if err != nil {
					s.T().Log(err)
				} else if txResp != nil && txResp.CheckTx.Code != 0 {
					s.T().Log(txResp.CheckTx)
				}
			}

			go smallSend(wg, s.user1, s.user2)
			go smallSend(wg, s.user3, s.user2)
			go smallSend(wg, s.user2, s.user1)

			wg.Wait()
		}

		// wait for 5 blocks
		// query height
		height, err := s.chain.Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.chain, height+5)

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
	// get nodes
	nodes := s.chain.Nodes()
	s.Require().True(len(nodes) > 0)

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

			wg := &sync.WaitGroup{}
			wg.Add(3)

			largeSend := func(wg *sync.WaitGroup, userA, userB ibc.Wallet) {
				defer wg.Done()
				txResp, err := s.SendCoinsMultiBroadcast(
					context.Background(),
					userA,
					userB,
					sdk.NewCoins(sdk.NewCoin(s.chain.Config().Denom, math.NewInt(sendAmt))),
					minBaseFeeCoins,
					gas,
					s.txConfig.LargeSendsNum,
				)
				if err != nil {
					s.T().Log(err)
				} else if txResp != nil && txResp.CheckTx.Code != 0 {
					s.T().Log(txResp.CheckTx)
				}
			}

			go largeSend(wg, s.user1, s.user2)
			go largeSend(wg, s.user3, s.user2)
			go largeSend(wg, s.user2, s.user1)

			wg.Wait()
		}

		// wait for 5 blocks
		// query height
		height, err := s.chain.Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.chain, height+5)

		gasPrice := s.QueryDefaultGasPrice()
		s.T().Log("gas price", gasPrice.String())

		amt, err := s.chain.GetBalance(context.Background(), s.user1.FormattedAddress(), gasPrice.Denom)
		s.Require().NoError(err)
		s.T().Log("balance:", amt.String())
	})
}

func (s *TestSuite) TestSendTxFailures() {
	sendAmt := int64(100)
	gas := int64(200000)

	s.Run("submit tx with no gas attached", func() {
		// send one tx with no  gas or fee attached
		txResp, err := s.SendCoinsMultiBroadcast(
			context.Background(),
			s.user1,
			s.user3,
			sdk.NewCoins(sdk.NewCoin(s.chain.Config().Denom, math.NewInt(sendAmt))),
			sdk.NewCoins(),
			0,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code != 0)
		s.T().Log(txResp.CheckTx.Log)
		s.Require().Contains(txResp.CheckTx.Log, "out of gas")
	})

	s.Run("submit tx with no fee", func() {
		txResp, err := s.SendCoinsMultiBroadcast(
			context.Background(),
			s.user1,
			s.user3,
			sdk.NewCoins(sdk.NewCoin(s.chain.Config().Denom, math.NewInt(sendAmt))),
			sdk.NewCoins(),
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code != 0)
		s.T().Log(txResp.CheckTx.Log)
		s.Require().Contains(txResp.CheckTx.Log, "no fee coin provided")
	})

	s.Run("fail a tx that uses full balance in fee - fail tx", func() {
		balance := s.QueryBalance(s.user3)

		// send one tx with no  gas or fee attached
		txResp, err := s.SendCoinsMultiBroadcast(
			context.Background(),
			s.user3,
			s.user1,
			sdk.NewCoins(balance),
			sdk.NewCoins(balance),
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code == 0)
		s.Require().True(txResp.TxResult.Code != 0)
		s.T().Log(txResp.TxResult.Log)
		s.Require().Contains(txResp.TxResult.Log, "insufficient funds")

		// ensure that balance is deducted for any tx passing checkTx
		newBalance := s.QueryBalance(s.user3)
		s.Require().True(newBalance.IsLT(balance), fmt.Sprintf("new balance: %d, original balance: %d",
			balance.Amount.Int64(),
			newBalance.Amount.Int64()))
	})

	s.Run("submit a tx for full balance - fail tx", func() {
		balance := s.QueryBalance(s.user1)

		defaultGasPrice := s.QueryDefaultGasPrice()
		minBaseFee := sdk.NewDecCoinFromDec(defaultGasPrice.Denom, defaultGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
		minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.TruncateInt().Add(math.
			NewInt(100))))
		txResp, err := s.SendCoinsMultiBroadcast(
			context.Background(),
			s.user1,
			s.user3,
			sdk.NewCoins(balance),
			minBaseFeeCoins,
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code == 0)
		s.Require().True(txResp.TxResult.Code != 0)
		s.T().Log(txResp.TxResult.Log)
		s.Require().Contains(txResp.TxResult.Log, "insufficient funds")

		// ensure that balance is deducted for any tx passing checkTx
		newBalance := s.QueryBalance(s.user3)
		s.Require().True(newBalance.IsLT(balance), fmt.Sprintf("new balance: %d, original balance: %d",
			balance.Amount.Int64(),
			newBalance.Amount.Int64()))
	})

	s.Run("submit a tx with fee greater than full balance - fail checktx", func() {
		balance := s.QueryBalance(s.user1)
		txResp, err := s.SendCoinsMultiBroadcast(
			context.Background(),
			s.user1,
			s.user3,
			sdk.NewCoins(sdk.NewCoin(s.chain.Config().Denom, math.NewInt(sendAmt))),
			sdk.NewCoins(balance.AddAmount(math.NewInt(110000))),
			gas,
			1,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.CheckTx.Code != 0)
		s.T().Log(txResp.CheckTx.Log)
		s.Require().Contains(txResp.CheckTx.Log, "error escrowing funds")

		// ensure that no balance is deducted for a tx failing checkTx
		newBalance := s.QueryBalance(s.user1)
		s.Require().True(newBalance.Equal(balance), fmt.Sprintf("new balance: %d, original balance: %d",
			balance.Amount.Int64(),
			newBalance.Amount.Int64()))
	})

	s.Run("submit 2 tx in the same block - fail checktx in 2nd", func() {
		balance := s.QueryBalance(s.user2)

		defaultGasPrice := s.QueryDefaultGasPrice()
		minBaseFee := sdk.NewDecCoinFromDec(defaultGasPrice.Denom, defaultGasPrice.Amount.Mul(math.LegacyNewDec(gas)))
		minBaseFeeCoins := sdk.NewCoins(sdk.NewCoin(minBaseFee.Denom, minBaseFee.Amount.TruncateInt().Add(math.
			NewInt(100))))
		// send one tx with no  gas or fee attached
		txResp, err := s.SendCoinsMultiBroadcastAsync(
			context.Background(),
			s.user2,
			s.user1,
			sdk.NewCoins(balance.SubAmount(minBaseFeeCoins.AmountOf(minBaseFee.Denom))),
			minBaseFeeCoins,
			gas,
			1,
			false,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.Code == 0)
		hash1 := txResp.Hash

		txResp, err = s.SendCoinsMultiBroadcastAsync(
			context.Background(),
			s.user2,
			s.user1,
			minBaseFeeCoins,
			minBaseFeeCoins,
			gas,
			1,
			true,
		)
		s.Require().NoError(err)
		s.Require().NotNil(txResp)
		s.Require().True(txResp.Code == 0)
		hash2 := txResp.Hash

		nodes := s.chain.Nodes()
		s.Require().True(len(nodes) > 0)

		// wait for 5 blocks
		// query height
		height, err := s.chain.Height(context.Background())
		s.Require().NoError(err)
		s.WaitForHeight(s.chain, height+5)

		// after waiting, we can now query the Tx Responses
		resp, err := nodes[0].TxHashToResponse(context.Background(), hash1.String())
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Require().True(resp.Code == 0)

		resp, err = nodes[0].TxHashToResponse(context.Background(), hash2.String())
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Require().True(resp.Code != 0)
		s.Require().Contains(resp.RawLog, "error escrowing funds")
		s.T().Log(resp.RawLog)

		// reset the users and balances
		s.user2 = s.GetAndFundTestUsers(context.Background(), s.T().Name(), 200000000000, s.chain)
	})
}
