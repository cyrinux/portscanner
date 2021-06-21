// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0-devel
// 	protoc        v3.15.2
// source: scanner/scanner.proto

package scanner

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type Scanner struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hosts                   string `protobuf:"bytes,1,opt,name=hosts,proto3" json:"hosts,omitempty"`
	Ports                   string `protobuf:"bytes,2,opt,name=ports,proto3" json:"ports,omitempty"`
	WithAggressiveScan      bool   `protobuf:"varint,3,opt,name=with_aggressive_scan,json=withAggressiveScan,proto3" json:"with_aggressive_scan,omitempty"`
	WithSynScan             bool   `protobuf:"varint,4,opt,name=with_syn_scan,json=withSynScan,proto3" json:"with_syn_scan,omitempty"`
	WithNullScan            bool   `protobuf:"varint,5,opt,name=with_null_scan,json=withNullScan,proto3" json:"with_null_scan,omitempty"`
	ServiceVersionDetection bool   `protobuf:"varint,6,opt,name=service_version_detection,json=serviceVersionDetection,proto3" json:"service_version_detection,omitempty"`
	ServiceOsDetection      bool   `protobuf:"varint,7,opt,name=service_os_detection,json=serviceOsDetection,proto3" json:"service_os_detection,omitempty"`
	ServiceDefaultScripts   bool   `protobuf:"varint,8,opt,name=service_default_scripts,json=serviceDefaultScripts,proto3" json:"service_default_scripts,omitempty"`
	Timeout                 int32  `protobuf:"varint,9,opt,name=timeout,proto3" json:"timeout,omitempty"`
}

func (x *Scanner) Reset() {
	*x = Scanner{}
	if protoimpl.UnsafeEnabled {
		mi := &file_scanner_scanner_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Scanner) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Scanner) ProtoMessage() {}

func (x *Scanner) ProtoReflect() protoreflect.Message {
	mi := &file_scanner_scanner_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Scanner.ProtoReflect.Descriptor instead.
func (*Scanner) Descriptor() ([]byte, []int) {
	return file_scanner_scanner_proto_rawDescGZIP(), []int{0}
}

func (x *Scanner) GetHosts() string {
	if x != nil {
		return x.Hosts
	}
	return ""
}

func (x *Scanner) GetPorts() string {
	if x != nil {
		return x.Ports
	}
	return ""
}

func (x *Scanner) GetWithAggressiveScan() bool {
	if x != nil {
		return x.WithAggressiveScan
	}
	return false
}

func (x *Scanner) GetWithSynScan() bool {
	if x != nil {
		return x.WithSynScan
	}
	return false
}

func (x *Scanner) GetWithNullScan() bool {
	if x != nil {
		return x.WithNullScan
	}
	return false
}

func (x *Scanner) GetServiceVersionDetection() bool {
	if x != nil {
		return x.ServiceVersionDetection
	}
	return false
}

func (x *Scanner) GetServiceOsDetection() bool {
	if x != nil {
		return x.ServiceOsDetection
	}
	return false
}

func (x *Scanner) GetServiceDefaultScripts() bool {
	if x != nil {
		return x.ServiceDefaultScripts
	}
	return false
}

func (x *Scanner) GetTimeout() int32 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

type Task struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Task) Reset() {
	*x = Task{}
	if protoimpl.UnsafeEnabled {
		mi := &file_scanner_scanner_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Task) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Task) ProtoMessage() {}

func (x *Task) ProtoReflect() protoreflect.Message {
	mi := &file_scanner_scanner_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Task.ProtoReflect.Descriptor instead.
func (*Task) Descriptor() ([]byte, []int) {
	return file_scanner_scanner_proto_rawDescGZIP(), []int{1}
}

func (x *Task) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type Host struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address   string  `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	OsVersion *string `protobuf:"bytes,2,opt,name=os_version,json=osVersion,proto3,oneof" json:"os_version,omitempty"`
}

func (x *Host) Reset() {
	*x = Host{}
	if protoimpl.UnsafeEnabled {
		mi := &file_scanner_scanner_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Host) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Host) ProtoMessage() {}

func (x *Host) ProtoReflect() protoreflect.Message {
	mi := &file_scanner_scanner_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Host.ProtoReflect.Descriptor instead.
func (*Host) Descriptor() ([]byte, []int) {
	return file_scanner_scanner_proto_rawDescGZIP(), []int{2}
}

func (x *Host) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *Host) GetOsVersion() string {
	if x != nil && x.OsVersion != nil {
		return *x.OsVersion
	}
	return ""
}

type PortVersion struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LowVersion  *string `protobuf:"bytes,1,opt,name=low_version,json=lowVersion,proto3,oneof" json:"low_version,omitempty"`
	HighVersion *string `protobuf:"bytes,2,opt,name=high_version,json=highVersion,proto3,oneof" json:"high_version,omitempty"`
	ExtraInfos  *string `protobuf:"bytes,3,opt,name=extra_infos,json=extraInfos,proto3,oneof" json:"extra_infos,omitempty"`
	Product     *string `protobuf:"bytes,4,opt,name=product,proto3,oneof" json:"product,omitempty"`
}

func (x *PortVersion) Reset() {
	*x = PortVersion{}
	if protoimpl.UnsafeEnabled {
		mi := &file_scanner_scanner_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortVersion) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortVersion) ProtoMessage() {}

func (x *PortVersion) ProtoReflect() protoreflect.Message {
	mi := &file_scanner_scanner_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortVersion.ProtoReflect.Descriptor instead.
func (*PortVersion) Descriptor() ([]byte, []int) {
	return file_scanner_scanner_proto_rawDescGZIP(), []int{3}
}

func (x *PortVersion) GetLowVersion() string {
	if x != nil && x.LowVersion != nil {
		return *x.LowVersion
	}
	return ""
}

func (x *PortVersion) GetHighVersion() string {
	if x != nil && x.HighVersion != nil {
		return *x.HighVersion
	}
	return ""
}

func (x *PortVersion) GetExtraInfos() string {
	if x != nil && x.ExtraInfos != nil {
		return *x.ExtraInfos
	}
	return ""
}

func (x *PortVersion) GetProduct() string {
	if x != nil && x.Product != nil {
		return *x.Product
	}
	return ""
}

type Port struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PortId      string       `protobuf:"bytes,1,opt,name=port_id,json=portId,proto3" json:"port_id,omitempty"`
	Protocol    string       `protobuf:"bytes,2,opt,name=protocol,proto3" json:"protocol,omitempty"`
	State       string       `protobuf:"bytes,3,opt,name=state,proto3" json:"state,omitempty"`
	ServiceName string       `protobuf:"bytes,4,opt,name=service_name,json=serviceName,proto3" json:"service_name,omitempty"`
	Version     *PortVersion `protobuf:"bytes,5,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *Port) Reset() {
	*x = Port{}
	if protoimpl.UnsafeEnabled {
		mi := &file_scanner_scanner_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Port) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Port) ProtoMessage() {}

func (x *Port) ProtoReflect() protoreflect.Message {
	mi := &file_scanner_scanner_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Port.ProtoReflect.Descriptor instead.
func (*Port) Descriptor() ([]byte, []int) {
	return file_scanner_scanner_proto_rawDescGZIP(), []int{4}
}

func (x *Port) GetPortId() string {
	if x != nil {
		return x.PortId
	}
	return ""
}

func (x *Port) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *Port) GetState() string {
	if x != nil {
		return x.State
	}
	return ""
}

func (x *Port) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *Port) GetVersion() *PortVersion {
	if x != nil {
		return x.Version
	}
	return nil
}

type HostResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host  *Host   `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Ports []*Port `protobuf:"bytes,2,rep,name=ports,proto3" json:"ports,omitempty"`
}

func (x *HostResult) Reset() {
	*x = HostResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_scanner_scanner_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HostResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HostResult) ProtoMessage() {}

func (x *HostResult) ProtoReflect() protoreflect.Message {
	mi := &file_scanner_scanner_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HostResult.ProtoReflect.Descriptor instead.
func (*HostResult) Descriptor() ([]byte, []int) {
	return file_scanner_scanner_proto_rawDescGZIP(), []int{5}
}

func (x *HostResult) GetHost() *Host {
	if x != nil {
		return x.Host
	}
	return nil
}

func (x *HostResult) GetPorts() []*Port {
	if x != nil {
		return x.Ports
	}
	return nil
}

// ScanResult is what is return when i Nmap scan, and- maybe - what
// i put in the database
// and i would like to return the same with GetScan
type ScanResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id         []byte        `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	HostResult []*HostResult `protobuf:"bytes,2,rep,name=host_result,json=hostResult,proto3" json:"host_result,omitempty"`
}

func (x *ScanResult) Reset() {
	*x = ScanResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_scanner_scanner_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ScanResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ScanResult) ProtoMessage() {}

func (x *ScanResult) ProtoReflect() protoreflect.Message {
	mi := &file_scanner_scanner_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ScanResult.ProtoReflect.Descriptor instead.
func (*ScanResult) Descriptor() ([]byte, []int) {
	return file_scanner_scanner_proto_rawDescGZIP(), []int{6}
}

func (x *ScanResult) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *ScanResult) GetHostResult() []*HostResult {
	if x != nil {
		return x.HostResult
	}
	return nil
}

var File_scanner_scanner_proto protoreflect.FileDescriptor

var file_scanner_scanner_proto_rawDesc = []byte{
	0x0a, 0x15, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x2f, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72,
	0x22, 0xf1, 0x02, 0x0a, 0x07, 0x53, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05,
	0x68, 0x6f, 0x73, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x68, 0x6f, 0x73,
	0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x30, 0x0a, 0x14, 0x77, 0x69, 0x74, 0x68,
	0x5f, 0x61, 0x67, 0x67, 0x72, 0x65, 0x73, 0x73, 0x69, 0x76, 0x65, 0x5f, 0x73, 0x63, 0x61, 0x6e,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x12, 0x77, 0x69, 0x74, 0x68, 0x41, 0x67, 0x67, 0x72,
	0x65, 0x73, 0x73, 0x69, 0x76, 0x65, 0x53, 0x63, 0x61, 0x6e, 0x12, 0x22, 0x0a, 0x0d, 0x77, 0x69,
	0x74, 0x68, 0x5f, 0x73, 0x79, 0x6e, 0x5f, 0x73, 0x63, 0x61, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x0b, 0x77, 0x69, 0x74, 0x68, 0x53, 0x79, 0x6e, 0x53, 0x63, 0x61, 0x6e, 0x12, 0x24,
	0x0a, 0x0e, 0x77, 0x69, 0x74, 0x68, 0x5f, 0x6e, 0x75, 0x6c, 0x6c, 0x5f, 0x73, 0x63, 0x61, 0x6e,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x77, 0x69, 0x74, 0x68, 0x4e, 0x75, 0x6c, 0x6c,
	0x53, 0x63, 0x61, 0x6e, 0x12, 0x3a, 0x0a, 0x19, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x64, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x17, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x30, 0x0a, 0x14, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6f, 0x73, 0x5f, 0x64,
	0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x12,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x73, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x36, 0x0a, 0x17, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x64, 0x65,
	0x66, 0x61, 0x75, 0x6c, 0x74, 0x5f, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x73, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x15, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x44, 0x65, 0x66, 0x61,
	0x75, 0x6c, 0x74, 0x53, 0x63, 0x72, 0x69, 0x70, 0x74, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x74, 0x69, 0x6d,
	0x65, 0x6f, 0x75, 0x74, 0x22, 0x16, 0x0a, 0x04, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x53, 0x0a, 0x04,
	0x48, 0x6f, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x22,
	0x0a, 0x0a, 0x6f, 0x73, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x00, 0x52, 0x09, 0x6f, 0x73, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x88,
	0x01, 0x01, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x6f, 0x73, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x22, 0xdd, 0x01, 0x0a, 0x0b, 0x50, 0x6f, 0x72, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x24, 0x0a, 0x0b, 0x6c, 0x6f, 0x77, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0a, 0x6c, 0x6f, 0x77, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x12, 0x26, 0x0a, 0x0c, 0x68, 0x69, 0x67, 0x68, 0x5f,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52,
	0x0b, 0x68, 0x69, 0x67, 0x68, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x12,
	0x24, 0x0a, 0x0b, 0x65, 0x78, 0x74, 0x72, 0x61, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x73, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x02, 0x52, 0x0a, 0x65, 0x78, 0x74, 0x72, 0x61, 0x49, 0x6e, 0x66,
	0x6f, 0x73, 0x88, 0x01, 0x01, 0x12, 0x1d, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x03, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63,
	0x74, 0x88, 0x01, 0x01, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x6c, 0x6f, 0x77, 0x5f, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x42, 0x0f, 0x0a, 0x0d, 0x5f, 0x68, 0x69, 0x67, 0x68, 0x5f, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x65, 0x78, 0x74, 0x72, 0x61, 0x5f,
	0x69, 0x6e, 0x66, 0x6f, 0x73, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63,
	0x74, 0x22, 0xa4, 0x01, 0x0a, 0x04, 0x50, 0x6f, 0x72, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x6f,
	0x72, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x6f, 0x72,
	0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12,
	0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2e, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x73, 0x63, 0x61, 0x6e,
	0x6e, 0x65, 0x72, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52,
	0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x54, 0x0a, 0x0a, 0x48, 0x6f, 0x73, 0x74,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x21, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x48,
	0x6f, 0x73, 0x74, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x05, 0x70, 0x6f, 0x72,
	0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x73, 0x63, 0x61, 0x6e, 0x6e,
	0x65, 0x72, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x52, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x22, 0x52,
	0x0a, 0x0a, 0x53, 0x63, 0x61, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x34, 0x0a, 0x0b,
	0x68, 0x6f, 0x73, 0x74, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x13, 0x2e, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x48, 0x6f, 0x73, 0x74,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x0a, 0x68, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x32, 0x6e, 0x0a, 0x0e, 0x53, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x2d, 0x0a, 0x04, 0x53, 0x63, 0x61, 0x6e, 0x12, 0x10, 0x2e, 0x73,
	0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x53, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x1a, 0x13,
	0x2e, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x53, 0x63, 0x61, 0x6e, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x12, 0x2d, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x53, 0x63, 0x61, 0x6e, 0x12, 0x0d,
	0x2e, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x1a, 0x13, 0x2e,
	0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x53, 0x63, 0x61, 0x6e, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x42, 0x0a, 0x5a, 0x08, 0x2f, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_scanner_scanner_proto_rawDescOnce sync.Once
	file_scanner_scanner_proto_rawDescData = file_scanner_scanner_proto_rawDesc
)

func file_scanner_scanner_proto_rawDescGZIP() []byte {
	file_scanner_scanner_proto_rawDescOnce.Do(func() {
		file_scanner_scanner_proto_rawDescData = protoimpl.X.CompressGZIP(file_scanner_scanner_proto_rawDescData)
	})
	return file_scanner_scanner_proto_rawDescData
}

var file_scanner_scanner_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_scanner_scanner_proto_goTypes = []interface{}{
	(*Scanner)(nil),     // 0: scanner.Scanner
	(*Task)(nil),        // 1: scanner.Task
	(*Host)(nil),        // 2: scanner.Host
	(*PortVersion)(nil), // 3: scanner.PortVersion
	(*Port)(nil),        // 4: scanner.Port
	(*HostResult)(nil),  // 5: scanner.HostResult
	(*ScanResult)(nil),  // 6: scanner.ScanResult
}
var file_scanner_scanner_proto_depIdxs = []int32{
	3, // 0: scanner.Port.version:type_name -> scanner.PortVersion
	2, // 1: scanner.HostResult.host:type_name -> scanner.Host
	4, // 2: scanner.HostResult.ports:type_name -> scanner.Port
	5, // 3: scanner.ScanResult.host_result:type_name -> scanner.HostResult
	0, // 4: scanner.ScannerService.Scan:input_type -> scanner.Scanner
	1, // 5: scanner.ScannerService.GetScan:input_type -> scanner.Task
	6, // 6: scanner.ScannerService.Scan:output_type -> scanner.ScanResult
	6, // 7: scanner.ScannerService.GetScan:output_type -> scanner.ScanResult
	6, // [6:8] is the sub-list for method output_type
	4, // [4:6] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_scanner_scanner_proto_init() }
func file_scanner_scanner_proto_init() {
	if File_scanner_scanner_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_scanner_scanner_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Scanner); i {
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
		file_scanner_scanner_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Task); i {
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
		file_scanner_scanner_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Host); i {
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
		file_scanner_scanner_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PortVersion); i {
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
		file_scanner_scanner_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Port); i {
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
		file_scanner_scanner_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HostResult); i {
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
		file_scanner_scanner_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ScanResult); i {
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
	file_scanner_scanner_proto_msgTypes[2].OneofWrappers = []interface{}{}
	file_scanner_scanner_proto_msgTypes[3].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_scanner_scanner_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_scanner_scanner_proto_goTypes,
		DependencyIndexes: file_scanner_scanner_proto_depIdxs,
		MessageInfos:      file_scanner_scanner_proto_msgTypes,
	}.Build()
	File_scanner_scanner_proto = out.File
	file_scanner_scanner_proto_rawDesc = nil
	file_scanner_scanner_proto_goTypes = nil
	file_scanner_scanner_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ScannerServiceClient is the client API for ScannerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ScannerServiceClient interface {
	// Create new scanner
	Scan(ctx context.Context, in *Scanner, opts ...grpc.CallOption) (*ScanResult, error)
	// Fetch a previous scan result
	// We are trying to fix this implem
	GetScan(ctx context.Context, in *Task, opts ...grpc.CallOption) (*ScanResult, error)
}

type scannerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewScannerServiceClient(cc grpc.ClientConnInterface) ScannerServiceClient {
	return &scannerServiceClient{cc}
}

func (c *scannerServiceClient) Scan(ctx context.Context, in *Scanner, opts ...grpc.CallOption) (*ScanResult, error) {
	out := new(ScanResult)
	err := c.cc.Invoke(ctx, "/scanner.ScannerService/Scan", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *scannerServiceClient) GetScan(ctx context.Context, in *Task, opts ...grpc.CallOption) (*ScanResult, error) {
	out := new(ScanResult)
	err := c.cc.Invoke(ctx, "/scanner.ScannerService/GetScan", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ScannerServiceServer is the server API for ScannerService service.
type ScannerServiceServer interface {
	// Create new scanner
	Scan(context.Context, *Scanner) (*ScanResult, error)
	// Fetch a previous scan result
	// We are trying to fix this implem
	GetScan(context.Context, *Task) (*ScanResult, error)
}

// UnimplementedScannerServiceServer can be embedded to have forward compatible implementations.
type UnimplementedScannerServiceServer struct {
}

func (*UnimplementedScannerServiceServer) Scan(context.Context, *Scanner) (*ScanResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Scan not implemented")
}
func (*UnimplementedScannerServiceServer) GetScan(context.Context, *Task) (*ScanResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetScan not implemented")
}

func RegisterScannerServiceServer(s *grpc.Server, srv ScannerServiceServer) {
	s.RegisterService(&_ScannerService_serviceDesc, srv)
}

func _ScannerService_Scan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Scanner)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScannerServiceServer).Scan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/scanner.ScannerService/Scan",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScannerServiceServer).Scan(ctx, req.(*Scanner))
	}
	return interceptor(ctx, in, info, handler)
}

func _ScannerService_GetScan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Task)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScannerServiceServer).GetScan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/scanner.ScannerService/GetScan",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScannerServiceServer).GetScan(ctx, req.(*Task))
	}
	return interceptor(ctx, in, info, handler)
}

var _ScannerService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "scanner.ScannerService",
	HandlerType: (*ScannerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Scan",
			Handler:    _ScannerService_Scan_Handler,
		},
		{
			MethodName: "GetScan",
			Handler:    _ScannerService_GetScan_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "scanner/scanner.proto",
}
