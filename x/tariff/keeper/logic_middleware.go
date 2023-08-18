package keeper

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	// Capability
	capabilityTypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	// IBC Core
	clientTypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	portTypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	// IBC Transfer
	transferTypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

var _ portTypes.ICS4Wrapper = &Keeper{}

// SendPacket implements the ICS4Wrapper interface.
func (k *Keeper) SendPacket(ctx sdk.Context, chanCap *capabilityTypes.Capability, sourcePort string, sourceChannel string, timeoutHeight clientTypes.Height, timeoutTimestamp uint64, data []byte) (sequence uint64, err error) {
	var packet transferTypes.FungibleTokenPacketData
	if err := transferTypes.ModuleCdc.UnmarshalJSON(data, &packet); err != nil {
		// not fungible token packet data, forward to next middleware
		return k.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	params := k.GetParams(ctx)

	if packet.Denom != params.TransferFeeDenom {
		// not fee collection denom, forward to next middleware
		return k.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	fullAmount, ok := math.NewIntFromString(packet.Amount)
	if !ok {
		return 0, fmt.Errorf("failed to parse packet amount to sdk.Int %s", packet.Amount)
	}

	feeDec := math.LegacyNewDecFromInt(fullAmount).Mul(math.LegacyNewDecWithPrec(1, 4)).MulInt(params.TransferFeeBps)
	feeInt := feeDec.TruncateInt()

	if feeInt.Equal(sdk.ZeroInt()) {
		// fees are zero, forward to next middleware
		return k.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	if feeInt.GT(params.TransferFeeMax) {
		feeInt = params.TransferFeeMax
	}

	// all of the packet funds have been escrowed. Collect fees from the escrow account.
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		transferTypes.GetEscrowAddress(sourcePort, sourceChannel),
		k.feeCollectorName,
		sdk.NewCoins(sdk.NewCoin(packet.Denom, feeInt)),
	); err != nil {
		return 0, err
	}

	remaining := fullAmount.Sub(feeInt)

	packet.Amount = remaining.String()

	newData, err := transferTypes.ModuleCdc.MarshalJSON(&packet)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal new packet data: %w", err)
	}

	return k.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, newData)
}

// WriteAcknowledgement implements the ICS4Wrapper interface.
func (k *Keeper) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilityTypes.Capability,
	packet exported.PacketI,
	ack exported.Acknowledgement,
) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// GetAppVersion implements the ICS4Wrapper interface.
func (k *Keeper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
