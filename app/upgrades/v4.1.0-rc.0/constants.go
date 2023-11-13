package v4m1p0rc0

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// UpgradeName is the name of this specific software upgrade used on-chain.
const UpgradeName = "v4.1.0-rc.0"

// TestnetChainID is the Chain ID of the Noble testnet (Grand).
const TestnetChainID = "grand-1"

// USDLRMetadata is the metadata stored in the x/bank module for $USDLR.
var USDLRMetadata = banktypes.Metadata{
	Description: "USDLR is a fiat-backed stablecoin issued by Stable. Stable pays DeFi protocols who distribute USDLR.",
	DenomUnits: []*banktypes.DenomUnit{
		{
			Denom:    "uusdlr",
			Exponent: 0,
			Aliases:  []string{"microusdlr"},
		},
		{
			Denom:    "usdlr",
			Exponent: 6,
		},
	},
	Base:    "uusdlr",
	Display: "usdlr",
	Name:    "USDLR by Stable",
	Symbol:  "USDLR",
}

// StableAddress is the address used by Stable initially.
// TODO: Update with correct address.
const StableAddress = "noble10uu75g7zl0gnzt0wz46htgqnl5ml27dnthcztx"
