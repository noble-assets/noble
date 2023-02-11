package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
)

// ABCIValidatorUpdate returns an abci.ValidatorUpdate from a staking validator type
// with the full validator power
func (v Validator) ABCIValidatorUpdate(unpacker cdctypes.AnyUnpacker, power int64) (abci.ValidatorUpdate, error) {
	var pubkey cryptotypes.PubKey
	if err := unpacker.UnpackAny(v.Pubkey, &pubkey); err != nil {
		return abci.ValidatorUpdate{}, err
	}
	return abci.ValidatorUpdate{
		PubKey: crypto.PublicKey{Sum: &crypto.PublicKey_Ed25519{Ed25519: pubkey.Bytes()}},
		Power:  power,
	}, nil
}
