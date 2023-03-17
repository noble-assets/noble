package keeper

import (
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
)

var _ types.QueryServer = Keeper{}
