package types

import (
	"encoding/binary"
	fmt "fmt"
)

func (n *Nonce) Bytes() []byte {
	sourceDomainBz := make([]byte, 4)
	binary.BigEndian.PutUint32(sourceDomainBz, n.SourceDomain)

	nonceBz := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBz, n.Nonce)

	return append(sourceDomainBz, nonceBz...)
}

func (n *Nonce) UnmarshalBytes(bz []byte) error {
	if len(bz) != 12 {
		return fmt.Errorf("used nonce length %d is invalid, must be 12", len(bz))
	}

	n.SourceDomain = binary.BigEndian.Uint32(bz[0:4])
	n.Nonce = binary.BigEndian.Uint64(bz[4:12])

	return nil
}
