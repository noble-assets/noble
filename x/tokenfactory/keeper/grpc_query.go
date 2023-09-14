package keeper

import (
	"github.com/strangelove-ventures/noble/v3/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}
