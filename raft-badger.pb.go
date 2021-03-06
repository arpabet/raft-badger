// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.13.0
// source: raft-badger.proto

package raft_badger

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RaftLogType int32

const (
	RaftLogType_LogCommand              RaftLogType = 0
	RaftLogType_LogNoop                 RaftLogType = 1
	RaftLogType_LogAddPeerDeprecated    RaftLogType = 2
	RaftLogType_LogRemovePeerDeprecated RaftLogType = 3
	RaftLogType_LogBarrier              RaftLogType = 4
	RaftLogType_LogConfiguration        RaftLogType = 5
)

// Enum value maps for RaftLogType.
var (
	RaftLogType_name = map[int32]string{
		0: "LogCommand",
		1: "LogNoop",
		2: "LogAddPeerDeprecated",
		3: "LogRemovePeerDeprecated",
		4: "LogBarrier",
		5: "LogConfiguration",
	}
	RaftLogType_value = map[string]int32{
		"LogCommand":              0,
		"LogNoop":                 1,
		"LogAddPeerDeprecated":    2,
		"LogRemovePeerDeprecated": 3,
		"LogBarrier":              4,
		"LogConfiguration":        5,
	}
)

func (x RaftLogType) Enum() *RaftLogType {
	p := new(RaftLogType)
	*p = x
	return p
}

func (x RaftLogType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RaftLogType) Descriptor() protoreflect.EnumDescriptor {
	return file_raft_badger_proto_enumTypes[0].Descriptor()
}

func (RaftLogType) Type() protoreflect.EnumType {
	return &file_raft_badger_proto_enumTypes[0]
}

func (x RaftLogType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RaftLogType.Descriptor instead.
func (RaftLogType) EnumDescriptor() ([]byte, []int) {
	return file_raft_badger_proto_rawDescGZIP(), []int{0}
}

type RaftLog struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Index      uint64      `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	Term       uint64      `protobuf:"varint,2,opt,name=term,proto3" json:"term,omitempty"`
	Type       RaftLogType `protobuf:"varint,3,opt,name=type,proto3,enum=raftbadger.RaftLogType" json:"type,omitempty"`
	Data       []byte      `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
	Extensions []byte      `protobuf:"bytes,5,opt,name=extensions,proto3" json:"extensions,omitempty"`
}

func (x *RaftLog) Reset() {
	*x = RaftLog{}
	if protoimpl.UnsafeEnabled {
		mi := &file_raft_badger_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RaftLog) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RaftLog) ProtoMessage() {}

func (x *RaftLog) ProtoReflect() protoreflect.Message {
	mi := &file_raft_badger_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RaftLog.ProtoReflect.Descriptor instead.
func (*RaftLog) Descriptor() ([]byte, []int) {
	return file_raft_badger_proto_rawDescGZIP(), []int{0}
}

func (x *RaftLog) GetIndex() uint64 {
	if x != nil {
		return x.Index
	}
	return 0
}

func (x *RaftLog) GetTerm() uint64 {
	if x != nil {
		return x.Term
	}
	return 0
}

func (x *RaftLog) GetType() RaftLogType {
	if x != nil {
		return x.Type
	}
	return RaftLogType_LogCommand
}

func (x *RaftLog) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *RaftLog) GetExtensions() []byte {
	if x != nil {
		return x.Extensions
	}
	return nil
}

var File_raft_badger_proto protoreflect.FileDescriptor

var file_raft_badger_proto_rawDesc = []byte{
	0x0a, 0x11, 0x72, 0x61, 0x66, 0x74, 0x2d, 0x62, 0x61, 0x64, 0x67, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x72, 0x61, 0x66, 0x74, 0x62, 0x61, 0x64, 0x67, 0x65, 0x72, 0x22,
	0x94, 0x01, 0x0a, 0x07, 0x52, 0x61, 0x66, 0x74, 0x4c, 0x6f, 0x67, 0x12, 0x14, 0x0a, 0x05, 0x69,
	0x6e, 0x64, 0x65, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x69, 0x6e, 0x64, 0x65,
	0x78, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x72, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x04, 0x74, 0x65, 0x72, 0x6d, 0x12, 0x2b, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e, 0x72, 0x61, 0x66, 0x74, 0x62, 0x61, 0x64, 0x67, 0x65, 0x72,
	0x2e, 0x52, 0x61, 0x66, 0x74, 0x4c, 0x6f, 0x67, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x65, 0x78, 0x74, 0x65,
	0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2a, 0x87, 0x01, 0x0a, 0x0b, 0x52, 0x61, 0x66, 0x74, 0x4c,
	0x6f, 0x67, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0e, 0x0a, 0x0a, 0x4c, 0x6f, 0x67, 0x43, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x4c, 0x6f, 0x67, 0x4e, 0x6f, 0x6f,
	0x70, 0x10, 0x01, 0x12, 0x18, 0x0a, 0x14, 0x4c, 0x6f, 0x67, 0x41, 0x64, 0x64, 0x50, 0x65, 0x65,
	0x72, 0x44, 0x65, 0x70, 0x72, 0x65, 0x63, 0x61, 0x74, 0x65, 0x64, 0x10, 0x02, 0x12, 0x1b, 0x0a,
	0x17, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x50, 0x65, 0x65, 0x72, 0x44, 0x65,
	0x70, 0x72, 0x65, 0x63, 0x61, 0x74, 0x65, 0x64, 0x10, 0x03, 0x12, 0x0e, 0x0a, 0x0a, 0x4c, 0x6f,
	0x67, 0x42, 0x61, 0x72, 0x72, 0x69, 0x65, 0x72, 0x10, 0x04, 0x12, 0x14, 0x0a, 0x10, 0x4c, 0x6f,
	0x67, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x10, 0x05,
	0x42, 0x1c, 0x5a, 0x1a, 0x67, 0x6f, 0x2e, 0x61, 0x72, 0x70, 0x61, 0x62, 0x65, 0x74, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x72, 0x61, 0x66, 0x74, 0x2d, 0x62, 0x61, 0x64, 0x67, 0x65, 0x72, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_raft_badger_proto_rawDescOnce sync.Once
	file_raft_badger_proto_rawDescData = file_raft_badger_proto_rawDesc
)

func file_raft_badger_proto_rawDescGZIP() []byte {
	file_raft_badger_proto_rawDescOnce.Do(func() {
		file_raft_badger_proto_rawDescData = protoimpl.X.CompressGZIP(file_raft_badger_proto_rawDescData)
	})
	return file_raft_badger_proto_rawDescData
}

var file_raft_badger_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_raft_badger_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_raft_badger_proto_goTypes = []interface{}{
	(RaftLogType)(0), // 0: raftbadger.RaftLogType
	(*RaftLog)(nil),  // 1: raftbadger.RaftLog
}
var file_raft_badger_proto_depIdxs = []int32{
	0, // 0: raftbadger.RaftLog.type:type_name -> raftbadger.RaftLogType
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_raft_badger_proto_init() }
func file_raft_badger_proto_init() {
	if File_raft_badger_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_raft_badger_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RaftLog); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_raft_badger_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_raft_badger_proto_goTypes,
		DependencyIndexes: file_raft_badger_proto_depIdxs,
		EnumInfos:         file_raft_badger_proto_enumTypes,
		MessageInfos:      file_raft_badger_proto_msgTypes,
	}.Build()
	File_raft_badger_proto = out.File
	file_raft_badger_proto_rawDesc = nil
	file_raft_badger_proto_goTypes = nil
	file_raft_badger_proto_depIdxs = nil
}
