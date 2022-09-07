package types

const (
	// ModuleName defines the module name
	ModuleName = "tokenfactory"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_tokenfactory"

	PausedKey       = "Paused/value/"
	MasterMinterKey = "MasterMinter/value/"
	PauserKey       = "Pauser/value/"
	BlacklisterKey  = "Blacklister/value/"
	OwnerKey        = "Owner/value/"
	AdminKey        = "Admin/value/"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
