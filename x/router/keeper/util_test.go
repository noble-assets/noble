package keeper_test

import (
	"bytes"
	"encoding/binary"
	"github.com/gogo/protobuf/proto"
	"github.com/strangelove-ventures/noble/x/router/keeper"
	"github.com/strangelove-ventures/noble/x/router/types"
	"math/big"
	"testing"

	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/stretchr/testify/require"
)

func TestDecodeMessage(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		msg      []byte
		expected keeper.Message
		err      error
	}{
		{
			desc: "Happy path",
			msg: bytesFromMessage(keeper.Message{
				Version:           1,
				SourceDomain:      2,
				DestinationDomain: 3,
				Nonce:             4,
				Sender:            fillByteArray(0, 32),
				Recipient:         fillByteArray(32, 32),
				DestinationCaller: fillByteArray(64, 32),
				MessageBody:       []byte("your average run of the mill message body"),
			}),
			expected: keeper.Message{
				Version:           1,
				SourceDomain:      2,
				DestinationDomain: 3,
				Nonce:             4,
				Sender:            fillByteArray(0, 32),
				Recipient:         fillByteArray(32, 32),
				DestinationCaller: fillByteArray(64, 32),
				MessageBody:       []byte("your average run of the mill message body"),
			},
		},
		{
			desc: "invalid",
			msg:  []byte("-1"),
			err:  types.ErrDecodingMessage,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := keeper.DecodeMessage(tc.msg)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.expected),
					nullify.Fill(*result),
				)
			}
		})
	}
}

func TestDecodeBurnMessage(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		msg      []byte
		expected keeper.BurnMessage
		err      error
	}{
		{
			desc: "Happy path",
			msg: bytesFromBurnMessage(keeper.BurnMessage{
				Version:       3,
				BurnToken:     []byte("01234567890123456789012345678912"),
				MintRecipient: []byte("01234567890123456789012345678912"),
				Amount:        *big.NewInt(int64(98999)),
				MessageSender: []byte("01234567890123456789012345678912"),
			}),
			expected: keeper.BurnMessage{
				Version:       3,
				BurnToken:     []byte("01234567890123456789012345678912"),
				MintRecipient: []byte("01234567890123456789012345678912"),
				Amount:        *big.NewInt(int64(98999)),
				MessageSender: []byte("01234567890123456789012345678912"),
			},
		},
		{
			desc: "invalid",
			msg:  []byte("-1"),
			err:  types.ErrDecodingBurnMessage,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := keeper.DecodeBurnMessage(tc.msg)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.expected),
					nullify.Fill(*result),
				)
			}
		})
	}
}

func TestDecodeIBCForward(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		msg      []byte
		expected types.IBCForwardMetadata
		err      error
	}{
		{
			desc: "Happy path",
			msg: marshalIBCForwardMetadata(&types.IBCForwardMetadata{
				Port:                 "1",
				Channel:              "2",
				DestinationReceiver:  "3",
				Memo:                 "4",
				TimeoutInNanoseconds: 0,
			}),
			expected: types.IBCForwardMetadata{
				Port:                 "1",
				Channel:              "2",
				DestinationReceiver:  "3",
				Memo:                 "4",
				TimeoutInNanoseconds: 0,
			},
		},
		{
			desc: "invalid",
			msg:  []byte("not a valid ibc forward message"),
			err:  types.ErrDecodingIBCForward,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := keeper.DecodeIBCForward(tc.msg)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.expected),
					nullify.Fill(result),
				)
			}
		})
	}
}

func bytesFromMessage(msg keeper.Message) []byte {
	result := make([]byte, keeper.MessageBodyIndex+len(msg.MessageBody))

	binary.BigEndian.PutUint32(result[keeper.VersionIndex:keeper.SourceDomainIndex], msg.Version)
	binary.BigEndian.PutUint32(result[keeper.SourceDomainIndex:keeper.DestinationDomainIndex], msg.SourceDomain)
	binary.BigEndian.PutUint32(result[keeper.DestinationDomainIndex:keeper.NonceIndex], msg.DestinationDomain)
	binary.BigEndian.PutUint64(result[keeper.NonceIndex:keeper.SenderIndex], msg.Nonce)

	copyBytes(msg.Sender, &result, 0, keeper.SenderIndex, keeper.Bytes32Len)
	copyBytes(msg.Recipient, &result, 0, keeper.RecipientIndex, keeper.Bytes32Len)
	copyBytes(msg.DestinationCaller, &result, 0, keeper.DestinationCallerIndex, keeper.Bytes32Len)
	copyBytes(msg.MessageBody, &result, 0, keeper.MessageBodyIndex, len(msg.MessageBody))

	return result
}

func bytesFromBurnMessage(msg keeper.BurnMessage) []byte {
	result := make([]byte, keeper.BurnMessageLen)

	binary.BigEndian.PutUint32(result[keeper.VersionIndex:keeper.BurnTokenIndex], msg.Version)
	amountBytes := uint256ToBytes(&msg.Amount)
	copy(result[keeper.AmountIndex:keeper.MsgSenderIndex], amountBytes[:])

	copyBytes(msg.BurnToken, &result, 0, keeper.BurnTokenIndex, keeper.BurnTokenLen)
	copyBytes(msg.MintRecipient, &result, 0, keeper.MintRecipientIndex, keeper.MintRecipientLen)
	copyBytes(msg.MessageSender, &result, 0, keeper.MsgSenderIndex, keeper.MsgSenderLen)

	return result
}

func copyBytes(src []byte, dest *[]byte, srcStartIndex int, destStartIndex int, length int) {
	for i := 0; i < length; i++ {
		(*dest)[destStartIndex+i] = src[srcStartIndex+i]
	}
}

func fillByteArray(start int, n int) []byte {
	res := make([]byte, n)
	for i := 0; i < n; i++ {
		res[i] = byte(start + i + 1)
	}
	return res
}

// Write uint256 to byte array in big-endian format
func uint256ToBytes(value *big.Int) []byte {
	// Create a buffer
	buf := new(bytes.Buffer)
	// Write the value into the buffer using big-endian byte order
	buf.Write(value.Bytes())

	// Pad the byte slice if it's not 32 bytes (256 bits) long
	padding := make([]byte, 32-len(buf.Bytes()))

	arr := make([]byte, 32)
	copy(arr, padding)
	copy(arr[len(padding):], buf.Bytes())

	return arr
}

func marshalIBCForwardMetadata(forward *types.IBCForwardMetadata) []byte {
	res, _ := proto.Marshal(forward)
	return res
}
