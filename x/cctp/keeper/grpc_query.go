package keeper

import (
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

var _ types.QueryServer = Keeper{}
