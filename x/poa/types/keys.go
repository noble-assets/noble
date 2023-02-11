package types

var (
	ValidatorsKey       = []byte{0x25} // prefix for each key to a validator
	VotesKey            = []byte{0x26} // prefix for each key to a vote
	VotesByValidatorKey = []byte{0x27} // prefix for each key to a validator
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
)
