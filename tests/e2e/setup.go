package e2e

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skip-mev/feemarket/x/feemarket/types"
)

type KeyringOverride struct {
	keyringOptions keyring.Option
	cdc            codec.Codec
}

// ChainBuilderFromChainSpec creates an interchaintest chain builder factory given a ChainSpec
// and returns the associated chain
func ChainBuilderFromChainSpec(t *testing.T, spec *interchaintest.ChainSpec) ibc.Chain {
	// require that NumFullNodes == NumValidators == 3
	require.Equal(t, *spec.NumValidators, 3)

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{spec})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	require.Len(t, chains, 1)
	chain := chains[0]

	_, ok := chain.(*cosmos.CosmosChain)
	require.True(t, ok)

	return chain
}

// BuildInterchain creates a new Interchain testing env with the configured Block SDK CosmosChain
func BuildInterchain(t *testing.T, ctx context.Context, chain ibc.Chain) *interchaintest.Interchain {
	ic := interchaintest.NewInterchain()
	ic.AddChain(chain)

	// create docker network
	dockerClient, networkID := interchaintest.DockerSetup(t)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// build the interchain
	err := ic.Build(ctx, nil, interchaintest.InterchainBuildOptions{
		SkipPathCreation: true,
		Client:           dockerClient,
		NetworkID:        networkID,
		TestName:         t.Name(),
	})
	require.NoError(t, err)

	return ic
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

	// cast chain to cosmos-chain
	cosmosChain, ok := s.chain.(*cosmos.CosmosChain)
	s.Require().True(ok)
	// get nodes
	nodes := cosmosChain.Nodes()
	s.Require().True(len(nodes) > 0)

	// make params query to first node
	resp, _, err := nodes[0].ExecQuery(context.Background(), "feemarket", "params")
	s.Require().NoError(err)

	// unmarshal params
	var params types.Params
	err = s.cdc.UnmarshalJSON(resp, &params)
	s.Require().NoError(err)
	return params
}

func (s *TestSuite) QueryState() types.State {
	s.T().Helper()

	// cast chain to cosmos-chain
	cosmosChain, ok := s.chain.(*cosmos.CosmosChain)
	s.Require().True(ok)
	// get nodes
	nodes := cosmosChain.Nodes()
	s.Require().True(len(nodes) > 0)

	// make params query to first node
	resp, _, err := nodes[0].ExecQuery(context.Background(), "feemarket", "state")
	s.Require().NoError(err)

	// unmarshal state
	var state types.State
	err = s.cdc.UnmarshalJSON(resp, &state)
	s.Require().NoError(err)
	return state
}

func (s *TestSuite) QueryBaseFee() sdk.Coins {
	s.T().Helper()

	// cast chain to cosmos-chain
	cosmosChain, ok := s.chain.(*cosmos.CosmosChain)
	s.Require().True(ok)
	// get nodes
	nodes := cosmosChain.Nodes()
	s.Require().True(len(nodes) > 0)

	// make params query to first node
	resp, _, err := nodes[0].ExecQuery(context.Background(), "feemarket", "base-fee")
	s.Require().NoError(err)

	// unmarshal state
	fees, err := sdk.ParseCoinsNormalized(string(resp))
	s.Require().NoError(err)
	return fees
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

// QueryAccountSequence
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
func (s *TestSuite) WaitForHeight(chain *cosmos.CosmosChain, height uint64) {
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

	bc := cosmos.NewBroadcaster(s.T(), s.chain.(*cosmos.CosmosChain))

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
	node := s.chain.(*cosmos.CosmosChain).Nodes()[0]

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

// SendCoins creates a executes a SendCoins message and broadcasts the transaction.
func (s *TestSuite) SendCoins(ctx context.Context, chain *cosmos.CosmosChain, keyName, sender, receiver string, amt, fees sdk.Coins, gas int64) (string, error) {
	resp, err := s.ExecTx(
		ctx,
		chain,
		keyName,
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
	chain *cosmos.CosmosChain,
) (ibc.Wallet, error) {
	chainCfg := chain.Config()
	keyName := fmt.Sprintf("%s-%s-%s", keyNamePrefix, chainCfg.ChainID, RandLowerCaseLetterString(3))
	user, err := chain.BuildWallet(ctx, keyName, mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to get source user wallet: %w", err)
	}

	_, err = s.SendCoins(
		ctx,
		chain,
		interchaintest.FaucetAccountKeyName,
		interchaintest.FaucetAccountKeyName,
		user.FormattedAddress(),
		sdk.NewCoins(sdk.NewCoin(chainCfg.Denom, sdk.NewInt(amount))),
		sdk.NewCoins(sdk.NewCoin(chainCfg.Denom, sdk.NewInt(1000000000000))),
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
	chains ...*cosmos.CosmosChain,
) []ibc.Wallet {
	users := make([]ibc.Wallet, len(chains))
	var eg errgroup.Group
	for i, chain := range chains {
		i := i
		chain := chain
		eg.Go(func() error {
			user, err := s.GetAndFundTestUserWithMnemonic(ctx, keyNamePrefix, "", amount, chain)
			if err != nil {
				return err
			}
			users[i] = user
			return nil
		})
	}
	s.Require().NoError(eg.Wait())

	chainHeights := make([]testutil.ChainHeighter, len(chains))
	for i := range chains {
		chainHeights[i] = chains[i]
	}
	return users
}

// ExecTx executes a cli command on a node, waits a block and queries the Tx to verify it was included on chain.
func (s *TestSuite) ExecTx(ctx context.Context, chain *cosmos.CosmosChain, keyName string, command ...string) (string, error) {
	node := chain.FullNodes[0]

	resp, err := node.ExecTx(ctx, keyName, command...)
	s.Require().NoError(err)

	height, err := chain.Height(context.Background())
	s.Require().NoError(err)
	s.WaitForHeight(chain, height+1)

	stdout, stderr, err := chain.FullNodes[0].ExecQuery(ctx, "tx", resp, "--type", "hash")
	s.Require().NoError(err)
	s.Require().Nil(stderr)

	return string(stdout), nil
}

// RandLowerCaseLetterString returns a lowercase letter string of given length
func RandLowerCaseLetterString(length int) string {
	chars := []byte("abcdefghijklmnopqrstuvwxyz")

	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
