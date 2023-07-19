package keeper

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/noble/testutil/sample"
	cctpmoduletypes "github.com/strangelove-ventures/noble/x/cctp/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

func FiatTokenfactoryKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramsSubspace := typesparams.NewSubspace(cdc,
		codec.NewLegacyAmino(),
		storeKey,
		nil,
		"TokenfactoryParams",
	)
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		paramsSubspace,
		MockBankKeeper{},
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}

type MockFiatTokenfactoryKeeper struct{}

func (k MockFiatTokenfactoryKeeper) GetAuthority(ctx sdk.Context) (val cctpmoduletypes.Authority, found bool) {
	return cctpmoduletypes.Authority{Address: sample.AccAddress()}, true
}

func (MockFiatTokenfactoryKeeper) Mint(ctx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	return &types.MsgMintResponse{}, nil
}

func (MockFiatTokenfactoryKeeper) Burn(ctx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	return &types.MsgBurnResponse{}, nil
}

func (MockFiatTokenfactoryKeeper) GetMintingDenom(ctx sdk.Context) (val types.MintingDenom) {
	return types.MintingDenom{Denom: "uusdc"}
}
