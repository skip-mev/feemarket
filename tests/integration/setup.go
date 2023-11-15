package integration

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type KeyringOverride struct {
	keyringOptions keyring.Option
	cdc            codec.Codec
}

// ChainBuilderFromChainSpec creates an interchaintest chain builder factory given a ChainSpec
// and returns the associated chain
func ChainBuilderFromChainSpec(t *testing.T, spec *interchaintest.ChainSpec) ibc.Chain {
	// require that NumFullNodes == NumValidators == 4
	require.Equal(t, *spec.NumValidators, 1)

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
	client, networkID := interchaintest.DockerSetup(t)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// build the interchain
	err := ic.Build(ctx, nil, interchaintest.InterchainBuildOptions{
		SkipPathCreation: true,
		Client:           client,
		NetworkID:        networkID,
		TestName:         t.Name(),
	})
	require.NoError(t, err)

	return ic
}

// CreateTx creates a new transaction to be signed by the given user, including a provided set of messages
func (s *TestSuite) CreateTx(ctx context.Context, chain *cosmos.CosmosChain, user cosmos.User, seqIncrement, height uint64, GasPrice int64, msgs ...sdk.Msg) []byte {
	// create tx factory + Client Context
	txf, err := s.bc.GetFactory(ctx, user)
	s.Require().NoError(err)

	cc, err := s.bc.GetClientContext(ctx, user)
	s.Require().NoError(err)

	txf = txf.WithSimulateAndExecute(true)

	txf, err = txf.Prepare(cc)
	s.Require().NoError(err)

	// set timeout height
	if height != 0 {
		txf = txf.WithTimeoutHeight(height)
	}

	// get gas for tx
	txf.WithGas(25000000)

	// update sequence number
	txf = txf.WithSequence(txf.Sequence() + seqIncrement)
	txf = txf.WithGasPrices(sdk.NewDecCoins(sdk.NewDecCoin(chain.Config().Denom, math.NewInt(GasPrice))).String())

	// sign the tx
	txBuilder, err := txf.BuildUnsignedTx(msgs...)
	s.Require().NoError(err)

	s.Require().NoError(tx.Sign(txf, cc.GetFromName(), txBuilder, true))

	// encode and return
	bz, err := cc.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	return bz
}

// SimulateTx simulates the provided messages, and checks whether the provided failure condition is met
func (s *TestSuite) SimulateTx(ctx context.Context, chain *cosmos.CosmosChain, user cosmos.User, height uint64, expectFail bool, msgs ...sdk.Msg) {
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

type Tx struct {
	User               cosmos.User
	Msgs               []sdk.Msg
	GasPrice           int64
	SequenceIncrement  uint64
	Height             uint64
	SkipInclusionCheck bool
	ExpectFail         bool
}

// BroadcastTxs broadcasts the given messages for each user. This function returns the broadcasted txs. If a message
// is not expected to be included in a block, set SkipInclusionCheck to true and the method
// will not block on the tx's inclusion in a block, otherwise this method will block on the tx's inclusion
func (s *TestSuite) BroadcastTxs(ctx context.Context, chain *cosmos.CosmosChain, txs []Tx) [][]byte {
	return s.BroadcastTxsWithCallback(ctx, chain, txs, nil)
}

// BroadcastTxsWithCallback broadcasts the given messages for each user. This function returns the broadcasted txs. If a message
// is not expected to be included in a block, set SkipInclusionCheck to true and the method
// will not block on the tx's inclusion in a block, otherwise this method will block on the tx's inclusion. The callback
// function is called for each tx that is included in a block.
func (s *TestSuite) BroadcastTxsWithCallback(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	txs []Tx,
	cb func(tx []byte, resp *rpctypes.ResultTx),
) [][]byte {
	rawTxs := make([][]byte, len(txs))

	for i, msg := range txs {
		rawTxs[i] = s.CreateTx(ctx, chain, msg.User, msg.SequenceIncrement, msg.Height, msg.GasPrice, msg.Msgs...)
	}

	// broadcast each tx
	s.Require().True(len(chain.Nodes()) > 0)
	client := chain.Nodes()[0].Client

	statusResp, err := client.Status(context.Background())
	s.Require().NoError(err)

	s.T().Logf("broadcasting transactions at latest height of %d", statusResp.SyncInfo.LatestBlockHeight)

	for i, tx := range rawTxs {
		// broadcast tx
		resp, err := client.BroadcastTxSync(ctx, tx)

		// check execution was successful
		if !txs[i].ExpectFail {
			s.Require().Equal(resp.Code, uint32(0))
		} else {
			if resp != nil {
				s.Require().NotEqual(resp.Code, uint32(0))
			} else {
				s.Require().Error(err)
			}
		}
	}

	// block on all txs being included in block
	eg := errgroup.Group{}
	for i, tx := range rawTxs {
		// if we don't expect this tx to be included.. skip it
		if txs[i].SkipInclusionCheck || txs[i].ExpectFail {
			continue
		}

		tx := tx // pin
		eg.Go(func() error {
			return testutil.WaitForCondition(30*time.Second, 500*time.Millisecond, func() (bool, error) {
				res, err := client.Tx(context.Background(), comettypes.Tx(tx).Hash(), false)
				if err != nil || res.TxResult.Code != uint32(0) {
					return false, nil
				}

				if cb != nil {
					cb(tx, res)
				}

				return true, nil
			})
		})
	}

	s.Require().NoError(eg.Wait())

	return rawTxs
}

// QueryValidators queries for all the network's validators
func (s *TestSuite) QueryValidators(chain *cosmos.CosmosChain) []sdk.ValAddress {
	s.T().Helper()

	// get grpc client of the node
	grpcAddr := chain.GetHostGRPCAddress()
	cc, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	defer cc.Close()

	client := stakingtypes.NewQueryClient(cc)

	// query validators
	resp, err := client.Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{})
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
