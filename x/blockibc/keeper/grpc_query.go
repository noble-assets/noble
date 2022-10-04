package keeper

import (
	"noble/x/blockibc/types"
)

var _ types.QueryServer = Keeper{}
