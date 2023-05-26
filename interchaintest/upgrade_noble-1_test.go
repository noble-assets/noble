package interchaintest_test

import (
	"testing"

	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	integration "github.com/strangelove-ventures/noble/interchaintest"
)

func TestNoble1ChainUpgrade(t *testing.T) {
	repo, version := integration.GetDockerImageInfo()

	const (
		noble1ChainID = "noble-1"
		numVals       = 4
		numFullNodes  = 0
	)

	var noble1Genesis = ibc.DockerImage{
		Repository: "ghcr.io/strangelove-ventures/noble",
		Version:    "v1.0.0",
		UidGid:     containerUidGid,
	}

	var noble1Upgrades = []chainUpgrade{
		{
			upgradeName: "neon",
			image: ibc.DockerImage{
				Repository: "ghcr.io/strangelove-ventures/noble",
				Version:    "v2.0.1",
				UidGid:     containerUidGid,
			},
		},
		{
			upgradeName: "radon",
			image: ibc.DockerImage{
				Repository: repo,
				Version:    version,
				UidGid:     containerUidGid,
			},
			postUpgrade: testPostRadonUpgrade,
		},
	}

	testNobleChainUpgrade(t, noble1ChainID, noble1Genesis, denomMetadataFrienzies, numVals, numFullNodes, noble1Upgrades)
}
