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
	cdc.RegisterConcrete(&MsgUpdateMasterMinter{}, "stabletokenfactory/UpdateMasterMinter", nil)
	cdc.RegisterConcrete(&MsgUpdatePauser{}, "stabletokenfactory/UpdatePauser", nil)
	cdc.RegisterConcrete(&MsgUpdateBlacklister{}, "stabletokenfactory/UpdateBlacklister", nil)
	cdc.RegisterConcrete(&MsgUpdateOwner{}, "stabletokenfactory/UpdateOwner", nil)
	cdc.RegisterConcrete(&MsgConfigureMinter{}, "stabletokenfactory/ConfigureMinter", nil)
	cdc.RegisterConcrete(&MsgRemoveMinter{}, "stabletokenfactory/RemoveMinter", nil)
	cdc.RegisterConcrete(&MsgMint{}, "stabletokenfactory/Mint", nil)
	cdc.RegisterConcrete(&MsgBurn{}, "stabletokenfactory/Burn", nil)
	cdc.RegisterConcrete(&MsgBlacklist{}, "stabletokenfactory/Blacklist", nil)
	cdc.RegisterConcrete(&MsgUnblacklist{}, "stabletokenfactory/Unblacklist", nil)
	cdc.RegisterConcrete(&MsgPause{}, "stabletokenfactory/Pause", nil)
	cdc.RegisterConcrete(&MsgUnpause{}, "stabletokenfactory/Unpause", nil)
	cdc.RegisterConcrete(&MsgConfigureMinterController{}, "stabletokenfactory/ConfigureMinterController", nil)
	cdc.RegisterConcrete(&MsgRemoveMinterController{}, "stabletokenfactory/RemoveMinterController", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateMasterMinter{},
		&MsgUpdatePauser{},
		&MsgUpdateBlacklister{},
		&MsgUpdateOwner{},
		&MsgConfigureMinter{},
		&MsgRemoveMinter{},
		&MsgMint{},
		&MsgBurn{},
		&MsgBlacklist{},
		&MsgUnblacklist{},
		&MsgPause{},
		&MsgUnpause{},
		&MsgConfigureMinterController{},
		&MsgRemoveMinterController{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
