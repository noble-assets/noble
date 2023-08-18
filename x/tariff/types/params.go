package types

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams creates a new Params object.
func NewParams(share math.LegacyDec, distributionEntities []DistributionEntity, transferFeeBPS math.Int, transferFeeMax math.Int, transferFeeDenom string) Params {
	return Params{
		Share:                share,
		DistributionEntities: distributionEntities,
		TransferFeeBps:       transferFeeBPS,
		TransferFeeMax:       transferFeeMax,
		TransferFeeDenom:     transferFeeDenom,
	}
}

// DefaultParams creates a default Params object.
func DefaultParams() Params {
	return Params{}
}

// Validate validates the provided params.
func (p *Params) Validate() error {
	if err := validateShare(p.Share); err != nil {
		return err
	}

	if err := validateDistributionEntities(p.DistributionEntities); err != nil {
		return err
	}

	if err := validateTransferFeeBPS(p.TransferFeeBps); err != nil {
		return err
	}

	if err := validateTransferFeeMax(p.TransferFeeMax); err != nil {
		return err
	}

	if err := validateTransferFeeDenom(p.TransferFeeDenom); err != nil {
		return err
	}

	return nil
}

//

func validateShare(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(math.LegacyZeroDec()) || v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("value must be a percentage")
	}

	return nil
}

func validateDistributionEntities(i interface{}) error {
	v, ok := i.([]DistributionEntity)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	sum := math.LegacyZeroDec()
	for _, entity := range v {
		if _, err := sdk.AccAddressFromBech32(entity.Address); err != nil {
			return err
		}

		count := 0
		for _, secondEntity := range v {
			if secondEntity.Address == entity.Address {
				count++
			}
		}

		if count > 1 {
			return fmt.Errorf("address occurred multiple times: %s", entity.Address)
		}

		if entity.Share.LT(math.LegacyZeroDec()) || entity.Share.GT(math.LegacyOneDec()) {
			return fmt.Errorf("entity share must be a percentage")
		}

		sum = sum.Add(entity.Share)
	}

	if len(v) > 0 && !sum.Equal(math.LegacyOneDec()) {
		return fmt.Errorf("sum of distribution entities is not 100%%: %s", sum)
	}

	return nil
}

func validateTransferFeeBPS(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(math.ZeroInt()) || v.GT(math.NewInt(10000)) {
		return fmt.Errorf("value is outside the range of 0 and 10000: %s", v)
	}

	return nil
}

func validateTransferFeeMax(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("value must be positive")
	}

	return nil
}

func validateTransferFeeDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == "" {
		return nil
	}

	return sdk.ValidateDenom(v)
}
