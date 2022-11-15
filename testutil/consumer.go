package testutil

import (
	"time"

	ibctypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	ccvconsumertypes "github.com/cosmos/interchain-security/x/ccv/consumer/types"
	ccvprovidertypes "github.com/cosmos/interchain-security/x/ccv/provider/types"
)

// This function creates consumer module genesis state that is used as starting point for modifications
// that allow Noble chain to be started locally without having to start the provider chain and the relayer.
// It is also used in tests that are starting the chain node.
func CreateMinimalConsumerTestGenesis() *ccvconsumertypes.GenesisState {
	genesisState := ccvconsumertypes.DefaultGenesisState()
	genesisState.Params.Enabled = true
	genesisState.NewChain = true
	genesisState.ProviderClientState = ccvprovidertypes.DefaultParams().TemplateClient
	genesisState.ProviderClientState.ChainId = "noble"
	genesisState.ProviderClientState.LatestHeight = ibctypes.Height{RevisionNumber: 0, RevisionHeight: 1}
	genesisState.ProviderClientState.TrustingPeriod = genesisState.Params.UnbondingPeriod / ccvprovidertypes.DefaultTrustingPeriodFraction
	genesisState.ProviderClientState.UnbondingPeriod = genesisState.Params.UnbondingPeriod
	genesisState.ProviderClientState.MaxClockDrift = ccvprovidertypes.DefaultMaxClockDrift
	genesisState.ProviderConsensusState = &ibctmtypes.ConsensusState{
		Timestamp: time.Now().UTC(),
		Root:      types.MerkleRoot{Hash: []byte("dummy")},
	}

	return genesisState
}
