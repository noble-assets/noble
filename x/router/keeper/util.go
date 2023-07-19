package keeper

import (
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/strangelove-ventures/noble/x/router/types"
)

type BurnMessage struct {
	Version       uint32
	BurnToken     []byte
	MintRecipient []byte
	Amount        big.Int
	MessageSender []byte
}

type Message struct {
	Version           uint32
	SourceDomain      uint32
	DestinationDomain uint32
	Nonce             uint64
	Sender            []byte
	Recipient         []byte
	DestinationCaller []byte
	MessageBody       []byte
}

const (
	// Indices of each field in message
	VersionIndex           = 0
	SourceDomainIndex      = 4
	DestinationDomainIndex = 8
	NonceIndex             = 12
	SenderIndex            = 20
	RecipientIndex         = 52
	DestinationCallerIndex = 84
	MessageBodyIndex       = 116

	// Indices of each field in BurnMessage
	BurnMsgVersionIndex = 0
	BurnTokenIndex      = 4
	BurnTokenLen        = 32
	MintRecipientIndex  = 36
	MintRecipientLen    = 32
	AmountIndex         = 68
	MsgSenderIndex      = 100
	MsgSenderLen        = 32
	// 4 byte version + 32 bytes burnToken + 32 bytes mintRecipient + 32 bytes amount + 32 bytes messageSender
	BurnMessageLen = 132

	Bytes32Len = 32
)

func DecodeIBCForward(msg []byte) (types.IBCForwardMetadata, error) {
	var res types.IBCForwardMetadata
	if err := proto.Unmarshal(msg, &res); err != nil {
		return types.IBCForwardMetadata{}, sdkerrors.Wrapf(types.ErrDecodingIBCForward, "error decoding ibc forward")
	}

	return res, nil
}

func bytesToBigInt(data []byte) big.Int {
	value := big.Int{}
	value.SetBytes(data)

	return value

}
