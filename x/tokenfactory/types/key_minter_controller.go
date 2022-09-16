package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// MinterControllerKeyPrefix is the prefix to retrieve all MinterController
	MinterControllerKeyPrefix = "MinterController/value/"
)

// MinterControllerKey returns the store key to retrieve a MinterController from the index fields
func MinterControllerKey(
	minterAddress string,
) []byte {
	var key []byte

	minterAddressBytes := []byte(minterAddress)
	key = append(key, minterAddressBytes...)
	key = append(key, []byte("/")...)

	return key
}
