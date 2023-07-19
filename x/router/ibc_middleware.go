package router

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/strangelove-ventures/noble/x/router/keeper"
)

var _ porttypes.IBCModule = IBCMiddleware{}

type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper *keeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application.
func NewIBCMiddleware(app porttypes.IBCModule, k *keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	if inFlightPacket, found := im.keeper.GetInFlightPacket(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence); found {
		var ack channeltypes.Acknowledgement
		if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
		}

		if ack.Success() {
			im.keeper.DeleteMint(ctx, inFlightPacket.SourceDomain, inFlightPacket.SourceDomainSender, inFlightPacket.Nonce)
			im.keeper.DeleteIBCForward(ctx, inFlightPacket.SourceDomain, inFlightPacket.SourceDomainSender, inFlightPacket.Nonce)
			im.keeper.DeleteInFlightPacket(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)

		} else { // error on destination
			im.keeper.DeleteInFlightPacket(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)

			// keep mint and mark IBCForward to indicate ack error for retry for future replaceDepositForBurnWithMetadata
			if existingIBCForward, found := im.keeper.GetIBCForward(ctx, inFlightPacket.SourceDomain, inFlightPacket.SourceDomainSender, inFlightPacket.Nonce); found {
				existingIBCForward.AckError = true
				im.keeper.SetIBCForward(ctx, existingIBCForward)
			}
		}
	}

	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		im.keeper.Logger(ctx).Error("error parsing packet data from timeout packet",
			"sequence", packet.Sequence,
			"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
			"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
			"error", err,
		)
		return im.app.OnTimeoutPacket(ctx, packet, relayer)
	}

	if inFlightPacket, found := im.keeper.GetInFlightPacket(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence); found {
		im.keeper.DeleteInFlightPacket(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
		// timeout should be retried. In order to do that, we need to handle this timeout to refund on this chain first.
		if err := im.app.OnTimeoutPacket(ctx, packet, relayer); err != nil {
			return err
		}

		existingIBCForward, found := im.keeper.GetIBCForward(ctx, inFlightPacket.SourceDomain, inFlightPacket.SourceDomainSender, inFlightPacket.Nonce)
		if !found {
			panic("no existing ibc forward metadata in store for in flight packet")
		}

		existingMint, found := im.keeper.GetMint(ctx, inFlightPacket.SourceDomain, inFlightPacket.SourceDomainSender, inFlightPacket.Nonce)
		if !found {
			panic("no existing mint in store for in flight packet")
		}
		return im.keeper.ForwardPacket(ctx, *existingIBCForward.Metadata, existingMint)
	}

	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// OnChanOpenInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *types.Capability, counterparty channeltypes.Counterparty, version string) error {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

// OnChanOpenTry implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID, channelID string, channelCap *types.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenAck(ctx sdk.Context, portID, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
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

// OnRecvPacket implements the IBCModule interface.
func (im IBCMiddleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	return im.app.OnRecvPacket(ctx, packet, relayer)
}
