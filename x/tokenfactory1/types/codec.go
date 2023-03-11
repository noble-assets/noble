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
	cdc.RegisterConcrete(&MsgUpdateMasterMinter{}, "tokenfactory1/UpdateMasterMinter", nil)
	cdc.RegisterConcrete(&MsgUpdatePauser{}, "tokenfactory1/UpdatePauser", nil)
	cdc.RegisterConcrete(&MsgUpdateBlacklister{}, "tokenfactory1/UpdateBlacklister", nil)
	cdc.RegisterConcrete(&MsgUpdateOwner{}, "tokenfactory1/UpdateOwner", nil)
	cdc.RegisterConcrete(&MsgConfigureMinter{}, "tokenfactory1/ConfigureMinter", nil)
	cdc.RegisterConcrete(&MsgRemoveMinter{}, "tokenfactory1/RemoveMinter", nil)
	cdc.RegisterConcrete(&MsgMint{}, "tokenfactory1/Mint", nil)
	cdc.RegisterConcrete(&MsgBurn{}, "tokenfactory1/Burn", nil)
	cdc.RegisterConcrete(&MsgBlacklist{}, "tokenfactory1/Blacklist", nil)
	cdc.RegisterConcrete(&MsgUnblacklist{}, "tokenfactory1/Unblacklist", nil)
	cdc.RegisterConcrete(&MsgPause{}, "tokenfactory1/Pause", nil)
	cdc.RegisterConcrete(&MsgUnpause{}, "tokenfactory1/Unpause", nil)
	cdc.RegisterConcrete(&MsgConfigureMinterController{}, "tokenfactory1/ConfigureMinterController", nil)
	cdc.RegisterConcrete(&MsgRemoveMinterController{}, "tokenfactory1/RemoveMinterController", nil)
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
