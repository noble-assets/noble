package interchaintest_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
)

// run `make local-image`to rebuild updated binary before running test
func TestGrand1ChainUpgrade(t *testing.T) {
	const (
		numValidators = 4
		numFullNodes  = 0
	)

	genesis := ghcrImage("v0.3.0")

	upgrades := []chainUpgrade{
		{
			// The upgrade was registered on-chain with name "v0.4.1" accidentally,
			// when "neon" was the upgrade name in the v0.4.1 code.
			// As such, v0.4.2 was required to complete the upgrade, which changed the upgrade
			// name in the code to "v0.4.1" as a workaround.
			upgradeName: "v0.4.1",
			// this is a mock image that gives us control of the
			// fiat-tokenfactory owner for testing purposes (postUpgrade tests)
			image: ghcrImage("mock-v0.4.2"),
		},
		{
			upgradeName: "radon",
			image:       ghcrImage("v0.5.1"), // testnet actually upgraded to v0.5.0, but that required a hack to fix the consensus min fee. v0.5.1 fixes that
			postUpgrade: testPostRadonUpgrade,
		},
		{
			// post radon patch upgrade (will be applied as rolling upgrade due to lack of upgradeName)
			image: ghcrImage("v3.0.0"),
		},
		{
			upgradeName: "argon",
			image:       ghcrImage("v4.0.0-alpha1"),
		},
		{
			// post argon patch upgrade (will be applied as rolling upgrade due to lack of upgradeName)
			// This upgrade is only relevant to the grand-1 testnet
			image: ghcrImage("v4.0.0-alpha2"),
		},
		{
			// This upgrade is only relevant to the grand-1 testnet
			upgradeName: "argon2",
			image:       ghcrImage("v4.0.0-alpha3"),
		},
		{
			// This upgrade is only relevant to the grand-1 testnet
			upgradeName: "argon3",
			image:       ghcrImage("v4.0.0-beta1"),
		},
		{
			// This upgrade is only relevant to the grand-1 testnet
			upgradeName: "v4.0.0-beta2",
			image:       ghcrImage("v4.0.0-beta2"),
		},
		{
			upgradeName: "v4.0.0-rc0",
			image:       ghcrImage("v4.0.0-rc0"),
			postUpgrade: testPostArgonUpgrade,
		},
		{
			upgradeName: "v4.1.0-rc.0",
			image:       ghcrImage("v4.1.0-rc.0"),
		},
		{
			// v4.1.0-rc.1 is a patch release to fix a consensus failure caused by a validator going offline.
			// The preUpgrade logic replicates this failure by bringing one validator offline.
			// The postUpgrade logic verifies that after applying the emergency upgrade, the offline validator is jailed.
			emergency: true,
			image:     ghcrImage("v4.1.0-rc.1"),
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
			// v4.1.0-rc.2 is a new release candidate that introduced a new
			// forwarding module, among other changes.
			upgradeName: "v4.1.0-rc.2",
			image:       ghcrImage("v4.1.0-rc.2"),
		},
		{
			// v4.1.0-rc.3 is a patch release that upgraded two core dependencies.
			// It is consensus breaking, and therefore is applied as an emergency upgrade.
			emergency: true,
			image:     ghcrImage("v4.1.0-rc.3"),
		},
		{
			// fusion is a new release candidate that introduced audit fixes.
			upgradeName: "fusion",
			image:       ghcrImage("v4.1.0-rc.4"),
		},
		{
			// v4.1.1 is a patch release that resolved a consensus failure.
			// It is consensus breaking, and therefore is applied as an emergency upgrade.
			emergency: true,
			image:     ghcrImage("v4.1.1"),
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
			image:       ghcrImage("v5.0.0-rc.0"),
		},
		{
			// xenon is a major release that introduced the halo module.
			upgradeName: "xenon",
			image:       ghcrImage("v6.0.0-rc.0"),
		},
		{
			// numus is a major release that introduced the florin module.
			upgradeName: "numus",
			image:       nobleImageInfo[0],
		},
	}

	testNobleChainUpgrade(t, "grand-1", genesis, denomMetadataUsdc, numValidators, numFullNodes, upgrades)
}
