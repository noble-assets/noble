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
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
