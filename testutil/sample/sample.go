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

// AddressBz returns a slice of base64 encoded bytes representing an address.
func AddressBz() []byte {
	pk := ed25519.GenPrivKey().PubKey()
	address := sdk.AccAddress(pk.Address()).String()
	_, bz, _ := bech32.DecodeAndConvert(address)
	return bz
}

// Account represents a bech32 encoded address and the base64 encoded slice of bytes representing said address.
type Account struct {
	Address   string
	AddressBz []byte
}

// TestAccount returns an Account representing a newly generated PubKey.
func TestAccount() Account {
	pk := ed25519.GenPrivKey().PubKey()
	address := sdk.AccAddress(pk.Address()).String()
	_, bz, _ := bech32.DecodeAndConvert(address)
	return Account{
		Address:   address,
		AddressBz: bz,
	}
}
