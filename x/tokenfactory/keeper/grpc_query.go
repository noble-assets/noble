package keeper

import (
	"github.com/noble-assets/noble/v6/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}
