package interchaintest_test

import (
	"testing"

	"github.com/strangelove-ventures/interchaintest/v4/ibc"
)

// run `make local-image`to rebuild updated binary before running test
func TestNoble1ChainUpgrade(t *testing.T) {
	const (
		noble1ChainID = "noble-1"
		numVals       = 4
		numFullNodes  = 0
	)

	noble1Genesis := ibc.DockerImage{
		Repository: "ghcr.io/strangelove-ventures/noble",
		Version:    "v1.0.0",
		UidGid:     containerUidGid,
	}

	noble1Upgrades := []chainUpgrade{
		{
			upgradeName: "neon",
			image: ibc.DockerImage{
				Repository: "ghcr.io/strangelove-ventures/noble",
				Version:    "v2.0.0",
				UidGid:     containerUidGid,
			},
		},
		{
			// omitting upgradeName due to huckleberry patch
			image: ibc.DockerImage{
				Repository: "ghcr.io/strangelove-ventures/noble",
				Version:    "v2.0.1",
				UidGid:     containerUidGid,
			},
		},
		{
			upgradeName: "radon",
			image: ibc.DockerImage{
				Repository: "ghcr.io/strangelove-ventures/noble",
				Version:    "v3.0.0",
				UidGid:     containerUidGid,
			},
			postUpgrade: testPostRadonUpgrade,
		},
		{
			upgradeName: "argon",
			image:       nobleImageInfo[0],
		},
	}

	testNobleChainUpgrade(t, noble1ChainID, noble1Genesis, denomMetadataFrienzies, numVals, numFullNodes, noble1Upgrades)
}
