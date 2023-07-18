package cli_test

import (
	"fmt"
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

func networkWithSendingAndReceivingMessagesPausedObjects(t *testing.T) (*network.Network, types.SendingAndReceivingMessagesPaused) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	SendingAndReceivingMessagesPaused := &types.SendingAndReceivingMessagesPaused{}
	nullify.Fill(&SendingAndReceivingMessagesPaused)
	state.SendingAndReceivingMessagesPaused = SendingAndReceivingMessagesPaused
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), *state.SendingAndReceivingMessagesPaused
}

func TestShowSendingAndReceivingMessagesPaused(t *testing.T) {
	net, obj := networkWithSendingAndReceivingMessagesPausedObjects(t)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc string
		args []string
		err  error
		obj  types.SendingAndReceivingMessagesPaused
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
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowSendingAndReceivingMessagesPaused(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetSendingAndReceivingMessagesPausedResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.Paused)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.Paused),
				)
			}
		})
	}
}
