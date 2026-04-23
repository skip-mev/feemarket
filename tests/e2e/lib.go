package e2e

import (
	"context"
	"testing"
	"time"

	interchaintest "github.com/cosmos/interchaintest/v11"
	"github.com/cosmos/interchaintest/v11/chain/cosmos"
	"github.com/cosmos/interchaintest/v11/ibc"
	"github.com/cosmos/interchaintest/v11/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// ChainConstructor returns the primary test chain and any additional chains. The first chain is the
// feemarket chain under test.
type ChainConstructor func(t *testing.T, spec *interchaintest.ChainSpec, gasPrices string) []*cosmos.CosmosChain

// Interchain captures relayer / IBC-path information when tests use multiple linked chains.
type Interchain interface {
	Relayer() ibc.Relayer
	Reporter() *testreporter.RelayerExecReporter
	IBCPath() string
}

// InterchainConstructor returns an interchain that will be used in the feemarket integration tests.
// The chains used in the interchain constructor should be the chains constructed via the ChainConstructor
type InterchainConstructor func(ctx context.Context, t *testing.T, chains []*cosmos.CosmosChain) Interchain

// DefaultChainConstructor is the default construct of a chan that will be used in the feemarket
// integration tests. There is only a single chain that is created.
func DefaultChainConstructor(t *testing.T, spec *interchaintest.ChainSpec, _ string) []*cosmos.CosmosChain {
	// require that NumFullNodes == NumValidators == 4
	require.Equal(t, 4, *spec.NumValidators)

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{spec})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	// require that the chain is a cosmos chain
	require.Len(t, chains, 1)
	chain := chains[0]

	cosmosChain, ok := chain.(*cosmos.CosmosChain)
	require.True(t, ok)

	return []*cosmos.CosmosChain{cosmosChain}
}

// DefaultInterchainConstructor is the default constructor of an interchain that will be used in the feemarket integration tests.
func DefaultInterchainConstructor(ctx context.Context, t *testing.T, chains []*cosmos.CosmosChain) Interchain {
	require.Len(t, chains, 1)

	ic := interchaintest.NewInterchain()
	ic.AddChain(chains[0])

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

	return nil
}
