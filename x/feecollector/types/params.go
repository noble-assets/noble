package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.Share.LT(sdk.ZeroDec()) || p.Share.GT(sdk.OneDec()) {
		return fmt.Errorf("share is outside of the range of 0 to 100%%: %s", p.Share.String())
	}

	i := sdk.ZeroDec()
	for _, d := range p.DistributionEntities {
		_, err := sdk.AccAddressFromBech32(d.Address)
		if err != nil {
			return fmt.Errorf("failed to parse bech32 address: %s", d.Address)
		}

		if d.Share.LT(sdk.ZeroDec()) || d.Share.GT(sdk.OneDec()) {
			return fmt.Errorf("distribution entity share is outside of the range of 0 to 100%%: %s", d.Share.String())
		}

		i = i.Add(d.Share)
	}

	if len(p.DistributionEntities) > 0 && !i.Equal(sdk.OneDec()) {
		return fmt.Errorf("sum of distribution entity shares don't equal 100%%: %s", i.String())
	}

	if p.TransferFeeBps.LT(sdk.ZeroInt()) || p.TransferFeeBps.GT(sdk.NewInt(10000)) {
		return fmt.Errorf("ibc transfer basis points fee is outside of the range of 0 to 10000: %s", p.TransferFeeBps.String())
	}

	if p.TransferFeeMax.LT(sdk.ZeroInt()) {
		return fmt.Errorf("ibc transfer max fee is less than 0: %s", p.TransferFeeMax.String())
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
