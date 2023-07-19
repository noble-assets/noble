package keeper

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

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
