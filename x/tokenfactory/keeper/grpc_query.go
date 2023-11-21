package keeper

import (
	"github.com/strangelove-ventures/noble/v5/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}
