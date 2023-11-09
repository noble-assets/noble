package keeper

import (
	"github.com/strangelove-ventures/noble/v4/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}
