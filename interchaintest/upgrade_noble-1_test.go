package interchaintest_test

import (
	"testing"
)

// run `make local-image`to rebuild updated binary before running test
func TestNoble1ChainUpgrade(t *testing.T) {

	const (
		noble1ChainID = "noble-1"
		numVals       = 4
		numFullNodes  = 0
	)

	var noble1Genesis = ghcrImage("v1.0.0")

	var noble1Upgrades = []chainUpgrade{
		{
			upgradeName: "neon",
			image:       ghcrImage("v2.0.0"),
		},
		{
			// omitting upgradeName due to huckleberry patch
			image: ghcrImage("v2.0.1"),
		},
		{
			upgradeName: "radon",
			image:       ghcrImage("v3.0.0"),
		},
		{
			upgradeName: "v3.1.0",
			image:       ghcrImage("v3.1.0"),
			postUpgrade: testPostRadonUpgrade,
		},
		{
			upgradeName: "argon",
			image:       nobleImageInfo[0],
			// postUpgrade: testPostArgonUpgradeMainnet,
		},
	}

	testNobleChainUpgrade(t, noble1ChainID, noble1Genesis, denomMetadataFrienzies, numVals, numFullNodes, noble1Upgrades)
}
