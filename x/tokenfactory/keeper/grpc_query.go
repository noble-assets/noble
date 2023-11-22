package keeper

import (
	"github.com/noble-assets/noble/v4/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}
