package types

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
)

// ABCIValidatorUpdate returns an abci.ValidatorUpdate from a staking validator type
// with the full validator power
func (v Validator) ABCIValidatorUpdate(power int64) (abci.ValidatorUpdate, error) {
	var pubkey crypto.PublicKey
	if err := ModuleCdc.UnpackAny(v.Pubkey, &pubkey); err != nil {
		return abci.ValidatorUpdate{}, err
	}
	return abci.ValidatorUpdate{
		PubKey: pubkey,
		Power:  power,
	}, nil
}
