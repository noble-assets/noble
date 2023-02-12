package types

import "strconv"

var (
	ValidatorsKey       = []byte{0x25} // prefix for each key to a validator account key
	ValidatorsByConsKey = []byte{0x26} // prefix for each key to a validator consensus key
	VotesKey            = []byte{0x27} // prefix for each key to a vote
	VotesByValidatorKey = []byte{0x28} // prefix for each key to a validator
	HistoricalInfoKey   = []byte{0x29} // prefix for historical entries
)

const (
	// ModuleName is the name of the module
	ModuleName = "poa"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for querier msgs
	QuerierRoute = ModuleName

	// TransientStoreKey defines the transient store key
	TransientStoreKey = "transient_poa"
)

// GetHistoricalInfoKey returns a key prefix for indexing HistoricalInfo objects.
func GetHistoricalInfoKey(height int64) []byte {
	return append(HistoricalInfoKey, []byte(strconv.FormatInt(height, 10))...)
}
