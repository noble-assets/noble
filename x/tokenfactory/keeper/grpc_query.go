package keeper

import (
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}
