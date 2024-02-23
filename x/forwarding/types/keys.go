package types

const (
	ModuleName        = "forwarding"
	StoreKey          = "forwarding"
	TransientStoreKey = "transient_forwarding"
)

var (
	NumOfAccountsPrefix   = []byte("num_of_accounts")
	NumOfForwardsPrefix   = []byte("num_of_forwards")
	TotalForwardedPrefix  = []byte("total_forwarded")
	PendingForwardsPrefix = []byte("pending_forwards")
)

func NumOfAccountsKey(channel string) []byte {
	return append(NumOfAccountsPrefix, []byte(channel)...)
}

func NumOfForwardsKey(channel string) []byte {
	return append(NumOfForwardsPrefix, []byte(channel)...)
}

func TotalForwardedKey(channel string) []byte {
	return append(TotalForwardedPrefix, []byte(channel)...)
}

func PendingForwardsKey(account *ForwardingAccount) []byte {
	return append(PendingForwardsPrefix, account.GetAddress()...)
}
