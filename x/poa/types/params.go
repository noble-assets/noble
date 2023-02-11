package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

const (
	// Default percentage of votes to join the set
	DefaultQuorum uint32 = 49

	// Default maximum number of bonded validators
	DefaultMaxValidators uint32 = 100
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(quorum uint32, maxValidators uint32) Params {
	return Params{
		Quorum:        quorum,
		MaxValidators: maxValidators,
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(DefaultQuorum, DefaultMaxValidators)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyQuorum, &p.Quorum, validateQuorum),
		paramtypes.NewParamSetPair(KeyMaxValidators, &p.MaxValidators, validateMaxValidators),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// nolint - Keys for parameter access
var (
	KeyQuorum        = []byte("Quorum")
	KeyMaxValidators = []byte("MaxValidators")
)

func validateQuorum(i interface{}) error {
	val, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid type: %T", val)
	}

	if val > 100 {
		return fmt.Errorf("quorum must be less than 100: %d", val)
	}

	return nil
}

func validateMaxValidators(i interface{}) error {
	val, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid type: %T", i)
	}

	if val == 0 {
		return fmt.Errorf("max validators must greater than 0: %d", val)
	}

	return nil
}
