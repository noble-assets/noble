package keeper

import (
	"github.com/noble-assets/noble/v5/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}
