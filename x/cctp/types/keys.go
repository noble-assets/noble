package types

import (
	"encoding/binary"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// ModuleName defines the module name
	ModuleName = "cctp"

	// StoreKey defines the primary module store key
	StoreKey = "cctp"

	// RouterKey defines the module's message routing key
	RouterKey = StoreKey

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + StoreKey

	AuthorityKey                         = "Authority/value/"
	BurningAndMintingPausedKey           = "BurningAndMintingPaused/value/"
	MaxMessageBodySizeKey                = "MaxMessageBodySize/value/"
	MinterAllowanceKeyPrefix             = "MinterAllowance/value/"
	NonceKeyPrefix                       = "Nonce/value/"
	PerMessageBurnLimitKey               = "PerMessageBurnLimit/value/"
	AttesterKeyPrefix                    = "Attester/value/"
	SendingAndReceivingMessagesPausedKey = "SendingAndReceivingMessagesPaused/value/"
	SignatureThresholdKeyPrefix          = "SignatureThreshold/value/"
	TokenPairKeyPrefix                   = "TokenPair/value/"
	UsedNonceKeyPrefix                   = "UsedNonce/value/"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func SignatureThresholdKey(key []byte) []byte {
	return append(key, []byte("/")...)
}

// PublicKeyKey returns the store key to retrieve a PublicKey from the index fields
func PublicKeyKey(key []byte) []byte {
	return append(key, []byte("/")...)
}

// UsedNonceKey returns the store key to retrieve a UsedNonce from the index fields
func UsedNonceKey(key []byte) []byte {
	return append(key, []byte("/")...)
}

// TokenPairKey returns the store key to retrieve a TokenPair from the index fields
func TokenPairKey(remoteDomain uint32, remoteToken string) []byte {

	remoteDomainBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(remoteDomainBytes, remoteDomain)

	combinedBytes := append(remoteDomainBytes, []byte(strings.ToLower(remoteToken))...)
	hashedKey := crypto.Keccak256(combinedBytes)

	return append(hashedKey, []byte("/")...)
}

func MinterAllowanceKey(key []byte) []byte {
	return append(key, []byte("/")...)
}
