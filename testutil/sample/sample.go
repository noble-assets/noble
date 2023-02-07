package sample

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// AccAddress returns a sample account address
func AccAddress() string {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr).String()
}

// PubKeyBytes returns a sample pubkey as a slice of bytes.
func PubKeyBytes() []byte {
	return ed25519.GenPrivKey().PubKey().Bytes()
}

// Account represents a bech32 encoded address and the base64 encoded slice of bytes representing said address.
type Account struct {
	Address  string
	PubKeyBz []byte
}

// TestAccount returns an Account representing a newly generated PubKey.
func TestAccount() Account {
	pk := ed25519.GenPrivKey().PubKey()
	address := sdk.AccAddress(pk.Address()).String()
	_, bz, _ := bech32.DecodeAndConvert(address)
	return Account{
		Address:  address,
		PubKeyBz: bz,
	}
}
