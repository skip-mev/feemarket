package integration_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"strconv"
	"testing"

	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/stretchr/testify/suite"

	"github.com/skip-mev/feemarket/testutils/networksuite"
)

// QueryTestSuite is a test suite for query tests
type NetworkTestSuite struct {
	networksuite.NetworkTestSuite
}

// TestQueryTestSuite runs test of the query suite
func TestNetworkTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkTestSuite))
}

func (suite *NetworkTestSuite) TestShowGenesisAccount() {
	ctx := suite.Network.Validators[0].ClientCtx

	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc       string
		idLaunchID string
		idAddress  string

		args []string
		err  error
		obj  types.GenesisAccount
	}{
		{
			desc:       "should show an existing genesis account",
			idLaunchID: strconv.Itoa(int(accs[0].LaunchID)),
			idAddress:  accs[0].Address,

			args: common,
			obj:  accs[0],
		},
		{
			desc:       "should send error for a non existing genesis account",
			idLaunchID: strconv.Itoa(100000),
			idAddress:  strconv.Itoa(100000),

			args: common,
			err:  status.Error(codes.NotFound, "not found"),
		},
	} {
		suite.T().Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idLaunchID,
				tc.idAddress,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowGenesisAccount(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetGenesisAccountResponse
				require.NoError(t, suite.Network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.GenesisAccount)
				require.Equal(t, tc.obj, resp.GenesisAccount)
			}
		})
	}
}
