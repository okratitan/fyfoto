// Code generated by protoc-gen-go. DO NOT EDIT.
// source: crypto.proto

package cryptogo

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
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

type CompressionAlgorithm int32

const (
	CompressionAlgorithm_UNKNOWN_COMPRESSION CompressionAlgorithm = 0
)

var CompressionAlgorithm_name = map[int32]string{
	0: "UNKNOWN_COMPRESSION",
}

var CompressionAlgorithm_value = map[string]int32{
	"UNKNOWN_COMPRESSION": 0,
}

func (x CompressionAlgorithm) String() string {
	return proto.EnumName(CompressionAlgorithm_name, int32(x))
}

func (CompressionAlgorithm) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_527278fb02d03321, []int{0}
}

type EncryptionAlgorithm int32

const (
	EncryptionAlgorithm_UNKNOWN_ENCRYPTION    EncryptionAlgorithm = 0
	EncryptionAlgorithm_AES_128_GCM_NOPADDING EncryptionAlgorithm = 1
	EncryptionAlgorithm_PBKDF2WITHHMACSHA1    EncryptionAlgorithm = 2
	EncryptionAlgorithm_RSA_ECB_OAEPPADDING   EncryptionAlgorithm = 3
	EncryptionAlgorithm_AES_256_GCM_NOPADDING EncryptionAlgorithm = 4
)

var EncryptionAlgorithm_name = map[int32]string{
	0: "UNKNOWN_ENCRYPTION",
	1: "AES_128_GCM_NOPADDING",
	2: "PBKDF2WITHHMACSHA1",
	3: "RSA_ECB_OAEPPADDING",
	4: "AES_256_GCM_NOPADDING",
}

var EncryptionAlgorithm_value = map[string]int32{
	"UNKNOWN_ENCRYPTION":    0,
	"AES_128_GCM_NOPADDING": 1,
	"PBKDF2WITHHMACSHA1":    2,
	"RSA_ECB_OAEPPADDING":   3,
	"AES_256_GCM_NOPADDING": 4,
}

func (x EncryptionAlgorithm) String() string {
	return proto.EnumName(EncryptionAlgorithm_name, int32(x))
}

func (EncryptionAlgorithm) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_527278fb02d03321, []int{1}
}

type SignatureAlgorithm int32

const (
	SignatureAlgorithm_UNKNOWN_SIGNATURE SignatureAlgorithm = 0
	SignatureAlgorithm_SHA512WITHRSA     SignatureAlgorithm = 1
	SignatureAlgorithm_SHA512WITHRSA_PSS SignatureAlgorithm = 2
)

var SignatureAlgorithm_name = map[int32]string{
	0: "UNKNOWN_SIGNATURE",
	1: "SHA512WITHRSA",
	2: "SHA512WITHRSA_PSS",
}

var SignatureAlgorithm_value = map[string]int32{
	"UNKNOWN_SIGNATURE": 0,
	"SHA512WITHRSA":     1,
	"SHA512WITHRSA_PSS": 2,
}

func (x SignatureAlgorithm) String() string {
	return proto.EnumName(SignatureAlgorithm_name, int32(x))
}

func (SignatureAlgorithm) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_527278fb02d03321, []int{2}
}

type PublicKeyFormat int32

const (
	PublicKeyFormat_UNKNOWN_PUBLIC_KEY_FORMAT PublicKeyFormat = 0
	PublicKeyFormat_PKCS1_PUBLIC              PublicKeyFormat = 1
	PublicKeyFormat_PKIX                      PublicKeyFormat = 2
	PublicKeyFormat_X509                      PublicKeyFormat = 3
)

var PublicKeyFormat_name = map[int32]string{
	0: "UNKNOWN_PUBLIC_KEY_FORMAT",
	1: "PKCS1_PUBLIC",
	2: "PKIX",
	3: "X509",
}

var PublicKeyFormat_value = map[string]int32{
	"UNKNOWN_PUBLIC_KEY_FORMAT": 0,
	"PKCS1_PUBLIC":              1,
	"PKIX":                      2,
	"X509":                      3,
}

func (x PublicKeyFormat) String() string {
	return proto.EnumName(PublicKeyFormat_name, int32(x))
}

func (PublicKeyFormat) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_527278fb02d03321, []int{3}
}

type PrivateKeyFormat int32

const (
	PrivateKeyFormat_UNKNOWN_PRIVATE_KEY_FORMAT PrivateKeyFormat = 0
	PrivateKeyFormat_PKCS1_PRIVATE              PrivateKeyFormat = 1
	PrivateKeyFormat_PKCS8                      PrivateKeyFormat = 2
)

var PrivateKeyFormat_name = map[int32]string{
	0: "UNKNOWN_PRIVATE_KEY_FORMAT",
	1: "PKCS1_PRIVATE",
	2: "PKCS8",
}

var PrivateKeyFormat_value = map[string]int32{
	"UNKNOWN_PRIVATE_KEY_FORMAT": 0,
	"PKCS1_PRIVATE":              1,
	"PKCS8":                      2,
}

func (x PrivateKeyFormat) String() string {
	return proto.EnumName(PrivateKeyFormat_name, int32(x))
}

func (PrivateKeyFormat) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_527278fb02d03321, []int{4}
}

type KeyShare struct {
	Name                 string           `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	PublicKey            []byte           `protobuf:"bytes,2,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	PublicFormat         PublicKeyFormat  `protobuf:"varint,3,opt,name=public_format,json=publicFormat,proto3,enum=crypto.PublicKeyFormat" json:"public_format,omitempty"`
	PrivateKey           []byte           `protobuf:"bytes,4,opt,name=private_key,json=privateKey,proto3" json:"private_key,omitempty"`
	PrivateFormat        PrivateKeyFormat `protobuf:"varint,5,opt,name=private_format,json=privateFormat,proto3,enum=crypto.PrivateKeyFormat" json:"private_format,omitempty"`
	Password             []byte           `protobuf:"bytes,6,opt,name=password,proto3" json:"password,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *KeyShare) Reset()         { *m = KeyShare{} }
func (m *KeyShare) String() string { return proto.CompactTextString(m) }
func (*KeyShare) ProtoMessage()    {}
func (*KeyShare) Descriptor() ([]byte, []int) {
	return fileDescriptor_527278fb02d03321, []int{0}
}

func (m *KeyShare) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeyShare.Unmarshal(m, b)
}
func (m *KeyShare) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeyShare.Marshal(b, m, deterministic)
}
func (m *KeyShare) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeyShare.Merge(m, src)
}
func (m *KeyShare) XXX_Size() int {
	return xxx_messageInfo_KeyShare.Size(m)
}
func (m *KeyShare) XXX_DiscardUnknown() {
	xxx_messageInfo_KeyShare.DiscardUnknown(m)
}

var xxx_messageInfo_KeyShare proto.InternalMessageInfo

func (m *KeyShare) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *KeyShare) GetPublicKey() []byte {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

func (m *KeyShare) GetPublicFormat() PublicKeyFormat {
	if m != nil {
		return m.PublicFormat
	}
	return PublicKeyFormat_UNKNOWN_PUBLIC_KEY_FORMAT
}

func (m *KeyShare) GetPrivateKey() []byte {
	if m != nil {
		return m.PrivateKey
	}
	return nil
}

func (m *KeyShare) GetPrivateFormat() PrivateKeyFormat {
	if m != nil {
		return m.PrivateFormat
	}
	return PrivateKeyFormat_UNKNOWN_PRIVATE_KEY_FORMAT
}

func (m *KeyShare) GetPassword() []byte {
	if m != nil {
		return m.Password
	}
	return nil
}

func init() {
	proto.RegisterEnum("crypto.CompressionAlgorithm", CompressionAlgorithm_name, CompressionAlgorithm_value)
	proto.RegisterEnum("crypto.EncryptionAlgorithm", EncryptionAlgorithm_name, EncryptionAlgorithm_value)
	proto.RegisterEnum("crypto.SignatureAlgorithm", SignatureAlgorithm_name, SignatureAlgorithm_value)
	proto.RegisterEnum("crypto.PublicKeyFormat", PublicKeyFormat_name, PublicKeyFormat_value)
	proto.RegisterEnum("crypto.PrivateKeyFormat", PrivateKeyFormat_name, PrivateKeyFormat_value)
	proto.RegisterType((*KeyShare)(nil), "crypto.KeyShare")
}

func init() { proto.RegisterFile("crypto.proto", fileDescriptor_527278fb02d03321) }

var fileDescriptor_527278fb02d03321 = []byte{
	// 508 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x93, 0xdf, 0x6e, 0xda, 0x30,
	0x14, 0x87, 0x1b, 0x4a, 0x2b, 0x38, 0x85, 0xce, 0xb8, 0xeb, 0x80, 0x4a, 0xdd, 0xd0, 0xae, 0x10,
	0x17, 0x74, 0x30, 0x31, 0x75, 0xd2, 0xa4, 0xc9, 0x84, 0x00, 0x51, 0x46, 0x62, 0xc5, 0xd0, 0x3f,
	0xbb, 0x89, 0x52, 0xe6, 0x41, 0x34, 0x82, 0xa3, 0x90, 0xae, 0xe2, 0x3d, 0xf6, 0xb2, 0xbb, 0x9b,
	0xe2, 0x24, 0x4c, 0x70, 0x15, 0xfb, 0x9c, 0xef, 0x7c, 0x3f, 0x4b, 0x76, 0xa0, 0x34, 0x0f, 0xb7,
	0x41, 0x24, 0xda, 0x41, 0x28, 0x22, 0x81, 0x4f, 0x93, 0xdd, 0xfb, 0xbf, 0x0a, 0x14, 0x0c, 0xbe,
	0x65, 0x4b, 0x37, 0xe4, 0x18, 0x43, 0x7e, 0xed, 0xfa, 0xbc, 0xa6, 0x34, 0x94, 0x66, 0xd1, 0x96,
	0x6b, 0x7c, 0x0d, 0x10, 0x3c, 0x3f, 0xad, 0xbc, 0xb9, 0xf3, 0x8b, 0x6f, 0x6b, 0xb9, 0x86, 0xd2,
	0x2c, 0xd9, 0xc5, 0xa4, 0x62, 0xf0, 0x2d, 0xfe, 0x02, 0xe5, 0xb4, 0xfd, 0x53, 0x84, 0xbe, 0x1b,
	0xd5, 0x8e, 0x1b, 0x4a, 0xf3, 0xbc, 0x5b, 0x6d, 0xa7, 0x69, 0x34, 0x23, 0x87, 0xb2, 0x6d, 0x97,
	0x12, 0x3a, 0xd9, 0xe1, 0x77, 0x70, 0x16, 0x84, 0xde, 0x6f, 0x37, 0xe2, 0xd2, 0x9e, 0x97, 0x76,
	0x48, 0x4b, 0xb1, 0xfe, 0x2b, 0x9c, 0x67, 0x40, 0xea, 0x3f, 0x91, 0xfe, 0xda, 0xce, 0xbf, 0x63,
	0xd3, 0x80, 0x72, 0xca, 0xa7, 0x09, 0x57, 0x50, 0x08, 0xdc, 0xcd, 0xe6, 0x45, 0x84, 0x3f, 0x6a,
	0xa7, 0x52, 0xbf, 0xdb, 0xb7, 0x6e, 0xe0, 0xb5, 0x2a, 0xfc, 0x20, 0xe4, 0x9b, 0x8d, 0x27, 0xd6,
	0x64, 0xb5, 0x10, 0xa1, 0x17, 0x2d, 0x7d, 0x5c, 0x85, 0x8b, 0x99, 0x69, 0x98, 0xd6, 0xbd, 0xe9,
	0xa8, 0xd6, 0x84, 0xda, 0x1a, 0x63, 0xba, 0x65, 0xa2, 0xa3, 0xd6, 0x1f, 0x05, 0x2e, 0xb4, 0xb5,
	0x4c, 0xde, 0x1b, 0x78, 0x03, 0x38, 0x1b, 0xd0, 0x4c, 0xd5, 0x7e, 0xa4, 0x53, 0xc9, 0xe3, 0x3a,
	0x5c, 0x12, 0x8d, 0x39, 0x9d, 0xee, 0xad, 0x33, 0x52, 0x27, 0x8e, 0x69, 0x51, 0x32, 0x18, 0xe8,
	0xe6, 0x08, 0x29, 0xf1, 0x08, 0xed, 0x1b, 0x83, 0x61, 0xf7, 0x5e, 0x9f, 0x8e, 0xc7, 0x13, 0xa2,
	0xb2, 0x31, 0xe9, 0xa0, 0x5c, 0x9c, 0x6d, 0x33, 0xe2, 0x68, 0x6a, 0xdf, 0xb1, 0x88, 0x46, 0xb3,
	0x81, 0xe3, 0xcc, 0xd5, 0xed, 0x7d, 0x3a, 0x70, 0xe5, 0x5b, 0x33, 0xc0, 0xcc, 0x5b, 0xac, 0xdd,
	0xe8, 0x39, 0xe4, 0xff, 0x0f, 0x75, 0x09, 0x95, 0xec, 0x50, 0x4c, 0x1f, 0x99, 0x64, 0x3a, 0xb3,
	0x35, 0x74, 0x84, 0x2b, 0x50, 0x66, 0x63, 0xd2, 0xeb, 0xc8, 0x60, 0x9b, 0x11, 0xa4, 0xc4, 0xe4,
	0x5e, 0xc9, 0xa1, 0x8c, 0xa1, 0x5c, 0xeb, 0x0e, 0x5e, 0x1d, 0xdc, 0x1e, 0xbe, 0x86, 0x7a, 0xe6,
	0xa4, 0xb3, 0xfe, 0x37, 0x5d, 0x75, 0x0c, 0xed, 0xd1, 0x19, 0x5a, 0xf6, 0x84, 0x4c, 0xd1, 0x11,
	0x46, 0x50, 0xa2, 0x86, 0xca, 0x3a, 0x69, 0x13, 0x29, 0xb8, 0x00, 0x79, 0x6a, 0xe8, 0x0f, 0x28,
	0x17, 0xaf, 0x1e, 0x7a, 0x1f, 0x3e, 0xa3, 0xe3, 0x16, 0x05, 0x74, 0x78, 0x6b, 0xf8, 0x2d, 0x5c,
	0xed, 0xc4, 0xb6, 0x7e, 0x47, 0xa6, 0xda, 0xbe, 0xb9, 0x02, 0xe5, 0xd4, 0x9c, 0x74, 0x91, 0x82,
	0x8b, 0x70, 0x12, 0x97, 0x6e, 0x51, 0xae, 0x4f, 0xa0, 0x3a, 0x17, 0x7e, 0xdb, 0x5d, 0xf1, 0x68,
	0xc9, 0x3d, 0xf7, 0xc5, 0x0d, 0x79, 0xfa, 0x3e, 0xfa, 0x67, 0xaa, 0xfc, 0xd2, 0xf8, 0xd1, 0x7f,
	0xaf, 0xef, 0x13, 0xc2, 0xbf, 0x49, 0xa8, 0x85, 0x78, 0x3a, 0x95, 0xbf, 0xc5, 0xc7, 0x7f, 0x01,
	0x00, 0x00, 0xff, 0xff, 0x81, 0x97, 0x8a, 0x59, 0x26, 0x03, 0x00, 0x00,
}