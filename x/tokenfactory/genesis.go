package tokenfactory

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/v3/x/tokenfactory/keeper"
	"github.com/strangelove-ventures/noble/v3/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, bankKeeper types.BankKeeper, genState types.GenesisState) {
	for _, elem := range genState.BlacklistedList {
		k.SetBlacklisted(ctx, elem)
	}

	if genState.Paused != nil {
		k.SetPaused(ctx, *genState.Paused)
	}

	if genState.MasterMinter != nil {
		k.SetMasterMinter(ctx, *genState.MasterMinter)
	}

	for _, elem := range genState.MintersList {
		k.SetMinters(ctx, elem)
	}

	if genState.Pauser != nil {
		k.SetPauser(ctx, *genState.Pauser)
	}

	if genState.Blacklister != nil {
		k.SetBlacklister(ctx, *genState.Blacklister)
	}

	if genState.Owner != nil {
		k.SetOwner(ctx, *genState.Owner)
	}

	for _, elem := range genState.MinterControllerList {
		k.SetMinterController(ctx, elem)
	}

	if genState.MintingDenom != nil {
		_, found := bankKeeper.GetDenomMetaData(ctx, genState.MintingDenom.Denom)
		if !found {
			panic(sdkerrors.Wrapf(types.ErrDenomNotRegistered, "tokenfactory minting denom %s is not registered in bank module denom_metadata", genState.MintingDenom.Denom))
		}
		k.SetMintingDenom(ctx, *genState.MintingDenom)
	}
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported GenesisState
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.BlacklistedList = k.GetAllBlacklisted(ctx)

	paused := k.GetPaused(ctx)
	genesis.Paused = &paused

	masterMinter, found := k.GetMasterMinter(ctx)
	if found {
		genesis.MasterMinter = &masterMinter
	}
	genesis.MintersList = k.GetAllMinters(ctx)

	pauser, found := k.GetPauser(ctx)
	if found {
		genesis.Pauser = &pauser
	}

	blacklister, found := k.GetBlacklister(ctx)
	if found {
		genesis.Blacklister = &blacklister
	}

	owner, found := k.GetOwner(ctx)
	if found {
		genesis.Owner = &owner
	}
	genesis.MinterControllerList = k.GetAllMinterControllers(ctx)

	mintingDenom := k.GetMintingDenom(ctx)
	genesis.MintingDenom = &mintingDenom
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
