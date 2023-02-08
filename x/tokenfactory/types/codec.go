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
	cdc.RegisterConcrete(&MsgUpdateMasterMinter{}, "tokenfactory/UpdateMasterMinter", nil)
	cdc.RegisterConcrete(&MsgUpdatePauser{}, "tokenfactory/UpdatePauser", nil)
	cdc.RegisterConcrete(&MsgUpdateBlacklister{}, "tokenfactory/UpdateBlacklister", nil)
	cdc.RegisterConcrete(&MsgUpdateOwner{}, "tokenfactory/UpdateOwner", nil)
	cdc.RegisterConcrete(&MsgConfigureMinter{}, "tokenfactory/ConfigureMinter", nil)
	cdc.RegisterConcrete(&MsgRemoveMinter{}, "tokenfactory/RemoveMinter", nil)
	cdc.RegisterConcrete(&MsgMint{}, "tokenfactory/Mint", nil)
	cdc.RegisterConcrete(&MsgBurn{}, "tokenfactory/Burn", nil)
	cdc.RegisterConcrete(&MsgBlacklist{}, "tokenfactory/Blacklist", nil)
	cdc.RegisterConcrete(&MsgUnblacklist{}, "tokenfactory/Unblacklist", nil)
	cdc.RegisterConcrete(&MsgPause{}, "tokenfactory/Pause", nil)
	cdc.RegisterConcrete(&MsgUnpause{}, "tokenfactory/Unpause", nil)
	cdc.RegisterConcrete(&MsgConfigureMinterController{}, "tokenfactory/ConfigureMinterController", nil)
	cdc.RegisterConcrete(&MsgRemoveMinterController{}, "tokenfactory/RemoveMinterController", nil)
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
