// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lease.proto

package leasepb

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/golang/protobuf/proto"
	etcdserverpb "go.etcd.io/etcd/api/v3/etcdserverpb"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Lease struct {
	ID                   int64    `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	TTL                  int64    `protobuf:"varint,2,opt,name=TTL,proto3" json:"TTL,omitempty"`
	RemainingTTL         int64    `protobuf:"varint,3,opt,name=RemainingTTL,proto3" json:"RemainingTTL,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Lease) Reset()         { *m = Lease{} }
func (m *Lease) String() string { return proto.CompactTextString(m) }
func (*Lease) ProtoMessage()    {}
func (*Lease) Descriptor() ([]byte, []int) {
	return fileDescriptor_3dd57e402472b33a, []int{0}
}
func (m *Lease) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Lease) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Lease.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Lease) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Lease.Merge(m, src)
}
func (m *Lease) XXX_Size() int {
	return m.Size()
}
func (m *Lease) XXX_DiscardUnknown() {
	xxx_messageInfo_Lease.DiscardUnknown(m)
}

var xxx_messageInfo_Lease proto.InternalMessageInfo

type LeaseInternalRequest struct {
	LeaseTimeToLiveRequest *etcdserverpb.LeaseTimeToLiveRequest `protobuf:"bytes,1,opt,name=LeaseTimeToLiveRequest,proto3" json:"LeaseTimeToLiveRequest,omitempty"`
	XXX_NoUnkeyedLiteral   struct{}                             `json:"-"`
	XXX_unrecognized       []byte                               `json:"-"`
	XXX_sizecache          int32                                `json:"-"`
}

func (m *LeaseInternalRequest) Reset()         { *m = LeaseInternalRequest{} }
func (m *LeaseInternalRequest) String() string { return proto.CompactTextString(m) }
func (*LeaseInternalRequest) ProtoMessage()    {}
func (*LeaseInternalRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3dd57e402472b33a, []int{1}
}
func (m *LeaseInternalRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *LeaseInternalRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_LeaseInternalRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *LeaseInternalRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LeaseInternalRequest.Merge(m, src)
}
func (m *LeaseInternalRequest) XXX_Size() int {
	return m.Size()
}
func (m *LeaseInternalRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_LeaseInternalRequest.DiscardUnknown(m)
}

var xxx_messageInfo_LeaseInternalRequest proto.InternalMessageInfo

type LeaseInternalResponse struct {
	LeaseTimeToLiveResponse *etcdserverpb.LeaseTimeToLiveResponse `protobuf:"bytes,1,opt,name=LeaseTimeToLiveResponse,proto3" json:"LeaseTimeToLiveResponse,omitempty"`
	XXX_NoUnkeyedLiteral    struct{}                              `json:"-"`
	XXX_unrecognized        []byte                                `json:"-"`
	XXX_sizecache           int32                                 `json:"-"`
}

func (m *LeaseInternalResponse) Reset()         { *m = LeaseInternalResponse{} }
func (m *LeaseInternalResponse) String() string { return proto.CompactTextString(m) }
func (*LeaseInternalResponse) ProtoMessage()    {}
func (*LeaseInternalResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3dd57e402472b33a, []int{2}
}
func (m *LeaseInternalResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *LeaseInternalResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_LeaseInternalResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *LeaseInternalResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LeaseInternalResponse.Merge(m, src)
}
func (m *LeaseInternalResponse) XXX_Size() int {
	return m.Size()
}
func (m *LeaseInternalResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_LeaseInternalResponse.DiscardUnknown(m)
}

var xxx_messageInfo_LeaseInternalResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Lease)(nil), "leasepb.Lease")
	proto.RegisterType((*LeaseInternalRequest)(nil), "leasepb.LeaseInternalRequest")
	proto.RegisterType((*LeaseInternalResponse)(nil), "leasepb.LeaseInternalResponse")
}

func init() { proto.RegisterFile("lease.proto", fileDescriptor_3dd57e402472b33a) }

var fileDescriptor_3dd57e402472b33a = []byte{
	// 256 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xce, 0x49, 0x4d, 0x2c,
	0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x07, 0x73, 0x0a, 0x92, 0xa4, 0x44, 0xd2,
	0xf3, 0xd3, 0xf3, 0xc1, 0x62, 0xfa, 0x20, 0x16, 0x44, 0x5a, 0x4a, 0x3e, 0xb5, 0x24, 0x39, 0x45,
	0x3f, 0xb1, 0x20, 0x53, 0x1f, 0xc4, 0x28, 0x4e, 0x2d, 0x2a, 0x4b, 0x2d, 0x2a, 0x48, 0xd2, 0x2f,
	0x2a, 0x48, 0x86, 0x28, 0x50, 0xf2, 0xe5, 0x62, 0xf5, 0x01, 0x99, 0x20, 0xc4, 0xc7, 0xc5, 0xe4,
	0xe9, 0x22, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0x1c, 0xc4, 0xe4, 0xe9, 0x22, 0x24, 0xc0, 0xc5, 0x1c,
	0x12, 0xe2, 0x23, 0xc1, 0x04, 0x16, 0x00, 0x31, 0x85, 0x94, 0xb8, 0x78, 0x82, 0x52, 0x73, 0x13,
	0x33, 0xf3, 0x32, 0xf3, 0xd2, 0x41, 0x52, 0xcc, 0x60, 0x29, 0x14, 0x31, 0xa5, 0x12, 0x2e, 0x11,
	0xb0, 0x71, 0x9e, 0x79, 0x25, 0xa9, 0x45, 0x79, 0x89, 0x39, 0x41, 0xa9, 0x85, 0xa5, 0xa9, 0xc5,
	0x25, 0x42, 0x31, 0x5c, 0x62, 0x60, 0xf1, 0x90, 0xcc, 0xdc, 0xd4, 0x90, 0x7c, 0x9f, 0xcc, 0xb2,
	0x54, 0xa8, 0x0c, 0xd8, 0x46, 0x6e, 0x23, 0x15, 0x3d, 0x64, 0xf7, 0xe9, 0x61, 0x57, 0x1b, 0x84,
	0xc3, 0x0c, 0xa5, 0x0a, 0x2e, 0x51, 0x34, 0x5b, 0x8b, 0x0b, 0xf2, 0xf3, 0x8a, 0x53, 0x85, 0xe2,
	0xb9, 0xc4, 0x31, 0xb4, 0x40, 0xa4, 0xa0, 0xf6, 0xaa, 0x12, 0xb0, 0x17, 0xa2, 0x38, 0x08, 0x97,
	0x29, 0x4e, 0x12, 0x27, 0x1e, 0xca, 0x31, 0x5c, 0x78, 0x28, 0xc7, 0x70, 0xe2, 0x91, 0x1c, 0xe3,
	0x85, 0x47, 0x72, 0x8c, 0x0f, 0x1e, 0xc9, 0x31, 0xce, 0x78, 0x2c, 0xc7, 0x90, 0xc4, 0x06, 0x0e,
	0x5f, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x7b, 0x8a, 0x94, 0xb9, 0xae, 0x01, 0x00, 0x00,
}

func (m *Lease) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Lease) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Lease) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.RemainingTTL != 0 {
		i = encodeVarintLease(dAtA, i, uint64(m.RemainingTTL))
		i--
		dAtA[i] = 0x18
	}
	if m.TTL != 0 {
		i = encodeVarintLease(dAtA, i, uint64(m.TTL))
		i--
		dAtA[i] = 0x10
	}
	if m.ID != 0 {
		i = encodeVarintLease(dAtA, i, uint64(m.ID))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *LeaseInternalRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LeaseInternalRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *LeaseInternalRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.LeaseTimeToLiveRequest != nil {
		{
			size, err := m.LeaseTimeToLiveRequest.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintLease(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *LeaseInternalResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LeaseInternalResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *LeaseInternalResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.LeaseTimeToLiveResponse != nil {
		{
			size, err := m.LeaseTimeToLiveResponse.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintLease(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintLease(dAtA []byte, offset int, v uint64) int {
	offset -= sovLease(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Lease) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ID != 0 {
		n += 1 + sovLease(uint64(m.ID))
	}
	if m.TTL != 0 {
		n += 1 + sovLease(uint64(m.TTL))
	}
	if m.RemainingTTL != 0 {
		n += 1 + sovLease(uint64(m.RemainingTTL))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *LeaseInternalRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.LeaseTimeToLiveRequest != nil {
		l = m.LeaseTimeToLiveRequest.Size()
		n += 1 + l + sovLease(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *LeaseInternalResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.LeaseTimeToLiveResponse != nil {
		l = m.LeaseTimeToLiveResponse.Size()
		n += 1 + l + sovLease(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovLease(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozLease(x uint64) (n int) {
	return sovLease(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Lease) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLease
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
			return fmt.Errorf("proto: Lease: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Lease: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			m.ID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLease
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ID |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TTL", wireType)
			}
			m.TTL = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLease
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TTL |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RemainingTTL", wireType)
			}
			m.RemainingTTL = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLease
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RemainingTTL |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipLease(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLease
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *LeaseInternalRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLease
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
			return fmt.Errorf("proto: LeaseInternalRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LeaseInternalRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LeaseTimeToLiveRequest", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLease
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthLease
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLease
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.LeaseTimeToLiveRequest == nil {
				m.LeaseTimeToLiveRequest = &etcdserverpb.LeaseTimeToLiveRequest{}
			}
			if err := m.LeaseTimeToLiveRequest.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLease(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLease
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *LeaseInternalResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLease
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
			return fmt.Errorf("proto: LeaseInternalResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LeaseInternalResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LeaseTimeToLiveResponse", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLease
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthLease
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLease
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.LeaseTimeToLiveResponse == nil {
				m.LeaseTimeToLiveResponse = &etcdserverpb.LeaseTimeToLiveResponse{}
			}
			if err := m.LeaseTimeToLiveResponse.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLease(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLease
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipLease(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLease
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
					return 0, ErrIntOverflowLease
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
					return 0, ErrIntOverflowLease
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
				return 0, ErrInvalidLengthLease
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupLease
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthLease
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthLease        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLease          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupLease = fmt.Errorf("proto: unexpected end of group")
)
