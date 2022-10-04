package types

const (
	// ModuleName defines the module name
	ModuleName = "blockibc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_blockibc"

	// Version defines the current version the IBC module supports
	Version = "blockibc-1"

	// PortID is the default port id that module binds to
	PortID = "blockibc"
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey = KeyPrefix("blockibc-port-")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
