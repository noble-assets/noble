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
	}

	testNobleChainUpgrade(t, noble1ChainID, noble1Genesis, denomMetadataFrienzies, numVals, numFullNodes, noble1Upgrades)
}
