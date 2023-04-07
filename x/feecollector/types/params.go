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
	if p.Share.LT(sdk.NewDec(0)) || p.Share.GT(sdk.NewDec(1)) {
		return fmt.Errorf("share is outside of the range of 0 to 100%%: %s", p.Share.String())
	}

	i := sdk.NewDec(0)
	for _, d := range p.DistributionEntities {
		if d.Share.LT(sdk.NewDec(0)) || d.Share.GT(sdk.NewDec(1)) {
			return fmt.Errorf("distribution entity share is outside of the range of 0 to 100%%: %s", d.Share.String())
		}
		i.Add(d.Share)
	}
	if !i.Equal(sdk.NewDec(1)) {
		return fmt.Errorf("sum of distribution entity shares don't equal 100%%: %s", i.String())
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
