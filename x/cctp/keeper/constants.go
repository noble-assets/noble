package keeper

const nobleDomainId = 4
const nobleVersion = 0
const messageBodyIndex = 116
const burnMessageLength = 132
const signatureLength = 65

const MessageBodyVersion = 0

// TokenMessengerMap maps destinationDomain to TokenMessenger
var TokenMessengerMap = map[uint32]string{
	0: "0xbd3fa81b58ba92a82136038b25adec7066af3155", // ethereum mainnet
	1: "0x6b25532e1060ce10cc3b0a99e5683b91bfde6982", // avalanche mainnet
	3: "0x19330d10D9Cc8751218eaf51E8885D058642E08A", // arbitrum mainnet
	//0: "0xd0c3da58f55358142b8d3e06c1c30c5c6114efe8", // ethereum testnet
	//1: "0xeb08f243e5d3fcff26a9e38ae5520a669f4019d0", // avalanche testnet
	//3: "0x12dcfd3fe2e9eac2859fd1ed86d2ab8c5a2f9352", // arbitrum testnet
}
