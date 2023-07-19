package types

import (
	"encoding/binary"
	fmt "fmt"
)

// UnmarshalBytes takes a CCTP Burn Message payload and unmarshals it into the BurnMessage struct.
func (bm *BurnMessage) UnmarshalBytes(msg []byte) error {
	if len(msg) != 132 {
		return fmt.Errorf("invalid burn message length: %d, required: 132", len(msg))
	}
	bm.Version = binary.BigEndian.Uint32(msg[0:4])
	bm.BurnToken = msg[4:36]
	bm.MintRecipient = msg[36:68]
	bm.Amount = binary.BigEndian.Uint64(msg[92:100])
	bm.MessageSender = msg[100:132]

	return nil
}

// Bytes takes a BurnMessage struct and marshals it into a CCTP Burn Message payload.
func (bm *BurnMessage) Bytes() []byte {
	result := make([]byte, 132)

	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, bm.Version)

	amountBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(amountBytes, bm.Amount)
	amountBytesPadded := append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, amountBytes...)

	copy(result[0:4], versionBytes)
	copy(result[4:36], bm.BurnToken)
	copy(result[36:68], bm.MintRecipient)
	copy(result[68:100], amountBytesPadded)
	copy(result[100:132], bm.MessageSender)

	return result
}
