package types

import (
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/cosmos/ibc-go/v4/modules/light-clients/07-tendermint/types"
)

func ParseChain(rawClientState exported.ClientState) string {
	switch clientState := rawClientState.(type) {
	case *types.ClientState:
		return clientState.ChainId
	default:
		return "UNKNOWN"
	}
}
