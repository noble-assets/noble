package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyShare                = []byte("Share")
	KeyDistributionEntities = []byte("DistributionEntities")
	KeyTransferFeeBPS       = []byte("TransferFeeBPS")
	KeyTransferFeeMax       = []byte("TransferFeeMax")
	KeyTransferFeeDenom     = []byte("TransferFeeDenom")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyShare, &p.Share, validateShare),
		paramtypes.NewParamSetPair(KeyDistributionEntities, &p.DistributionEntities, validateDistributionEntityParams),
		paramtypes.NewParamSetPair(KeyTransferFeeBPS, &p.TransferFeeBps, validateTransferFeeBPS),
		paramtypes.NewParamSetPair(KeyTransferFeeMax, &p.TransferFeeMax, validateTransferFeeMax),
		paramtypes.NewParamSetPair(KeyTransferFeeDenom, &p.TransferFeeDenom, validateTransferFeeDenom),
	}
}

func validateDistributionEntityParams(i interface{}) error {
	distributionEntities, ok := i.([]DistributionEntity)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	// ensure each denom is only registered one time.
	sum := sdk.ZeroDec()
	for _, d := range distributionEntities {
		adr, err := sdk.AccAddressFromBech32(d.Address)
		if err != nil {
			return fmt.Errorf("failed to parse bech32 address: %s", d.Address)
		}
		count := 0
		for _, dd := range distributionEntities {
			if dd.Address == adr.String() {
				count++
			}
		}
		if count > 1 {
			return fmt.Errorf("address is already added as a distribution entity: %s", adr)
		}

		if d.Share.LTE(sdk.ZeroDec()) || d.Share.GT(sdk.OneDec()) {
			return fmt.Errorf("distribution entity share must be greater than 0 and less than or equal to 100%%: %s", d.Share.String())
		}

		sum = sum.Add(d.Share)
	}

	if len(distributionEntities) > 0 && !sum.Equal(sdk.OneDec()) {
		return fmt.Errorf("sum of distribution entity shares don't equal 100%%: %s", sum.String())
	}

	return nil
}

func validateShare(i interface{}) error {
	share, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if share.LT(sdk.ZeroDec()) || share.GT(sdk.OneDec()) {
		return fmt.Errorf("share is outside of the range of 0 to 100%%: %s", share.String())
	}
	return nil
}

func validateTransferFeeBPS(i interface{}) error {
	transferFeeBPS, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if transferFeeBPS.LT(sdk.ZeroInt()) || transferFeeBPS.GT(sdk.NewInt(10000)) {
		return fmt.Errorf("ibc transfer basis points fee is outside of the range of 0 to 10000: %s", transferFeeBPS.String())
	}
	return nil
}

func validateTransferFeeMax(i interface{}) error {
	transferFeeMax, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if transferFeeMax.LT(sdk.ZeroInt()) {
		return fmt.Errorf("ibc transfer max fee is less than 0: %s", transferFeeMax.String())
	}
	return nil
}

func validateTransferFeeDenom(i interface{}) error {
	transferFeeDenom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if transferFeeDenom == "" {
		return nil
	}
	return sdk.ValidateDenom(transferFeeDenom)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateShare(p.Share); err != nil {
		return err
	}

	if err := validateDistributionEntityParams(p.DistributionEntities); err != nil {
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

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
