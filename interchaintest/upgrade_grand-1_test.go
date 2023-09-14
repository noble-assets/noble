package interchaintest_test

import (
	"testing"

	"github.com/strangelove-ventures/interchaintest/v3/ibc"
)

// run `make local-image`to rebuild updated binary before running test
func TestGrand1ChainUpgrade(t *testing.T) {

	const (
		grand1ChainID = "grand-1"
		numVals       = 4
		numFullNodes  = 0
	)

	var grand1Genesis = ibc.DockerImage{
		Repository: "ghcr.io/strangelove-ventures/noble",
		Version:    "v0.3.0",
		UidGid:     containerUidGid,
	}

	var grand1Upgrades = []chainUpgrade{
		{
			// The upgrade was registered on-chain with name "v0.4.1" accidentally,
			// when "neon" was the upgrade name in the v0.4.1 code.
			// As such, v0.4.2 was required to complete the upgrade, which changed the upgrade
			// name in the code to "v0.4.1" as a workaround.
			upgradeName: "v0.4.1",
			image: ibc.DockerImage{
				Repository: "ghcr.io/strangelove-ventures/noble",
				Version:    "v0.4.2",
				UidGid:     containerUidGid,
			},
		},
		{
			upgradeName: "radon",
			image: ibc.DockerImage{
				Repository: "ghcr.io/strangelove-ventures/noble",
				// testnet actually upgraded to v0.5.0, but that required a hack to fix the consensus min fee. v0.5.1 fixes that.
				Version: "v0.5.1",
				UidGid:  containerUidGid,
			},
			postUpgrade: testPostRadonUpgrade,
		},
		{
			// post radon patch upgrade (will be applied as rolling upgrade due to lack of upgradeName)
			image: ibc.DockerImage{
				Repository: "ghcr.io/strangelove-ventures/noble",
				Version:    "v3.0.0",
				UidGid:     containerUidGid,
			},
		},
		{
			upgradeName: "v3.1.0",
			image:       nobleImageInfo[0],
		},
	}

	testNobleChainUpgrade(t, grand1ChainID, grand1Genesis, denomMetadataDrachma, numVals, numFullNodes, grand1Upgrades)
}
