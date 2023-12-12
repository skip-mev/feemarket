package integration_test

import (
	"fmt"
	"testing"

	tmcli "github.com/cometbft/cometbft/libs/cli"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/status"

	"github.com/skip-mev/feemarket/testutils/networksuite"
	"github.com/skip-mev/feemarket/x/feemarket/client/cli"
	"github.com/skip-mev/feemarket/x/feemarket/types"
)

// NetworkTestSuite is a test suite for network integration tests.
type NetworkTestSuite struct {
	networksuite.NetworkTestSuite
}

// TestQueryTestSuite runs test of network integration tests.
func TestNetworkTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkTestSuite))
}

func (s *NetworkTestSuite) TestGetParams() {
	s.T().Parallel()

	ctx := s.Network.Validators[0].ClientCtx

	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		name string

		args []string
		err  error
		obj  types.Params
	}{
		{
			name: "should return default params",
			args: common,
			obj:  types.DefaultParams(),
		},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			tc := tc
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.GetParamsCmd(), tc.args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.ParamsResponse
				require.NoError(t, s.Network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp.Params))
				require.NotNil(t, resp.Params)
				require.Equal(t, tc.obj, resp.Params)
			}
		})
	}
}

func (s *NetworkTestSuite) TestGetState() {
	s.T().Parallel()

	ctx := s.Network.Validators[0].ClientCtx

	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		name string

		args []string
		err  error
		obj  types.State
	}{
		{
			name: "should return default state",
			args: common,
			obj:  types.DefaultState(),
		},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			tc := tc
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.GetStateCmd(), tc.args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.StateResponse
				require.NoError(t, s.Network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp.State))
				require.NotNil(t, resp.State)
				require.Equal(t, tc.obj, resp.State)
			}
		})
	}
}

func (s *NetworkTestSuite) TestSpamTx() {
	s.T().Parallel()

	ctx := s.Network.Validators[0].ClientCtx

	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		name string

		args []string
		err  error
		obj  types.State
	}{
		{
			name: "should return default state",
			args: common,
			obj:  types.DefaultState(),
		},
	} {
		// TODO SPAM TX

		s.T().Run(tc.name, func(t *testing.T) {
			tc := tc
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.GetStateCmd(), tc.args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.StateResponse
				require.NoError(t, s.Network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp.State))
				require.NotNil(t, resp.State)
				require.Equal(t, tc.obj, resp.State)
			}
		})
	}
}
