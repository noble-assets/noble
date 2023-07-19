package cli_test

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"strconv"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/status"

	"github.com/strangelove-ventures/noble/testutil/network"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/cctp/client/cli"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func networkWithTokenPairObjects(t *testing.T, n int) (*network.Network, []types.TokenPairs) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < n; i++ {
		tokenPair := types.TokenPairs{
			RemoteDomain: uint32(i),
			RemoteToken:  strconv.Itoa(i),
			LocalToken:   strconv.Itoa(i),
		}
		nullify.Fill(&tokenPair)
		state.TokenPairList = append(state.TokenPairList, tokenPair)
	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), state.TokenPairList
}

func TestShowTokenPair(t *testing.T) {
	net, objs := networkWithTokenPairObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc         string
		remoteDomain string
		remoteToken  string

		args []string
		err  error
		obj  types.TokenPairs
	}{
		{
			desc:         "found",
			remoteDomain: strconv.Itoa(int(objs[0].RemoteDomain)),
			remoteToken:  objs[0].RemoteToken,
			args:         common,
			obj:          objs[0],
		},
		{
			desc:         "not found",
			remoteDomain: "notakey",
			remoteToken:  objs[0].RemoteToken,
			args:         common,
			err:          status.Error(codes.NotFound, "not found"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.remoteDomain,
				tc.remoteToken,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowTokenPair(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetTokenPairResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.Pair.RemoteDomain)
				require.NotNil(t, resp.Pair.RemoteToken)
				require.NotNil(t, resp.Pair.LocalToken)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.Pair),
				)
			}
		})
	}
}
