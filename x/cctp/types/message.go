package types

import (
	"encoding/binary"
	fmt "fmt"
)

// UnmarshalBytes takes a CCTP Message payload and unmarshals it into the Message struct.
func (m *Message) UnmarshalBytes(msg []byte) error {
	if len(msg) < 116 {
		return fmt.Errorf("invalid message: %d bytes is too short, must be at least 116 bytes", len(msg))
	}
	m.Version = binary.BigEndian.Uint32(msg[0:4])
	m.SourceDomainBytes = msg[4:8]
	m.SourceDomain = binary.BigEndian.Uint32(msg[4:8])
	m.DestinationDomain = binary.BigEndian.Uint32(msg[8:12])
	m.NonceBytes = msg[12:20]
	m.Nonce = binary.BigEndian.Uint64(msg[12:20])
	m.Sender = msg[20:52]
	m.Recipient = msg[52:84]
	m.DestinationCaller = msg[84:116]
	m.MessageBody = msg[116:]

	return nil
}

// Bytes takes a Message struct and marshals it into a CCTP Message payload.
func (m *Message) Bytes() []byte {
	result := make([]byte, 116+len(m.MessageBody))

	versionBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(versionBytes, m.Version)

	sourceDomainBytes := make([]byte, 32)
	binary.BigEndian.PutUint32(sourceDomainBytes, m.SourceDomain)

	destinationDomain := make([]byte, 32)
	binary.BigEndian.PutUint32(destinationDomain, m.DestinationDomain)

	nonceBytes := make([]byte, 32)
	binary.BigEndian.PutUint64(nonceBytes, m.Nonce)

	copy(result[0:4], versionBytes)
	copy(result[4:8], sourceDomainBytes)
	copy(result[8:12], destinationDomain)
	copy(result[12:20], nonceBytes)
	copy(result[20:52], m.Sender)
	copy(result[52:84], m.Recipient)
	copy(result[84:116], m.DestinationCaller)
	copy(result[116:], m.MessageBody)

	return result
}
