package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	chantypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/strangelove-ventures/noble/x/tariff/types"
)

var _ porttypes.ICS4Wrapper = Keeper{}

type (
	Keeper struct {
		paramstore       paramtypes.Subspace
		authKeeper       types.AccountKeeper
		bankKeeper       types.BankKeeper
		feeCollectorName string // name of the FeeCollector ModuleAccount
		ics4Wrapper      porttypes.ICS4Wrapper
	}
)

// NewKeeper constructs a new fee collector keeper.
func NewKeeper(
	ps paramtypes.Subspace,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	feeCollectorName string,
	ics4Wrapper porttypes.ICS4Wrapper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		paramstore:       ps,
		authKeeper:       authKeeper,
		bankKeeper:       bankKeeper,
		feeCollectorName: feeCollectorName,
		ics4Wrapper:      ics4Wrapper,
	}
}

// SendPacket implements the ICS4Wrapper interface.
func (k Keeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
) error {
	chanPacket, ok := packet.(chantypes.Packet)
	if !ok {
		// not channel packet, forward to next middleware
		return k.ics4Wrapper.SendPacket(ctx, chanCap, packet)
	}

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(chanPacket.Data, &data); err != nil {
		// not fungible token packet data, forward to next middleware
		return k.ics4Wrapper.SendPacket(ctx, chanCap, packet)
	}

	params := k.GetParams(ctx)
	bpsFee, maxFee, feeDenom := params.TransferFeeBps, params.TransferFeeMax, params.TransferFeeDenom

	if data.Denom != feeDenom {
		// not fee collection denom, forward to next middleware
		return k.ics4Wrapper.SendPacket(ctx, chanCap, packet)
	}

	fullAmount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return fmt.Errorf("failed to parse packet amount to sdk.Int %s", data.Amount)
	}

	feeDec := fullAmount.ToDec().Mul(sdk.NewDecWithPrec(1, 4)).MulInt(bpsFee)
	feeInt := feeDec.TruncateInt()

	if feeInt.Equal(sdk.ZeroInt()) {
		// fees are zero, forward to next middleware
		return k.ics4Wrapper.SendPacket(ctx, chanCap, packet)
	}

	if feeInt.GT(maxFee) {
		feeInt = maxFee
	}

	// all of the packet funds have been escrowed. Collect fees from the escrow account.
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		transfertypes.GetEscrowAddress(chanPacket.SourcePort, chanPacket.SourceChannel),
		k.feeCollectorName,
		sdk.NewCoins(sdk.NewCoin(data.Denom, feeInt)),
	); err != nil {
		return err
	}

	remaining := fullAmount.Sub(feeInt)

	data.Amount = remaining.String()

	newData, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	if err != nil {
		return fmt.Errorf("failed to marshal new packet data: %w", err)
	}

	chanPacket.Data = newData

	return k.ics4Wrapper.SendPacket(ctx, chanCap, chanPacket)
}

// WriteAcknowledgement implements the ICS4Wrapper interface.
func (k Keeper) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
	ack exported.Acknowledgement,
) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

func (k Keeper) GetAppVersion(
	ctx sdk.Context,
	portID,
	channelID string,
) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
