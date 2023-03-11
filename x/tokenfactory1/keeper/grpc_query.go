package keeper

import (
	"github.com/strangelove-ventures/noble/x/tokenfactory1/types"
)

var _ types.QueryServer = Keeper{}
