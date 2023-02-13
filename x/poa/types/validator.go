package types

import (
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmprotocrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

var _ stakingtypes.ValidatorI = Validator{}

const ValidatorActivePower = int64(10)

// ToStakingValidator creates a staking representation of the validator
// to satisfy the staking keeper interface shims for use outside of the module.
func (v *Validator) ToStakingValidator() stakingtypes.Validator {
	return stakingtypes.Validator{
		OperatorAddress:   sdk.ValAddress(v.Address).String(),
		ConsensusPubkey:   v.Pubkey,
		Jailed:            v.Jailed,
		Status:            v.GetStatus(),
		Tokens:            v.GetTokens(),
		DelegatorShares:   v.GetDelegatorShares(),
		Description:       v.Description,
		MinSelfDelegation: v.GetMinSelfDelegation(),
	}
}

// ABCIValidatorUpdate returns an abci.ValidatorUpdate from a staking validator type
// with the full validator power
func (v Validator) ABCIValidatorUpdate(power int64) (abci.ValidatorUpdate, error) {
	pubKey, err := v.TmConsPublicKey()
	if err != nil {
		return abci.ValidatorUpdate{}, err
	}
	return abci.ValidatorUpdate{
		PubKey: pubKey,
		Power:  power,
	}, nil
}

func (v Validator) IsJailed() bool {
	return v.Jailed
}

func (v *Validator) EligibleToJoinSet() bool {
	return v.IsAccepted && !v.Jailed
}

func (v Validator) GetMoniker() string {
	return v.Description.Moniker
}

func (v Validator) GetStatus() stakingtypes.BondStatus {
	return stakingtypes.Bonded
}

func (v Validator) IsBonded() bool {
	return v.InSet
}

func (v Validator) IsUnbonded() bool {
	return false
}

func (v Validator) IsUnbonding() bool {
	return false
}

func (v Validator) GetOperator() sdk.ValAddress {
	return sdk.ValAddress(v.Address)
}

func (v Validator) ConsPubKey() (cryptotypes.PubKey, error) {
	var pubKey cryptotypes.PubKey
	if err := ModuleCdc.UnpackAny(v.Pubkey, &pubKey); err != nil {
		return nil, err
	}
	return pubKey, nil
}

func (v Validator) TmConsPublicKey() (tmprotocrypto.PublicKey, error) {
	pubKey, err := v.ConsPubKey()
	if err != nil {
		return tmprotocrypto.PublicKey{}, err
	}
	return cryptocodec.ToTmProtoPublicKey(pubKey)
}

func (v Validator) GetConsAddr() (sdk.ConsAddress, error) {
	pubKey, err := v.ConsPubKey()
	if err != nil {
		return nil, err
	}
	return sdk.ConsAddress(pubKey.Address().Bytes()), nil
}

func (v Validator) GetTokens() sdk.Int {
	return sdk.ZeroInt()
}

func (v Validator) GetBondedTokens() sdk.Int {
	return sdk.ZeroInt()
}

func (v Validator) GetConsensusPower(sdk.Int) int64 {
	if v.InSet {
		return ValidatorActivePower
	}
	return 0
}

func (v Validator) GetCommission() sdk.Dec {
	return sdk.ZeroDec()
}

func (v Validator) GetMinSelfDelegation() sdk.Int {
	return sdk.ZeroInt()
}

func (v Validator) GetDelegatorShares() sdk.Dec {
	return sdk.ZeroDec()
}

func (v Validator) TokensFromShares(sdk.Dec) sdk.Dec {
	return sdk.ZeroDec()
}

func (v Validator) TokensFromSharesTruncated(sdk.Dec) sdk.Dec {
	return sdk.ZeroDec()
}

func (v Validator) TokensFromSharesRoundUp(sdk.Dec) sdk.Dec {
	return sdk.ZeroDec()
}

func (v Validator) SharesFromTokens(amt sdk.Int) (sdk.Dec, error) {
	return sdk.ZeroDec(), nil
}

func (v Validator) SharesFromTokensTruncated(amt sdk.Int) (sdk.Dec, error) {
	return sdk.ZeroDec(), nil
}
