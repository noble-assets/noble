package e2e_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	"github.com/noble-assets/noble/e2e"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
)

// TestAddressCompatibility ensures compatibility with `penumbra` and `penumbracompat`
// bech32m addresses. Specifically, this is necessary for the fiat-tokenfactory ante handler.
// Since Penumbra is not fully supported in Interchaintest, the Penumbra addresses have been pre-generated.
//
// The penumbra addresses created in this test use this mnemonic:
// approve bracket canyon yard such jungle patch decade monster scissors burden gold stone essay shield scatter net dynamic salad umbrella play trophy lake blossom
func TestAddressCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	nw, _, _, r, _, _, eRep, _, _ := e2e.NobleSpinUpIBC(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	sender := interchaintest.GetAndFundTestUsers(t, ctx, "wallet", math.OneInt(), noble)[0]

	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", sender.FormattedAddress(), "100uusdc")
	require.NoError(t, err)

	channels, err := r.GetChannels(ctx, eRep, noble.Config().ChainID)
	require.NoError(t, err)

	// test normal penumbra address
	penumbra := "penumbra1ld2kghffzgwq4597ejpgmnwxa7ju0cndytuxtsjh8qhjyfuwq0rwd5flnw4a3fgclw7m5puh50nskn2c88flhne2hzchnpxru609d5wgmqqvhdf0sy2tktqfcm2p2tmxceqwvv"

	sendAmount := math.NewInt(10)
	_, err = val.SendIBCTransfer(ctx, channels[0].ChannelID, sender.KeyName(), ibc.WalletAmount{
		Address: penumbra,
		Denom:   "uusdc",
		Amount:  sendAmount,
	},
		ibc.TransferOptions{},
	)
	require.NoError(t, err)

	// test "compat" penumbra address
	penumbracompat := "penumbracompat11ld2kghffzgwq4597ejpgmnwxa7ju0cndytuxtsjh8qhjyfuwq0rwd5flnw4a3fgclw7m5puh50nskn2c88flhne2hzchnpxru609d5wgmqqvhdf0sy2tktqfcm2p2tmxeuc86n"

	_, err = val.SendIBCTransfer(ctx, channels[0].ChannelID, sender.KeyName(), ibc.WalletAmount{
		Address: penumbracompat,
		Denom:   "uusdc",
		Amount:  sendAmount,
	},
		ibc.TransferOptions{},
	)
	require.NoError(t, err)

	// test invalid penumbra address
	invalidPenumbra := "penumbra1ld2kghffzgwq4597ejpgmnwxa7ju0cndytuxtsjh8qhjyfuwq0rwd5flnw4a3fgclw7m5puh50nskn2c88flhne2hzchnpxru609d5wgmqqvhdf0sy2tktqfcm2p2toooooooooinvalid"

	_, err = val.SendIBCTransfer(ctx, channels[0].ChannelID, sender.KeyName(), ibc.WalletAmount{
		Address: invalidPenumbra,
		Denom:   "uusdc",
		Amount:  sendAmount,
	},
		ibc.TransferOptions{},
	)
	require.ErrorContains(t, err, "error decoding address")
}
