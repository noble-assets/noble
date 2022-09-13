package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgChangeAdmin{}, "tokenfactory/ChangeAdmin", nil)
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
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgChangeAdmin{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateMasterMinter{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdatePauser{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateBlacklister{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateOwner{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgConfigureMinter{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRemoveMinter{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMint{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgBurn{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgBlacklist{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUnblacklist{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgPause{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUnpause{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
