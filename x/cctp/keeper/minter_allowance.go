package keeper

import (
	"context"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

// Queries a MinterAllowance by index
func (k Keeper) MinterAllowance(context.Context, *types.QueryGetMinterAllowanceRequest) (*types.QueryGetMinterAllowanceResponse, error) {
	panic("not implemented")
}

// Queries a list of MinterAllowances
func (k Keeper) MinterAllowances(context.Context, *types.QueryAllMinterAllowancesRequest) (*types.QueryAllMinterAllowancesResponse, error) {
	panic("not implemented")
}

func (m msgServer) UpdateMinterAllowance(context.Context, *types.MsgUpdateMinterAllowance) (*types.MsgUpdateMinterAllowanceResponse, error) {
	panic("not implemented")
}
