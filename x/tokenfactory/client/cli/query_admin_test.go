package cli_test

import (
	"fmt"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/status"

	"noble/testutil/network"
	"noble/testutil/nullify"
	"noble/x/tokenfactory/client/cli"
	"noble/x/tokenfactory/types"
)

func networkWithAdminObjects(t *testing.T) (*network.Network, types.Admin) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	admin := &types.Admin{}
	nullify.Fill(&admin)
	state.Admin = admin
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), *state.Admin
}

func TestShowAdmin(t *testing.T) {
	net, obj := networkWithAdminObjects(t)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc string
		args []string
		err  error
		obj  types.Admin
	}{
		{
			desc: "get",
			args: common,
			obj:  obj,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			var args []string
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowAdmin(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetAdminResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.Admin)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.Admin),
				)
			}
		})
	}
}
