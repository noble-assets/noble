package types

const (
	// ModuleName defines the module name
	ModuleName = "usdctokenfactory"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName

	PausedKey                 = "Paused/value/"
	MasterMinterKey           = "MasterMinter/value/"
	PauserKey                 = "Pauser/value/"
	BlacklisterKey            = "Blacklister/value/"
	OwnerKey                  = "Owner/value/"
	PendingOwnerKey           = "PendingOwner/value/"
	BlacklistedKeyPrefix      = "Blacklisted/value/"
	MintersKeyPrefix          = "Minters/value/"
	MinterControllerKeyPrefix = "MinterController/value/"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// BlacklistedKey returns the store key to retrieve a Blacklisted from the index fields
func BlacklistedKey(addressBz []byte) []byte {
	return append(addressBz, []byte("/")...)
}

// MintersKey returns the store key to retrieve a Minters from the index fields
func MintersKey(address string) []byte {
	return append([]byte(address), []byte("/")...)
}

// MinterControllerKey returns the store key to retrieve a MinterController from the index fields
func MinterControllerKey(controllerAddress string) []byte {
	return append([]byte(controllerAddress), []byte("/")...)

}

const (
	MintingDenomKey = "MintingDenom/value/"
)
