package keeper

import (
	dextypes "github.com/NicholasDotSol/duality/x/dex/types"
	"github.com/NicholasDotSol/duality/x/ibc-swap/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"
)

// Middleware must implement types.ChannelKeeper and types.PortKeeper, expected interfaces
// so that it can wrap IBC channel and port logic for underlying application.
var (
	_ types.ChannelKeeper = Keeper{}
	_ types.PortKeeper    = Keeper{}
)

// Keeper defines the swap middleware keeper.
type Keeper struct {
	cdc              codec.BinaryCodec
	paramSpace       paramtypes.Subspace
	msgServiceRouter *baseapp.MsgServiceRouter

	ics4Wrapper   porttypes.ICS4Wrapper
	channelKeeper types.ChannelKeeper
	portKeeper    types.PortKeeper
}

// NewKeeper creates a new swap Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	paramSpace paramtypes.Subspace,
	msgServiceRouter *baseapp.MsgServiceRouter,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
) Keeper {
	return Keeper{
		cdc:              cdc,
		paramSpace:       paramSpace,
		msgServiceRouter: msgServiceRouter,

		ics4Wrapper:   ics4Wrapper,
		channelKeeper: channelKeeper,
		portKeeper:    portKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+host.ModuleName+"-"+types.ModuleName)
}

// Swap calls into the base app's msg service router so that the appropriate handler is called when sending the swap msg.
func (k Keeper) Swap(ctx sdk.Context, msg *dextypes.MsgSwap) (*dextypes.MsgSwapResponse, error) {
	k.Logger(ctx).Error("INSIDE KEEPER SWAP CALL")
	swapHandler := k.msgServiceRouter.Handler(msg)
	if swapHandler == nil {
		k.Logger(ctx).Error("MsgService handler not found for MsgSwap")
	}

	res, err := swapHandler(ctx, msg)
	if err != nil {
		k.Logger(ctx).Error("Call to swap handler failed", "err", err)
		return nil, err
	}

	msgSwapRes := &dextypes.MsgSwapResponse{}
	if err := proto.Unmarshal(res.Data, msgSwapRes); err != nil {
		return nil, err
	}

	return msgSwapRes, nil
}

// BindPort defines a wrapper function for the port Keeper's function in
// order to expose it to module's InitGenesis function.
func (k Keeper) BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability {
	return k.portKeeper.BindPort(ctx, portID)
}

// GetChannel wraps IBC ChannelKeeper's GetChannel function.
func (k Keeper) GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool) {
	return k.channelKeeper.GetChannel(ctx, portID, channelID)
}

// GetPacketCommitment wraps IBC ChannelKeeper's GetPacketCommitment function.
func (k Keeper) GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte {
	return k.channelKeeper.GetPacketCommitment(ctx, portID, channelID, sequence)
}

// GetNextSequenceSend wraps IBC ChannelKeeper's GetNextSequenceSend function.
func (k Keeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	return k.channelKeeper.GetNextSequenceSend(ctx, portID, channelID)
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (k Keeper) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI) error {
	return k.ics4Wrapper.SendPacket(ctx, chanCap, packet)
}

// WriteAcknowledgement wraps IBC ChannelKeeper's WriteAcknowledgement function.
// ICS29 WriteAcknowledgement is used for asynchronous acknowledgements.
func (k Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, acknowledgement)
}
