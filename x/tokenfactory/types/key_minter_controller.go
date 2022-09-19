package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// MinterControllerKeyPrefix is the prefix to retrieve all MinterController
	MinterControllerKeyPrefix = "MinterController/value/"
)

// MinterControllerKey returns the store key to retrieve a MinterController from the index fields
func MinterControllerKey(
	controllerAddress string,
) []byte {
	var key []byte

	controllerAddressBytes := []byte(controllerAddress)
	key = append(key, controllerAddressBytes...)
	key = append(key, []byte("/")...)

	return key
}
