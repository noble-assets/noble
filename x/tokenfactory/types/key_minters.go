package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// MintersKeyPrefix is the prefix to retrieve all Minters
	MintersKeyPrefix = "Minters/value/"
)

// MintersKey returns the store key to retrieve a Minters from the index fields
func MintersKey(
	address string,
) []byte {
	var key []byte

	addressBytes := []byte(address)
	key = append(key, addressBytes...)
	key = append(key, []byte("/")...)

	return key
}
