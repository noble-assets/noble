package blockibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/strangelove-ventures/noble/x/tokenfactory/keeper"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
	keeper_1 "github.com/strangelove-ventures/noble/x/usdc_tokenfactory/keeper"
	types_1 "github.com/strangelove-ventures/noble/x/usdc_tokenfactory/types"
)

var _ porttypes.IBCModule = &IBCMiddleware{}

// IBCMiddleware implements the tokenfactory keeper in order to check against blacklisted addresses.
type IBCMiddleware struct {
	app      porttypes.IBCModule
	keeper   *keeper.Keeper
	keeper_1 *keeper_1.Keeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application.
func NewIBCMiddleware(app porttypes.IBCModule, k *keeper.Keeper, k_1 *keeper_1.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:      app,
		keeper:   k,
		keeper_1: k_1,
	}
}

// OnChanOpenInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

// OnChanOpenTry implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID, channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID, channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface.
func (im IBCMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket intercepts the packet data and checks the sender and receiver address against
// the blacklisted addresses held in the tokenfactory keeper. If the address is found in the blacklist, an
// acknowledgment error is returned.
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {

	var data transfertypes.FungibleTokenPacketData
	var ackErr error
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		ackErr = sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
		return channeltypes.NewErrorAcknowledgement(ackErr.Error())
	}

	denomTrace := transfertypes.ParseDenomTrace(data.Denom)

	mintingDenom := im.keeper.GetMintingDenom(ctx)
	mintingDenom_1 := im.keeper_1.GetMintingDenom(ctx)

	switch {
	// denom is not tokenfactory denom
	case denomTrace.BaseDenom != mintingDenom.Denom && denomTrace.BaseDenom != mintingDenom_1.Denom:
		return im.app.OnRecvPacket(ctx, packet, relayer)
	// denom is tokenfactory asset
	case denomTrace.BaseDenom == mintingDenom.Denom:
		if im.keeper.GetPaused(ctx).Paused {
			return channeltypes.NewErrorAcknowledgement(types.ErrPaused.Error())
		}

		_, addressBz, err := bech32.DecodeAndConvert(data.Receiver)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err.Error())
		}

		_, found := im.keeper.GetBlacklisted(ctx, addressBz)
		if found {
			ackErr = sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "receiver address is blacklisted")
			return channeltypes.NewErrorAcknowledgement(ackErr.Error())
		}

		_, addressBz, err = bech32.DecodeAndConvert(data.Sender)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err.Error())
		}

		_, found = im.keeper.GetBlacklisted(ctx, addressBz)
		if found {
			ackErr = sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "sender address is blacklisted")
			return channeltypes.NewErrorAcknowledgement(ackErr.Error())
		}
	// denom is usdc_tokenfactory asset
	case denomTrace.BaseDenom == mintingDenom_1.Denom:
		if im.keeper_1.GetPaused(ctx).Paused {
			return channeltypes.NewErrorAcknowledgement(types_1.ErrPaused.Error())
		}

		_, addressBz, err := bech32.DecodeAndConvert(data.Receiver)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err.Error())
		}

		_, found := im.keeper_1.GetBlacklisted(ctx, addressBz)
		if found {
			ackErr = sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "receiver address is blacklisted")
			return channeltypes.NewErrorAcknowledgement(ackErr.Error())
		}

		_, addressBz, err = bech32.DecodeAndConvert(data.Sender)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err.Error())
		}

		_, found = im.keeper_1.GetBlacklisted(ctx, addressBz)
		if found {
			ackErr = sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "sender address is blacklisted")
			return channeltypes.NewErrorAcknowledgement(ackErr.Error())
		}

	}
	return im.app.OnRecvPacket(ctx, packet, relayer)

}

// OnAcknowledgementPacket implements the IBCModule interface.
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}
