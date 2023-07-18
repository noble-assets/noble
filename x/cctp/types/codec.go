package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateAuthority{}, "cctp/UpdateAuthority", nil)
	cdc.RegisterConcrete(&MsgAddPublicKey{}, "cctp/AddPublicKey", nil)
	cdc.RegisterConcrete(&MsgRemovePublicKey{}, "cctp/DeletePublicKey", nil)
	cdc.RegisterConcrete(&MsgPauseBurningAndMinting{}, "cctp/PauseBurningAndMinting", nil)
	cdc.RegisterConcrete(&MsgUnpauseBurningAndMinting{}, "cctp/UnpauseBurningAndMinting", nil)
	cdc.RegisterConcrete(&MsgPauseSendingAndReceivingMessages{}, "cctp/PauseSendingAndReceivingMessages", nil)
	cdc.RegisterConcrete(&MsgUnpauseSendingAndReceivingMessages{}, "cctp/UnpauseSendingAndReceivingMessages", nil)
	cdc.RegisterConcrete(&MsgUpdateMaxMessageBodySize{}, "cctp/UpdateMaxMessageBodySize", nil)
	cdc.RegisterConcrete(&MsgUpdatePerMessageBurnLimit{}, "cctp/UpdatePerMessageBurnLimit", nil)
	cdc.RegisterConcrete(&MsgDepositForBurn{}, "cctp/DepositForBurn", nil)
	cdc.RegisterConcrete(&MsgDepositForBurnWithCaller{}, "cctp/DepositForBurnWithCaller", nil)
	cdc.RegisterConcrete(&MsgReplaceDepositForBurn{}, "cctp/ReplaceDepositForBurn", nil)
	cdc.RegisterConcrete(&MsgReceiveMessage{}, "cctp/ReceiveMessage", nil)
	cdc.RegisterConcrete(&MsgSendMessage{}, "cctp/SendMessage", nil)
	cdc.RegisterConcrete(&MsgSendMessageWithCaller{}, "cctp/SendMessageWithCaller", nil)
	cdc.RegisterConcrete(&MsgReplaceMessage{}, "cctp/ReplaceMessage", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateAuthority{},
		&MsgAddPublicKey{},
		&MsgRemovePublicKey{},
		&MsgPauseBurningAndMinting{},
		&MsgUnpauseBurningAndMinting{},
		&MsgPauseSendingAndReceivingMessages{},
		&MsgUnpauseSendingAndReceivingMessages{},
		&MsgUpdateMaxMessageBodySize{},
		&MsgUpdatePerMessageBurnLimit{},
		&MsgDepositForBurn{},
		&MsgDepositForBurnWithCaller{},
		&MsgReplaceDepositForBurn{},
		&MsgReceiveMessage{},
		&MsgSendMessage{},
		&MsgSendMessageWithCaller{},
		&MsgReplaceMessage{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
