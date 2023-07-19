package keeper

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func ParseIntoMessage(msg []byte) types.Message {
	message := types.Message{
		Version:           binary.BigEndian.Uint32(msg[0:4]),
		SourceDomainBytes: msg[4:8],
		SourceDomain:      binary.BigEndian.Uint32(msg[4:8]),
		DestinationDomain: binary.BigEndian.Uint32(msg[8:12]),
		NonceBytes:        msg[12:20],
		Nonce:             binary.BigEndian.Uint64(msg[12:20]),
		Sender:            msg[20:52],
		Recipient:         msg[52:84],
		DestinationCaller: msg[84:116],
		MessageBody:       msg[116:],
	}

	return message
}

func ParseIntoBurnMessage(msg []byte) types.BurnMessage {
	message := types.BurnMessage{
		Version:       binary.BigEndian.Uint32(msg[0:4]),
		BurnToken:     msg[4:36],
		MintRecipient: msg[36:68],
		Amount:        binary.BigEndian.Uint64(msg[92:100]),
		MessageSender: msg[100:132],
	}

	return message
}

func ParseBurnMessageIntoBytes(msg types.BurnMessage) []byte {
	result := make([]byte, burnMessageLength)

	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, msg.Version)

	amountBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(amountBytes, msg.Amount)

	amountBytesPadded := append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, amountBytes...)

	copy(result[0:4], versionBytes)
	copy(result[4:36], msg.BurnToken)
	copy(result[36:68], msg.MintRecipient)
	copy(result[68:100], amountBytesPadded)
	copy(result[100:132], msg.MessageSender)

	return result
}

func ParseIntoMessageBytes(msg types.Message) []byte {

	result := make([]byte, messageBodyIndex+len(msg.MessageBody))

	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, msg.Version)

	sourceDomainBytes := make([]byte, 32)
	binary.BigEndian.PutUint32(sourceDomainBytes, msg.SourceDomain)

	destinationDomain := make([]byte, 32)
	binary.BigEndian.PutUint32(destinationDomain, msg.DestinationDomain)

	nonceBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(nonceBytes, msg.Nonce)

	copy(result[0:4], versionBytes)
	copy(result[4:8], sourceDomainBytes)
	copy(result[8:12], destinationDomain)
	copy(result[12:20], nonceBytes)
	copy(result[20:52], msg.Sender)
	copy(result[52:84], msg.Recipient)
	copy(result[84:116], msg.DestinationCaller)
	copy(result[116:], msg.MessageBody)

	return result
}

/*
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
	publicKeys []types.PublicKeys,
	signatureThreshold uint32) (bool, error) {

	if uint32(len(attestation)) != signatureLength*signatureThreshold {
		return false, sdkerrors.Wrap(types.ErrSignatureVerification, "invalid attestation length")
	}

	if signatureThreshold == 0 {
		return false, sdkerrors.Wrap(types.ErrSignatureVerification, "signature verification threshold cannot be 0")
	}

	// public keys cannot be empty, so the recovered key should be bigger than latestPublicKey
	var latestECDSA ecdsa.PublicKey

	digest := crypto.Keccak256(message)

	for i := uint32(0); i < signatureThreshold; i++ {
		signature := attestation[i*signatureLength : (i*signatureLength)+signatureLength]

		if signature[len(signature)-1] == 27 || signature[len(signature)-1] == 28 {
			signature[len(signature)-1] -= 27
		}

		recoveredKey, err := crypto.Ecrecover(digest, signature)
		if err != nil {
			return false, sdkerrors.Wrap(types.ErrSignatureVerification, "failed to recover public key")
		}

		// Signatures must be in increasing order of address, and may not duplicate signatures from same address.

		recoveredECSDA := ecdsa.PublicKey{
			X: new(big.Int).SetBytes(recoveredKey[1:33]),
			Y: new(big.Int).SetBytes(recoveredKey[33:]),
		}

		if latestECDSA.X != nil && latestECDSA.Y != nil && bytes.Compare(
			crypto.PubkeyToAddress(latestECDSA).Bytes(),
			crypto.PubkeyToAddress(recoveredECSDA).Bytes()) > -1 {
			return false, sdkerrors.Wrap(types.ErrSignatureVerification, "invalid signature order or dupe")
		}

		// check that recovered key is a valid
		contains := false
		for _, key := range publicKeys {
			hexBz, err := hex.DecodeString(key.Key)
			if err != nil {
				return false, sdkerrors.Wrap(types.ErrSignatureVerification, "failed to decode public key in module state")
			}
			if bytes.Equal(hexBz, recoveredKey) {
				contains = true
				break
			}
		}

		if !contains {
			return false, sdkerrors.Wrap(types.ErrSignatureVerification, "invalid signature: not an attester")
		}

		latestECDSA.X = recoveredECSDA.X
		latestECDSA.Y = recoveredECSDA.Y
	}
	return true, nil
}
