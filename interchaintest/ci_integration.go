package interchaintest

import (
	"os"
)

// The remote runner sets the BRANCH_CI env var. If present, interchaintest
// will use the docker image pushed up to repo.
// If testing locally, user should run `make local-image` and interchaintest
// will use the local image.
func GetDockerImageInfo() (repo, version string) {
	branchVersion, found := os.LookupEnv("BRANCH_CI")
	repo = "ghcr.io/strangelove-ventures/noble"
	if !found {
		repo = "noble"
		branchVersion = "local"
	}
	return repo, branchVersion
}
