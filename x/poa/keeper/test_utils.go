package keeper

import (
	"bytes"
	"encoding/hex"
	"strconv"

	"testing"

	"github.com/strangelove-ventures/noble/x/poa/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"

	dbm "github.com/tendermint/tm-db"
)

const (
	SamplePubKey  = "b7a3c12dc0c8c748ab07525b701122b88bd78f600c76342d27f25e5f92444cde"
	SamplePubKey2 = "b7a3c12dc0c8c748ab07525b701122b88bd78f600c76342d27f25e5f92444cdf"
)

// NewPubKeyFromHex returns a PubKey from a hex string.
func NewPubKeyFromHex(pk string) (res cryptotypes.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	if len(pkBytes) != ed25519.PubKeySize {
		panic(errors.Wrap(errors.ErrInvalidPubKey, "invalid pubkey size"))
	}
	return &ed25519.PubKey{Key: pkBytes}
}

// CreateTestPubKeys returns a total of numPubKeys public keys in ascending order.
func CreateTestPubKeys(numPubKeys int) []cryptotypes.PubKey {
	var publicKeys []cryptotypes.PubKey
	var buffer bytes.Buffer

	// start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 100; i < (numPubKeys + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") // base pubkey string
		buffer.WriteString(numString)                                                       // adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, NewPubKeyFromHex(buffer.String()))
		buffer.Reset()
	}

	return publicKeys
}

func MakeTestCtxAndKeeper(t *testing.T) (sdk.Context, Keeper) {
	var cdc = codec.NewProtoCodec()
	codec.RegisterCrypto(cdc)

	keyPoa := sdk.NewKVStoreKey(types.StoreKey)
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyPoa, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	_ = ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, true, nil)

	pk := params.NewKeeper(
		cdc,
		keyParams,
		tkeyParams,
	)

	accountKeeper := auth.NewAccountKeeper(
		cdc,
		keyAcc,
		pk.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	bk := bank.NewBaseKeeper(
		accountKeeper,
		pk.Subspace(bank.DefaultParamspace),
		nil,
	)

	keeper := NewKeeper(
		bk,
		cdc,
		keyPoa,
		pk.Subspace(DefaultParamspace),
	)
	keeper.SetParams(ctx, types.DefaultParams())

	return ctx, keeper
}
