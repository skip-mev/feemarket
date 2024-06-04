package e2e

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pelletier/go-toml/v2"
	"github.com/skip-mev/chaintestutil/sample"
	oracleconfig "github.com/skip-mev/slinky/oracle/config"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

const (
	oracleConfigPath = "oracle.json"
	appConfigPath    = "config/app.toml"
)

type KeyringOverride struct {
	keyringOptions keyring.Option
	cdc            codec.Codec
}

// SimulateTx simulates the provided messages, and checks whether the provided failure condition is met
func (s *TestSuite) SimulateTx(ctx context.Context, user cosmos.User, height uint64, expectFail bool, msgs ...sdk.Msg) {
	// create tx factory + Client Context
	txf, err := s.bc.GetFactory(ctx, user)
	s.Require().NoError(err)

	cc, err := s.bc.GetClientContext(ctx, user)
	s.Require().NoError(err)

	txf, err = txf.Prepare(cc)
	s.Require().NoError(err)

	// set timeout height
	if height != 0 {
		txf = txf.WithTimeoutHeight(height)
	}

	// get gas for tx
	_, _, err = tx.CalculateGas(cc, txf, msgs...)
	s.Require().Equal(err != nil, expectFail)
}

func (s *TestSuite) QueryParams() types.Params {
	s.T().Helper()

	// get grpc address
	grpcAddr := s.chain.GetHostGRPCAddress()

	// create the client
	cc, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	// create the oracle client
	c := types.NewQueryClient(cc)

	resp, err := c.Params(context.Background(), &types.ParamsRequest{})
	s.Require().NoError(err)

	return resp.Params
}

func (s *TestSuite) QueryState() types.State {
	s.T().Helper()

	// get grpc address
	grpcAddr := s.chain.GetHostGRPCAddress()

	// create the client
	cc, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	// create the oracle client
	c := types.NewQueryClient(cc)

	resp, err := c.State(context.Background(), &types.StateRequest{})
	s.Require().NoError(err)

	return resp.State
}

func (s *TestSuite) QueryDefaultGasPrice() sdk.DecCoin {
	s.T().Helper()

	// get grpc address
	grpcAddr := s.chain.GetHostGRPCAddress()

	// create the client
	cc, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	// create the oracle client
	c := types.NewQueryClient(cc)

	resp, err := c.GasPrice(context.Background(), &types.GasPriceRequest{
		Denom: sdk.DefaultBondDenom,
	})
	s.Require().NoError(err)

	return resp.GetPrice()
}

// QueryValidators queries for all the network's validators
func (s *TestSuite) QueryValidators(chain *cosmos.CosmosChain) []sdk.ValAddress {
	s.T().Helper()

	// get grpc client of the node
	grpcAddr := chain.GetHostGRPCAddress()
	cc, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	defer cc.Close()

	nodeClient := stakingtypes.NewQueryClient(cc)

	// query validators
	resp, err := nodeClient.Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{})
	s.Require().NoError(err)

	addrs := make([]sdk.ValAddress, len(resp.Validators))

	// unmarshal validators
	for i, val := range resp.Validators {
		addrBz, err := sdk.GetFromBech32(val.OperatorAddress, chain.Config().Bech32Prefix+sdk.PrefixValidator+sdk.PrefixOperator)
		s.Require().NoError(err)

		addrs[i] = sdk.ValAddress(addrBz)
	}
	return addrs
}

// QueryAccountBalance queries a given account's balance on the chain
func (s *TestSuite) QueryAccountBalance(chain ibc.Chain, address, denom string) int64 {
	s.T().Helper()

	// cast the chain to a cosmos-chain
	cosmosChain, ok := chain.(*cosmos.CosmosChain)
	s.Require().True(ok)
	// get nodes
	balance, err := cosmosChain.GetBalance(context.Background(), address, denom)
	s.Require().NoError(err)
	return balance.Int64()
}

func (s *TestSuite) QueryAccountSequence(chain *cosmos.CosmosChain, address string) uint64 {
	s.T().Helper()

	// get nodes
	nodes := chain.Nodes()
	s.Require().True(len(nodes) > 0)

	resp, _, err := nodes[0].ExecQuery(context.Background(), "auth", "account", address)
	s.Require().NoError(err)
	// unmarshal json response
	var accResp codectypes.Any
	s.Require().NoError(json.Unmarshal(resp, &accResp))

	// unmarshal into baseAccount
	var acc authtypes.BaseAccount
	s.Require().NoError(acc.Unmarshal(accResp.Value))

	return acc.GetSequence()
}

// Block returns the block at the given height
func (s *TestSuite) Block(chain *cosmos.CosmosChain, height int64) *rpctypes.ResultBlock {
	s.T().Helper()

	// get nodes
	nodes := chain.Nodes()
	s.Require().True(len(nodes) > 0)

	client := nodes[0].Client

	resp, err := client.Block(context.Background(), &height)
	s.Require().NoError(err)

	return resp
}

// WaitForHeight waits for the chain to reach the given height
func (s *TestSuite) WaitForHeight(chain *cosmos.CosmosChain, height int64) {
	s.T().Helper()

	// wait for next height
	err := testutil.WaitForCondition(30*time.Second, 100*time.Millisecond, func() (bool, error) {
		pollHeight, err := chain.Height(context.Background())
		if err != nil {
			return false, err
		}
		return pollHeight >= height, nil
	})
	s.Require().NoError(err)
}

// VerifyBlock takes a Block and verifies that it contains the given bid at the 0-th index, and the bundled txs immediately after
func (s *TestSuite) VerifyBlock(block *rpctypes.ResultBlock, offset int, bidTxHash string, txs [][]byte) {
	s.T().Helper()

	// verify the block
	if bidTxHash != "" {
		s.Require().Equal(bidTxHash, TxHash(block.Block.Data.Txs[offset+1]))
		offset += 1
	}

	// verify the txs in sequence
	for i, tx := range txs {
		s.Require().Equal(TxHash(tx), TxHash(block.Block.Data.Txs[i+offset+1]))
	}
}

// VerifyBlockWithExpectedBlock takes in a list of raw tx bytes and compares each tx hash to the tx hashes in the block.
// The expected block is the block that should be returned by the chain at the given height.
func (s *TestSuite) VerifyBlockWithExpectedBlock(chain *cosmos.CosmosChain, height uint64, txs [][]byte) {
	s.T().Helper()

	block := s.Block(chain, int64(height))
	blockTxs := block.Block.Data.Txs[1:]

	s.T().Logf("verifying block %d", height)
	s.Require().Equal(len(txs), len(blockTxs))
	for i, tx := range txs {
		s.T().Logf("verifying tx %d; expected %s, got %s", i, TxHash(tx), TxHash(blockTxs[i]))
		s.Require().Equal(TxHash(tx), TxHash(blockTxs[i]))
	}
}

func TxHash(tx []byte) string {
	return strings.ToUpper(hex.EncodeToString(comettypes.Tx(tx).Hash()))
}

func (s *TestSuite) setupBroadcaster() {
	s.T().Helper()

	bc := cosmos.NewBroadcaster(s.T(), s.chain)

	if s.broadcasterOverrides == nil {
		s.bc = bc
		return
	}

	// get the key-ring-dir from the node locally
	keyringDir := s.keyringDirFromNode()

	// create a new keyring
	kr, err := keyring.New("", keyring.BackendTest, keyringDir, os.Stdin, s.broadcasterOverrides.cdc, s.broadcasterOverrides.keyringOptions)
	s.Require().NoError(err)

	// override factory + client context keyrings
	bc.ConfigureFactoryOptions(
		func(factory tx.Factory) tx.Factory {
			return factory.WithKeybase(kr)
		},
	)
	bc.ConfigureClientContextOptions(
		func(cc client.Context) client.Context {
			return cc.WithKeyring(kr)
		},
	)

	s.bc = bc
}

// sniped from here: https://github.com/strangelove-ventures/interchaintest ref: 9341b001214d26be420f1ca1ab0f15bad17faee6
func (s *TestSuite) keyringDirFromNode() string {
	node := s.chain.Nodes()[0]

	// create a temp-dir
	localDir := s.T().TempDir()

	containerKeyringDir := path.Join(node.HomeDir(), "keyring-test")
	reader, _, err := node.DockerClient.CopyFromContainer(context.Background(), node.ContainerID(), containerKeyringDir)
	s.Require().NoError(err)

	s.Require().NoError(os.Mkdir(path.Join(localDir, "keyring-test"), os.ModePerm))

	tr := tar.NewReader(reader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		s.Require().NoError(err)

		var fileBuff bytes.Buffer
		_, err = io.Copy(&fileBuff, tr)
		s.Require().NoError(err)

		name := hdr.Name
		extractedFileName := path.Base(name)
		isDirectory := extractedFileName == ""
		if isDirectory {
			continue
		}

		filePath := path.Join(localDir, "keyring-test", extractedFileName)
		s.Require().NoError(os.WriteFile(filePath, fileBuff.Bytes(), os.ModePerm))
	}

	return localDir
}

func (s *TestSuite) SendCoinsMultiBroadcast(ctx context.Context, sender, receiver ibc.Wallet, amt, fees sdk.Coins, gas int64, numMsg int) (*coretypes.ResultBroadcastTxCommit, error) {
	msgs := make([]sdk.Msg, numMsg)
	for i := 0; i < numMsg; i++ {
		msgs[i] = &banktypes.MsgSend{
			FromAddress: sender.FormattedAddress(),
			ToAddress:   receiver.FormattedAddress(),
			Amount:      amt,
		}
	}

	tx := s.CreateTx(s.chain, sender, fees.String(), gas, msgs...)

	// get an rpc endpoint for the chain
	c := s.chain.Nodes()[0].Client
	return c.BroadcastTxCommit(ctx, tx)
}

// SendCoins creates a executes a SendCoins message and broadcasts the transaction.
func (s *TestSuite) SendCoins(ctx context.Context, keyName, sender, receiver string, amt, fees sdk.Coins, gas int64) (string, error) {
	resp, err := s.ExecTx(
		ctx,
		s.chain,
		keyName,
		false,
		"bank",
		"send",
		sender,
		receiver,
		amt.String(),
		"--fees",
		fees.String(),
		"--gas",
		strconv.FormatInt(gas, 10),
	)

	return resp, err
}

// GetAndFundTestUserWithMnemonic restores a user using the given mnemonic
// and funds it with the native chain denom.
// The caller should wait for some blocks to complete before the funds will be accessible.
func (s *TestSuite) GetAndFundTestUserWithMnemonic(
	ctx context.Context,
	keyNamePrefix, mnemonic string,
	amount int64,
	chain ibc.Chain,
) (ibc.Wallet, error) {
	chainCfg := chain.Config()
	keyName := fmt.Sprintf("%s-%s-%s", keyNamePrefix, chainCfg.ChainID, sample.AlphaString(r, 3))
	user, err := chain.BuildWallet(ctx, keyName, mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to get source user wallet: %w", err)
	}

	_, err = s.SendCoins(
		ctx,
		interchaintest.FaucetAccountKeyName,
		interchaintest.FaucetAccountKeyName,
		user.FormattedAddress(),
		sdk.NewCoins(sdk.NewCoin(chainCfg.Denom, math.NewInt(amount))),
		sdk.NewCoins(sdk.NewCoin(chainCfg.Denom, math.NewInt(1000000000000))),
		1000000,
	)
	s.Require().NoError(err, "failed to get funds from faucet")
	return user, nil
}

// GetAndFundTestUsers generates and funds chain users with the native chain denom.
// The caller should wait for some blocks to complete before the funds will be accessible.
func (s *TestSuite) GetAndFundTestUsers(
	ctx context.Context,
	keyNamePrefix string,
	amount int64,
	chain ibc.Chain,
) ibc.Wallet {
	user, err := s.GetAndFundTestUserWithMnemonic(ctx, keyNamePrefix, "", amount, chain)
	s.Require().NoError(err)

	return user
}

// ExecTx executes a cli command on a node, waits a block and queries the Tx to verify it was included on chain.
func (s *TestSuite) ExecTx(ctx context.Context, chain *cosmos.CosmosChain, keyName string, blocking bool, command ...string) (string, error) {
	node := chain.Validators[0]

	resp, err := node.ExecTx(ctx, keyName, command...)
	s.Require().NoError(err)

	if !blocking {
		return resp, nil
	}

	height, err := chain.Height(context.Background())
	s.Require().NoError(err)
	s.WaitForHeight(chain, height+1)

	stdout, stderr, err := chain.FullNodes[0].ExecQuery(ctx, "tx", resp, "--type", "hash")
	s.Require().NoError(err)
	s.Require().Nil(stderr)

	return string(stdout), nil
}

// CreateTx creates a new transaction to be signed by the given user, including a provided set of messages
func (s *TestSuite) CreateTx(chain *cosmos.CosmosChain, user cosmos.User, fee string, gas int64, msgs ...sdk.Msg) []byte {
	bc := cosmos.NewBroadcaster(s.T(), chain)

	ctx := context.Background()
	// create tx factory + Client Context
	txf, err := bc.GetFactory(ctx, user)
	s.Require().NoError(err)

	cc, err := bc.GetClientContext(ctx, user)
	s.Require().NoError(err)

	txf = txf.WithSimulateAndExecute(true)

	txf, err = txf.Prepare(cc)
	s.Require().NoError(err)

	// get gas for tx
	txf = txf.WithGas(uint64(gas))
	txf = txf.WithGasAdjustment(0)
	txf = txf.WithGasPrices("")
	txf = txf.WithFees(fee)

	// update sequence number
	txf = txf.WithSequence(txf.Sequence())

	// sign the tx
	txBuilder, err := txf.BuildUnsignedTx(msgs...)
	s.Require().NoError(err)
	s.Require().NoError(tx.Sign(cc.CmdContext, txf, cc.GetFromName(), txBuilder, true))

	// encode and return
	bz, err := cc.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	return bz
}

// SetOracleConfigsOnApp writes the oracle configuration to the given node's application config.
func SetOracleConfigsOnApp(node *cosmos.ChainNode) {
	oracle := GetOracleSideCar(node)

	// read the app config from the node
	bz, err := node.ReadFile(context.Background(), appConfigPath)
	if err != nil {
		panic(err)
	}

	// Unmarshall the app config to update the oracle and metrics file paths.
	var appConfig map[string]interface{}
	err = toml.Unmarshal(bz, &appConfig)
	if err != nil {
		panic(err)
	}

	oracleAppConfig, ok := appConfig["oracle"].(map[string]interface{})
	if !ok {
		panic("oracle config not found")
	}

	// Update the file paths to the oracle and metrics configs.
	oracleAppConfig["enabled"] = true
	oracleAppConfig["oracle_address"] = fmt.Sprintf("%s:%s", oracle.HostName(), "8080")
	oracleAppConfig["client_timeout"] = "1s"
	oracleAppConfig["metrics_enabled"] = true
	oracleAppConfig["prometheus_server_address"] = fmt.Sprintf("localhost:%s", "8081")

	appConfig["oracle"] = oracleAppConfig
	bz, err = toml.Marshal(appConfig)
	if err != nil {
		panic(err)
	}

	// Write back the app config.
	err = node.WriteFile(context.Background(), bz, appConfigPath)
	if err != nil {
		panic(err)
	}
}

// AddSidecarToNode adds the sidecar configured by the given config to the given node. These are configured
// so that the sidecar is started before the node is started.
func AddSidecarToNode(node *cosmos.ChainNode, conf ibc.SidecarConfig) {
	// create the sidecar process
	node.NewSidecarProcess(
		context.Background(),
		true,
		conf.ProcessName,
		node.DockerClient,
		node.NetworkID,
		conf.Image,
		conf.HomeDir,
		conf.Ports,
		conf.StartCmd,
		conf.Env,
	)
}

// SetOracleConfigsOnOracle writes the oracle and metrics configs to the given node's
// oracle sidecar.
func SetOracleConfigsOnOracle(
	oracle *cosmos.SidecarProcess,
	oracleCfg oracleconfig.OracleConfig,
) {
	// marshal the oracle config
	bz, err := json.Marshal(oracleCfg)
	if err != nil {
		panic(err)
	}

	// write the oracle config to the node
	err = oracle.WriteFile(context.Background(), bz, oracleConfigPath)
	if err != nil {
		panic(err)
	}
}

// RestartOracle restarts the oracle sidecar for a given node
func RestartOracle(node *cosmos.ChainNode) error {
	if len(node.Sidecars) != 1 {
		panic("expected node to have oracle sidecar")
	}

	oracle := node.Sidecars[0]

	if err := oracle.StopContainer(context.Background()); err != nil {
		return err
	}

	return oracle.StartContainer(context.Background())
}

// StopOracle stops the oracle sidecar for a given node
func StopOracle(node *cosmos.ChainNode) error {
	if len(node.Sidecars) != 1 {
		panic("expected node to have oracle sidecar")
	}

	oracle := node.Sidecars[0]

	return oracle.StopContainer(context.Background())
}

// StartOracle starts the oracle sidecar for a given node
func StartOracle(node *cosmos.ChainNode) error {
	if len(node.Sidecars) != 1 {
		panic("expected node to have oracle sidecar")
	}

	oracle := node.Sidecars[0]

	return oracle.StartContainer(context.Background())
}
