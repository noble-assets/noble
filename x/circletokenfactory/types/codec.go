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
	cdc.RegisterConcrete(&MsgUpdateMasterMinter{}, "circletokenfactory/UpdateMasterMinter", nil)
	cdc.RegisterConcrete(&MsgUpdatePauser{}, "circletokenfactory/UpdatePauser", nil)
	cdc.RegisterConcrete(&MsgUpdateBlacklister{}, "circletokenfactory/UpdateBlacklister", nil)
	cdc.RegisterConcrete(&MsgUpdateOwner{}, "circletokenfactory/UpdateOwner", nil)
	cdc.RegisterConcrete(&MsgConfigureMinter{}, "circletokenfactory/ConfigureMinter", nil)
	cdc.RegisterConcrete(&MsgRemoveMinter{}, "circletokenfactory/RemoveMinter", nil)
	cdc.RegisterConcrete(&MsgMint{}, "circletokenfactory/Mint", nil)
	cdc.RegisterConcrete(&MsgBurn{}, "circletokenfactory/Burn", nil)
	cdc.RegisterConcrete(&MsgBlacklist{}, "circletokenfactory/Blacklist", nil)
	cdc.RegisterConcrete(&MsgUnblacklist{}, "circletokenfactory/Unblacklist", nil)
	cdc.RegisterConcrete(&MsgPause{}, "circletokenfactory/Pause", nil)
	cdc.RegisterConcrete(&MsgUnpause{}, "circletokenfactory/Unpause", nil)
	cdc.RegisterConcrete(&MsgConfigureMinterController{}, "circletokenfactory/ConfigureMinterController", nil)
	cdc.RegisterConcrete(&MsgRemoveMinterController{}, "circletokenfactory/RemoveMinterController", nil)
	// this line is used by starport scaffolding # 2
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

	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
