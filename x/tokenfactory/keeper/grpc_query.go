package keeper

import (
	"github.com/noble-assets/noble/v7/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}
