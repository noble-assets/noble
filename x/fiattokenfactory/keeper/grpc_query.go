package keeper

import (
	"github.com/strangelove-ventures/noble/v3/x/fiattokenfactory/types"
)

var _ types.QueryServer = Keeper{}
