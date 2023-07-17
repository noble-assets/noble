package keeper

import (
	"bytes"
	"encoding/binary"
	"github.com/strangelove-ventures/noble/x/cctp"

	sdkerrors "cosmossdk.io/errors"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func decodeMessage(msg []byte) types.Message {
	message := types.Message{
		Version:           binary.BigEndian.Uint32(msg[cctp.VersionIndex:cctp.SourceDomainIndex]),
		SourceDomainBytes: msg[cctp.SourceDomainIndex:cctp.DestinationDomainIndex],
		SourceDomain:      binary.BigEndian.Uint32(msg[cctp.SourceDomainIndex:cctp.DestinationDomainIndex]),
		DestinationDomain: binary.BigEndian.Uint32(msg[cctp.DestinationDomainIndex:cctp.NonceIndex]),
		NonceBytes:        msg[cctp.NonceIndex:cctp.SenderIndex],
		Nonce:             binary.BigEndian.Uint64(msg[cctp.NonceIndex:cctp.SenderIndex]),
		Sender:            msg[cctp.SenderIndex:cctp.RecipientIndex],
		Recipient:         msg[cctp.RecipientIndex:cctp.DestinationCallerIndex],
		DestinationCaller: msg[cctp.DestinationCallerIndex:cctp.MessageBodyIndex],
		MessageBody:       msg[cctp.MessageBodyIndex:],
	}

	return message
}

func decodeBurnMessage(msg []byte) types.BurnMessage {
	message := types.BurnMessage{
		Version:       binary.BigEndian.Uint32(msg[cctp.BurnMsgVersionIndex:cctp.BurnTokenIndex]),
		BurnToken:     msg[cctp.BurnTokenIndex:cctp.MintRecipientIndex],
		MintRecipient: msg[cctp.MintRecipientIndex:cctp.AmountIndex],
		Amount:        binary.BigEndian.Uint64(msg[cctp.AmountIndex:cctp.MsgSenderIndex]),
		MessageSender: msg[cctp.MsgSenderIndex:cctp.BurnMessageLen],
	}

	return message
}

func parseBurnMessageIntoBytes(msg types.BurnMessage) []byte {
	result := make([]byte, cctp.BurnMessageLen)

	versionBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, msg.Version)

	amountBytes := make([]byte, cctp.Bytes32Len)
	binary.LittleEndian.PutUint64(amountBytes, msg.Amount)

	copyBytes(cctp.BurnMsgVersionIndex, cctp.BurnTokenIndex, versionBytes, &result)
	copyBytes(cctp.BurnTokenIndex, cctp.MintRecipientIndex, msg.BurnToken, &result)
	copyBytes(cctp.MintRecipientIndex, cctp.AmountIndex, msg.MintRecipient, &result)
	copyBytes(cctp.AmountIndex, cctp.MsgSenderIndex, amountBytes, &result)

	return result
}

func copyBytes(start int, end int, copyFrom []byte, copyInto *[]byte) {
	for i := start; i < end; i++ {
		(*copyInto)[i] = copyFrom[i]
	}
}

/*
VerifyAttestationSignatures
* Rules for valid attestation:
* 1. length of `_attestation` == 65 (signature length) * signatureThreshold
* 2. addresses recovered from attestation must be in increasing order.
* 	For example, if signature A is signed by address 0x1..., and signature B
* 		is signed by address 0x2..., attestation must be passed as AB.
* 3. no duplicate signers
* 4. all signers must be enabled attesters
*/
func VerifyAttestationSignatures(
	message []byte,
	attestation []byte,
	attesters []types.Attester,
	signatureThreshold uint32,
) (bool, error) {
	if uint32(len(attestation)) != crypto.SignatureLength*signatureThreshold {
		return false, sdkerrors.Wrap(types.ErrSignatureVerification, "invalid attestation length")
	}

	if signatureThreshold == 0 {
		return false, sdkerrors.Wrap(types.ErrSignatureVerification, "signature threshold is 0")
	}

	// attester cannot be empty, so the recovered address should be bigger than latestAttesterAddress
	var latestAttesterAddress []byte

	digest := crypto.Keccak256Hash(message).Bytes()

	for i := uint32(0); i < signatureThreshold; i++ {
		signature := attestation[i*crypto.SignatureLength : (i*crypto.SignatureLength)+crypto.SignatureLength]

		recoveredAttester, err := crypto.Ecrecover(digest, signature)
		if err != nil {
			return false, sdkerrors.Wrap(types.ErrSignatureVerification, "failed to recover attester")
		}

		// Signatures must be in increasing order of address, and may not duplicate signatures from same address
		if bytes.Compare(latestAttesterAddress, recoveredAttester) > -1 {
			return false, sdkerrors.Wrap(types.ErrSignatureVerification, "Invalid signature order or dupe")
		}

		// check that recovered attester is valid
		contains := false
		for _, attester := range attesters {
			if attester.Attester == string(recoveredAttester) {
				contains = true
			}
		}

		if !contains {
			return false, sdkerrors.Wrap(types.ErrSignatureVerification, "Invalid signature: not attester")
		}

		latestAttesterAddress = recoveredAttester

	}
	return true, nil
}
