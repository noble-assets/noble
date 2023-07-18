// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: router/in_flight_packet.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// InFlightPacket contains information about the initially minted funds
// @param source_domain_sender
// @param nonce
type InFlightPacket struct {
	SourceDomain       uint32 `protobuf:"varint,1,opt,name=source_domain,json=sourceDomain,proto3" json:"source_domain,omitempty"`
	SourceDomainSender string `protobuf:"bytes,2,opt,name=source_domain_sender,json=sourceDomainSender,proto3" json:"source_domain_sender,omitempty"`
	Nonce              uint64 `protobuf:"varint,3,opt,name=nonce,proto3" json:"nonce,omitempty"`
	ChannelId          string `protobuf:"bytes,4,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	PortId             string `protobuf:"bytes,5,opt,name=port_id,json=portId,proto3" json:"port_id,omitempty"`
	Sequence           uint64 `protobuf:"varint,6,opt,name=sequence,proto3" json:"sequence,omitempty"`
}

func (m *InFlightPacket) Reset()         { *m = InFlightPacket{} }
func (m *InFlightPacket) String() string { return proto.CompactTextString(m) }
func (*InFlightPacket) ProtoMessage()    {}
func (*InFlightPacket) Descriptor() ([]byte, []int) {
	return fileDescriptor_93456951126717f2, []int{0}
}
func (m *InFlightPacket) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InFlightPacket) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InFlightPacket.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InFlightPacket) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InFlightPacket.Merge(m, src)
}
func (m *InFlightPacket) XXX_Size() int {
	return m.Size()
}
func (m *InFlightPacket) XXX_DiscardUnknown() {
	xxx_messageInfo_InFlightPacket.DiscardUnknown(m)
}

var xxx_messageInfo_InFlightPacket proto.InternalMessageInfo

func (m *InFlightPacket) GetSourceDomain() uint32 {
	if m != nil {
		return m.SourceDomain
	}
	return 0
}

func (m *InFlightPacket) GetSourceDomainSender() string {
	if m != nil {
		return m.SourceDomainSender
	}
	return ""
}

func (m *InFlightPacket) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func (m *InFlightPacket) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

func (m *InFlightPacket) GetPortId() string {
	if m != nil {
		return m.PortId
	}
	return ""
}

func (m *InFlightPacket) GetSequence() uint64 {
	if m != nil {
		return m.Sequence
	}
	return 0
}

func init() {
	proto.RegisterType((*InFlightPacket)(nil), "noble.router.InFlightPacket")
}

func init() { proto.RegisterFile("router/in_flight_packet.proto", fileDescriptor_93456951126717f2) }

var fileDescriptor_93456951126717f2 = []byte{
	// 311 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0xbf, 0x4e, 0x02, 0x41,
	0x10, 0xc6, 0x59, 0x05, 0x94, 0x0d, 0x58, 0x6c, 0x48, 0xbc, 0x90, 0xb0, 0x21, 0xda, 0xd0, 0xc0,
	0x4a, 0x2c, 0xed, 0x8c, 0x31, 0xa1, 0x33, 0x18, 0x1b, 0x9b, 0xcb, 0xdd, 0xde, 0x78, 0x5c, 0x3c,
	0x66, 0xce, 0xdd, 0x3d, 0xa2, 0x6f, 0xe1, 0x63, 0x59, 0x62, 0x67, 0x69, 0xe0, 0x45, 0xcc, 0xed,
	0x11, 0xa3, 0xdd, 0x7c, 0x7f, 0xe6, 0x57, 0x7c, 0x7c, 0x68, 0xa8, 0x74, 0x60, 0x54, 0x86, 0xe1,
	0x53, 0x9e, 0xa5, 0x4b, 0x17, 0x16, 0x91, 0x7e, 0x06, 0x37, 0x2d, 0x0c, 0x39, 0x12, 0x5d, 0xa4,
	0x38, 0x87, 0x69, 0x5d, 0x1a, 0x48, 0x4d, 0x76, 0x45, 0x56, 0xc5, 0x91, 0x05, 0xb5, 0x9e, 0xc5,
	0xe0, 0xa2, 0x99, 0xd2, 0x94, 0x61, 0xdd, 0x1e, 0xf4, 0x53, 0x4a, 0xc9, 0x9f, 0xaa, 0xba, 0x6a,
	0xf7, 0xec, 0x93, 0xf1, 0x93, 0x39, 0xde, 0x7a, 0xfa, 0x9d, 0x87, 0x8b, 0x73, 0xde, 0xb3, 0x54,
	0x1a, 0x0d, 0x61, 0x42, 0xab, 0x28, 0xc3, 0x80, 0x8d, 0xd8, 0xb8, 0xb7, 0xe8, 0xd6, 0xe6, 0x8d,
	0xf7, 0xc4, 0x05, 0xef, 0xff, 0x2b, 0x85, 0x16, 0x30, 0x01, 0x13, 0x1c, 0x8c, 0xd8, 0xb8, 0xb3,
	0x10, 0x7f, 0xbb, 0xf7, 0x3e, 0x11, 0x7d, 0xde, 0x42, 0x42, 0x0d, 0xc1, 0xe1, 0x88, 0x8d, 0x9b,
	0x8b, 0x5a, 0x88, 0x21, 0xe7, 0x7a, 0x19, 0x21, 0x42, 0x1e, 0x66, 0x49, 0xd0, 0xf4, 0xdf, 0x9d,
	0xbd, 0x33, 0x4f, 0xc4, 0x29, 0x3f, 0x2a, 0xc8, 0xb8, 0x2a, 0x6b, 0xf9, 0xac, 0x5d, 0xc9, 0x79,
	0x22, 0x06, 0xfc, 0xd8, 0xc2, 0x4b, 0x09, 0x15, 0xb0, 0xed, 0x81, 0xbf, 0xfa, 0xfa, 0xe1, 0x63,
	0x2b, 0xd9, 0x66, 0x2b, 0xd9, 0xf7, 0x56, 0xb2, 0xf7, 0x9d, 0x6c, 0x6c, 0x76, 0xb2, 0xf1, 0xb5,
	0x93, 0x8d, 0xc7, 0xab, 0x34, 0x73, 0xcb, 0x32, 0x9e, 0x6a, 0x5a, 0x29, 0xeb, 0x4c, 0x84, 0x29,
	0xe4, 0xb4, 0x86, 0xc9, 0x1a, 0xd0, 0x95, 0x06, 0xac, 0xf2, 0x8b, 0x4e, 0xf6, 0xb3, 0xbf, 0xaa,
	0xfd, 0xe1, 0xde, 0x0a, 0xb0, 0x71, 0xdb, 0x2f, 0x76, 0xf9, 0x13, 0x00, 0x00, 0xff, 0xff, 0xdc,
	0x2f, 0xcc, 0xfe, 0x96, 0x01, 0x00, 0x00,
}

func (m *InFlightPacket) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InFlightPacket) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InFlightPacket) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Sequence != 0 {
		i = encodeVarintInFlightPacket(dAtA, i, uint64(m.Sequence))
		i--
		dAtA[i] = 0x30
	}
	if len(m.PortId) > 0 {
		i -= len(m.PortId)
		copy(dAtA[i:], m.PortId)
		i = encodeVarintInFlightPacket(dAtA, i, uint64(len(m.PortId)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.ChannelId) > 0 {
		i -= len(m.ChannelId)
		copy(dAtA[i:], m.ChannelId)
		i = encodeVarintInFlightPacket(dAtA, i, uint64(len(m.ChannelId)))
		i--
		dAtA[i] = 0x22
	}
	if m.Nonce != 0 {
		i = encodeVarintInFlightPacket(dAtA, i, uint64(m.Nonce))
		i--
		dAtA[i] = 0x18
	}
	if len(m.SourceDomainSender) > 0 {
		i -= len(m.SourceDomainSender)
		copy(dAtA[i:], m.SourceDomainSender)
		i = encodeVarintInFlightPacket(dAtA, i, uint64(len(m.SourceDomainSender)))
		i--
		dAtA[i] = 0x12
	}
	if m.SourceDomain != 0 {
		i = encodeVarintInFlightPacket(dAtA, i, uint64(m.SourceDomain))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintInFlightPacket(dAtA []byte, offset int, v uint64) int {
	offset -= sovInFlightPacket(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *InFlightPacket) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.SourceDomain != 0 {
		n += 1 + sovInFlightPacket(uint64(m.SourceDomain))
	}
	l = len(m.SourceDomainSender)
	if l > 0 {
		n += 1 + l + sovInFlightPacket(uint64(l))
	}
	if m.Nonce != 0 {
		n += 1 + sovInFlightPacket(uint64(m.Nonce))
	}
	l = len(m.ChannelId)
	if l > 0 {
		n += 1 + l + sovInFlightPacket(uint64(l))
	}
	l = len(m.PortId)
	if l > 0 {
		n += 1 + l + sovInFlightPacket(uint64(l))
	}
	if m.Sequence != 0 {
		n += 1 + sovInFlightPacket(uint64(m.Sequence))
	}
	return n
}

func sovInFlightPacket(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozInFlightPacket(x uint64) (n int) {
	return sovInFlightPacket(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *InFlightPacket) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowInFlightPacket
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: InFlightPacket: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InFlightPacket: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SourceDomain", wireType)
			}
			m.SourceDomain = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInFlightPacket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SourceDomain |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SourceDomainSender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInFlightPacket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInFlightPacket
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInFlightPacket
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SourceDomainSender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nonce", wireType)
			}
			m.Nonce = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInFlightPacket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Nonce |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChannelId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInFlightPacket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInFlightPacket
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInFlightPacket
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChannelId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PortId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInFlightPacket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInFlightPacket
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInFlightPacket
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PortId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sequence", wireType)
			}
			m.Sequence = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInFlightPacket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Sequence |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipInFlightPacket(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthInFlightPacket
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipInFlightPacket(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowInFlightPacket
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowInFlightPacket
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowInFlightPacket
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthInFlightPacket
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupInFlightPacket
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthInFlightPacket
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthInFlightPacket        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowInFlightPacket          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupInFlightPacket = fmt.Errorf("proto: unexpected end of group")
)