package keeper

import (
	"github.com/strangelove-ventures/noble/x/router/types"
)

var _ types.QueryServer = Keeper{}
