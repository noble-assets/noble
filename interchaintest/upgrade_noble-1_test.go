package interchaintest_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	tarifftypes "github.com/noble-assets/noble/v6/x/tariff/types"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
)

type QueryValidatorsResponse struct {
	Validators []Validator `json:"validators"`
}
type Validator struct {
	Jailed bool `json:"jailed"`
}

// run `make local-image`to rebuild updated binary before running test
func TestNoble1ChainUpgrade(t *testing.T) {
	const (
		numValidators = 4
		numFullNodes  = 0
	)

	genesis := ghcrImage("v1.0.0")

	upgrades := []chainUpgrade{
		{
			upgradeName: "neon",
			// this is a mock image that gives us control of the
			// fiat-tokenfactory owner for testing purposes (postUpgrade tests)
			image: ghcrImage("mock-v2.0.0"),
		},
		{
			// omitting upgradeName due to huckleberry patch
			image: ghcrImage("v2.0.1"),
		},
		{
			upgradeName: "radon",
			image:       ghcrImage("v3.0.0"),
			postUpgrade: testPostRadonUpgrade,
		},
		{
			upgradeName: "v3.1.0",
			image:       ghcrImage("v3.1.0"),
		},
		{
			upgradeName: "argon",
			// this is a mock image that gives us control of the
			// cctp owner for testing purposes (postUpgrade tests)
			// (this is not needed for the `upgrade_grand-1_test` because
			// the v4.0.0-alpha1 upgrade handler was only run in the testnet
			// making the cctp owner the same as the paramauthority. This is
			// not the case in mainnet; the cctp owner is a separate wallet)
			image:       ghcrImage("mock-v4.0.0"),
			postUpgrade: testPostArgonUpgrade,
		},
		{
			// v4.0.1 is a patch release to fix a consensus failure caused by a validator going offline.
			// The preUpgrade logic replicates this failure by bringing one validator offline.
			// The postUpgrade logic verifies that after applying the emergency upgrade, the offline validator is jailed.
			emergency: true,
			image:     ghcrImage("v4.0.1"),
			preUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, paramAuthority ibc.Wallet) {
				// Select one validator to go offline.
				validator := noble.Validators[numValidators-1]

				// Take the selected validator offline.
				require.NoError(t, validator.StopContainer(ctx))

				// Wait 5 blocks (+1) to exceed the downtime window.
				timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, 42*time.Second)
				defer timeoutCtxCancel()

				_ = testutil.WaitForBlocks(timeoutCtx, 6, noble)
			},
			postUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, paramAuthority ibc.Wallet) {
				raw, _, err := noble.Validators[0].ExecQuery(ctx, "staking", "validators")
				require.NoError(t, err)

				var res QueryValidatorsResponse
				require.NoError(t, json.Unmarshal(raw, &res))

				numJailed := 0
				for _, validator := range res.Validators {
					if validator.Jailed {
						numJailed += 1
					}
				}

				require.Equal(t, numJailed, 1)
			},
		},
		{
			// v4.0.2 is a patch release that introduces a new query to the tariff module.
			// It is non-consensus breaking, and therefore is applied as a rolling upgrade.
			image: ghcrImage("v4.0.2"),
			postUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, paramAuthority ibc.Wallet) {
				raw, _, err := noble.Validators[0].ExecQuery(ctx, "tariff", "params")
				require.NoError(t, err)

				var res tarifftypes.QueryParamsResponse
				require.NoError(t, json.Unmarshal(raw, &res))
			},
		},
		{
			// v4.0.3 is a patch release that upgraded two core dependencies.
			// It is consensus breaking, and therefore is applied as an emergency upgrade.
			emergency: true,
			image:     ghcrImage("v4.0.3"),
		},
		{
			// fusion is a minor release to the v4 argon line, that introduced a new forwarding module.
			// v4.1.0 was retracted due to a consensus failure, and so we use v4.1.1.
			upgradeName: "fusion",
			image:       ghcrImage("v4.1.1"),
		},
		{
			// v4.1.2 is a patch release that upgraded one core dependency.
			// It is consensus breaking, and therefore is applied as an emergency upgrade.
			emergency: true,
			image:     ghcrImage("v4.1.2"),
		},
		{
			// v4.1.3 is a patch release that upgraded one core dependency.
			// It is consensus breaking, and therefore is applied as an emergency upgrade.
			emergency: true,
			image:     ghcrImage("v4.1.3"),
		},
		{
			// krypton is a major release that introduced the aura module.
			upgradeName: "krypton",
			image:       nobleImageInfo[0],
		},
	}

	testNobleChainUpgrade(t, "noble-1", genesis, denomMetadataFrienzies, numValidators, numFullNodes, upgrades)
}
