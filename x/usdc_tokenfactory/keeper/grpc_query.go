package keeper

import (
	"github.com/strangelove-ventures/noble/x/usdc_tokenfactory/types"
)

var _ types.QueryServer = Keeper{}
