package interchaintest_test

import (
	"testing"
)

// run `make local-image`to rebuild updated binary before running test
func TestGrand1ChainUpgrade(t *testing.T) {

	const (
		grand1ChainID = "grand-1"
		numVals       = 4
		numFullNodes  = 0
	)

	var grand1Genesis = ghcrImage("v0.3.0")

	var grand1Upgrades = []chainUpgrade{
		{
			// The upgrade was registered on-chain with name "v0.4.1" accidentally,
			// when "neon" was the upgrade name in the v0.4.1 code.
			// As such, v0.4.2 was required to complete the upgrade, which changed the upgrade
			// name in the code to "v0.4.1" as a workaround.
			upgradeName: "v0.4.1",
			image:       ghcrImage("dan-neon-control-test"), //this is an adjusted version that gives us control of the fiat-tokenfactory owner
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
			// This upgrade is only relevant to the grand-1 testnet
			upgradeName: "argon2",
			image:       nobleImageInfo[0],
			postUpgrade: testPostArgonUpgradeTestnet,
		},
	}

	testNobleChainUpgrade(t, grand1ChainID, grand1Genesis, denomMetadataDrachma, numVals, numFullNodes, grand1Upgrades)
}
