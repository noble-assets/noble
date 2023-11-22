package cli_test

import (
	"fmt"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/status"

	"github.com/noble-assets/noble/v4/testutil/network"
	"github.com/noble-assets/noble/v4/testutil/nullify"
	"github.com/noble-assets/noble/v4/x/tokenfactory/client/cli"
	"github.com/noble-assets/noble/v4/x/tokenfactory/types"
)

func networkWithMasterMinterObjects(t *testing.T) (*network.Network, types.MasterMinter) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	masterMinter := &types.MasterMinter{}
	nullify.Fill(&masterMinter)
	state.MasterMinter = masterMinter
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), *state.MasterMinter
}

func TestShowMasterMinter(t *testing.T) {
	net, obj := networkWithMasterMinterObjects(t)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc string
		args []string
		err  error
		obj  types.MasterMinter
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
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowMasterMinter(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetMasterMinterResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.MasterMinter)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.MasterMinter),
				)
			}
		})
	}
}
