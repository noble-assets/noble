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

func networkWithPublicKeyObjects(t *testing.T, n int) (*network.Network, []types.PublicKeys) {
	t.Helper()
	cfg := network.DefaultConfig()
	state := types.GenesisState{}
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &state))

	for i := 0; i < n; i++ {
		publicKeys := types.PublicKeys{
			Key: strconv.Itoa(i),
		}
		nullify.Fill(&publicKeys)
		state.PublicKeysList = append(state.PublicKeysList, publicKeys)
	}
	buf, err := cfg.Codec.MarshalJSON(&state)
	require.NoError(t, err)
	cfg.GenesisState[types.ModuleName] = buf
	return network.New(t, cfg), state.PublicKeysList
}

func TestShowPublicKey(t *testing.T) {
	net, objs := networkWithPublicKeyObjects(t, 2)

	ctx := net.Validators[0].ClientCtx
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc  string
		idKey string

		args []string
		err  error
		obj  types.PublicKeys
	}{
		{
			desc:  "found",
			idKey: objs[0].Key,
			args:  common,
			obj:   objs[0],
		},
		{
			desc:  "not found",
			idKey: "123",
			args:  common,
			err:   status.Error(codes.NotFound, "not found"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args := []string{
				tc.idKey,
			}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowPublicKey(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
			} else {
				require.NoError(t, err)
				var resp types.QueryGetPublicKeyResponse
				require.NoError(t, net.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				require.NotNil(t, resp.PublicKey)
				require.Equal(t,
					nullify.Fill(&tc.obj),
					nullify.Fill(&resp.PublicKey),
				)
			}
		})
	}
}
