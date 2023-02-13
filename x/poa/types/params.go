package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

const (
	// Default percentage of vouches to join the set
	DefaultQuorum uint32 = 49

	// Default maximum number of bonded validators
	DefaultMaxValidators uint32 = 100

	// DefaultHistorical entries is 10000. Apps that don't use IBC can ignore this
	// value by not adding the staking module to the application module manager's
	// SetOrderBeginBlockers.
	DefaultHistoricalEntries uint32 = 10000

	// DefaultUnbondingTime reflects three weeks in seconds as the default
	// unbonding time.
	DefaultUnbondingTime time.Duration = time.Hour * 24 * 7 * 3

	// DefaultMinJailTime is the minimum amount of time a validator must be in jail before unjailing.
	DefaultMinJailTime time.Duration = time.Hour * 24
)

// nolint - Keys for parameter access
var (
	KeyQuorum            = []byte("Quorum")
	KeyMaxValidators     = []byte("MaxValidators")
	KeyUnbondingTime     = []byte("UnbondingTime")
	KeyHistoricalEntries = []byte("HistoricalEntries")
	KeyPowerReduction    = []byte("PowerReduction")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(quorum, maxValidators, historicalEntries uint32, unbondingTime time.Duration) Params {
	return Params{
		Quorum:            quorum,
		MaxValidators:     maxValidators,
		HistoricalEntries: historicalEntries,
		UnbondingTime:     unbondingTime,
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(DefaultQuorum, DefaultMaxValidators, DefaultHistoricalEntries, DefaultUnbondingTime)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyQuorum, &p.Quorum, validateQuorum),
		paramtypes.NewParamSetPair(KeyMaxValidators, &p.MaxValidators, validateMaxValidators),
		paramtypes.NewParamSetPair(KeyUnbondingTime, &p.UnbondingTime, validateUnbondingTime),
		paramtypes.NewParamSetPair(KeyHistoricalEntries, &p.HistoricalEntries, validateHistoricalEntries),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateQuorum(p.Quorum); err != nil {
		return err
	}

	if err := validateUnbondingTime(p.UnbondingTime); err != nil {
		return err
	}

	if err := validateMaxValidators(p.MaxValidators); err != nil {
		return err
	}

	return nil
}

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

func validateUnbondingTime(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("unbonding time must be positive: %d", v)
	}

	return nil
}

func validateHistoricalEntries(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func ValidatePowerReduction(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(sdk.NewInt(1)) {
		return fmt.Errorf("power reduction cannot be lower than 1")
	}

	return nil
}
