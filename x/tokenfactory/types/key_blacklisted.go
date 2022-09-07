package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// BlacklistedKeyPrefix is the prefix to retrieve all Blacklisted
	BlacklistedKeyPrefix = "Blacklisted/value/"
)

// BlacklistedKey returns the store key to retrieve a Blacklisted from the index fields
func BlacklistedKey(address string) []byte {
	return append([]byte(address), []byte("/")...)
}
