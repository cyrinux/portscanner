// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0-devel
// 	protoc        v3.15.2
// source: proto/service.proto

package proto

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
		mi := &file_proto_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Host) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Host) ProtoMessage() {}

func (x *Host) ProtoReflect() protoreflect.Message {
	mi := &file_proto_service_proto_msgTypes[0]
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
	return file_proto_service_proto_rawDescGZIP(), []int{0}
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
		mi := &file_proto_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortVersion) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortVersion) ProtoMessage() {}

func (x *PortVersion) ProtoReflect() protoreflect.Message {
	mi := &file_proto_service_proto_msgTypes[1]
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
	return file_proto_service_proto_rawDescGZIP(), []int{1}
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
		mi := &file_proto_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Port) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Port) ProtoMessage() {}

func (x *Port) ProtoReflect() protoreflect.Message {
	mi := &file_proto_service_proto_msgTypes[2]
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
	return file_proto_service_proto_rawDescGZIP(), []int{2}
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
		mi := &file_proto_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HostResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HostResult) ProtoMessage() {}

func (x *HostResult) ProtoReflect() protoreflect.Message {
	mi := &file_proto_service_proto_msgTypes[3]
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
	return file_proto_service_proto_rawDescGZIP(), []int{3}
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

type ScannerResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HostResult []*HostResult `protobuf:"bytes,1,rep,name=host_result,json=hostResult,proto3" json:"host_result,omitempty"`
}

func (x *ScannerResponse) Reset() {
	*x = ScannerResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ScannerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ScannerResponse) ProtoMessage() {}

func (x *ScannerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ScannerResponse.ProtoReflect.Descriptor instead.
func (*ScannerResponse) Descriptor() ([]byte, []int) {
	return file_proto_service_proto_rawDescGZIP(), []int{4}
}

func (x *ScannerResponse) GetHostResult() []*HostResult {
	if x != nil {
		return x.HostResult
	}
	return nil
}

type SetScannerRequest struct {
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

func (x *SetScannerRequest) Reset() {
	*x = SetScannerRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetScannerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetScannerRequest) ProtoMessage() {}

func (x *SetScannerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetScannerRequest.ProtoReflect.Descriptor instead.
func (*SetScannerRequest) Descriptor() ([]byte, []int) {
	return file_proto_service_proto_rawDescGZIP(), []int{5}
}

func (x *SetScannerRequest) GetHosts() string {
	if x != nil {
		return x.Hosts
	}
	return ""
}

func (x *SetScannerRequest) GetPorts() string {
	if x != nil {
		return x.Ports
	}
	return ""
}

func (x *SetScannerRequest) GetWithAggressiveScan() bool {
	if x != nil {
		return x.WithAggressiveScan
	}
	return false
}

func (x *SetScannerRequest) GetWithSynScan() bool {
	if x != nil {
		return x.WithSynScan
	}
	return false
}

func (x *SetScannerRequest) GetWithNullScan() bool {
	if x != nil {
		return x.WithNullScan
	}
	return false
}

func (x *SetScannerRequest) GetServiceVersionDetection() bool {
	if x != nil {
		return x.ServiceVersionDetection
	}
	return false
}

func (x *SetScannerRequest) GetServiceOsDetection() bool {
	if x != nil {
		return x.ServiceOsDetection
	}
	return false
}

func (x *SetScannerRequest) GetServiceDefaultScripts() bool {
	if x != nil {
		return x.ServiceDefaultScripts
	}
	return false
}

func (x *SetScannerRequest) GetTimeout() int32 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

type GetScannerResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *GetScannerResponse) Reset() {
	*x = GetScannerResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_service_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetScannerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetScannerResponse) ProtoMessage() {}

func (x *GetScannerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_service_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetScannerResponse.ProtoReflect.Descriptor instead.
func (*GetScannerResponse) Descriptor() ([]byte, []int) {
	return file_proto_service_proto_rawDescGZIP(), []int{6}
}

func (x *GetScannerResponse) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

type ServerResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Value   string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Error   string `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *ServerResponse) Reset() {
	*x = ServerResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_service_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerResponse) ProtoMessage() {}

func (x *ServerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_service_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerResponse.ProtoReflect.Descriptor instead.
func (*ServerResponse) Descriptor() ([]byte, []int) {
	return file_proto_service_proto_rawDescGZIP(), []int{7}
}

func (x *ServerResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *ServerResponse) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *ServerResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_proto_service_proto protoreflect.FileDescriptor

var file_proto_service_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x53, 0x0a, 0x04,
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
	0x74, 0x22, 0xa2, 0x01, 0x0a, 0x04, 0x50, 0x6f, 0x72, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x6f,
	0x72, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x6f, 0x72,
	0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12,
	0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2c, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x76,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x50, 0x0a, 0x0a, 0x48, 0x6f, 0x73, 0x74, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x12, 0x1f, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x48, 0x6f, 0x73, 0x74, 0x52,
	0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x50, 0x6f, 0x72,
	0x74, 0x52, 0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x22, 0x45, 0x0a, 0x0f, 0x53, 0x63, 0x61, 0x6e,
	0x6e, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x32, 0x0a, 0x0b, 0x68,
	0x6f, 0x73, 0x74, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x11, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x48, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x52, 0x0a, 0x68, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22,
	0xfb, 0x02, 0x0a, 0x11, 0x53, 0x65, 0x74, 0x53, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x68, 0x6f, 0x73, 0x74, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x68, 0x6f, 0x73, 0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x70,
	0x6f, 0x72, 0x74, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x70, 0x6f, 0x72, 0x74,
	0x73, 0x12, 0x30, 0x0a, 0x14, 0x77, 0x69, 0x74, 0x68, 0x5f, 0x61, 0x67, 0x67, 0x72, 0x65, 0x73,
	0x73, 0x69, 0x76, 0x65, 0x5f, 0x73, 0x63, 0x61, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x12, 0x77, 0x69, 0x74, 0x68, 0x41, 0x67, 0x67, 0x72, 0x65, 0x73, 0x73, 0x69, 0x76, 0x65, 0x53,
	0x63, 0x61, 0x6e, 0x12, 0x22, 0x0a, 0x0d, 0x77, 0x69, 0x74, 0x68, 0x5f, 0x73, 0x79, 0x6e, 0x5f,
	0x73, 0x63, 0x61, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x77, 0x69, 0x74, 0x68,
	0x53, 0x79, 0x6e, 0x53, 0x63, 0x61, 0x6e, 0x12, 0x24, 0x0a, 0x0e, 0x77, 0x69, 0x74, 0x68, 0x5f,
	0x6e, 0x75, 0x6c, 0x6c, 0x5f, 0x73, 0x63, 0x61, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x0c, 0x77, 0x69, 0x74, 0x68, 0x4e, 0x75, 0x6c, 0x6c, 0x53, 0x63, 0x61, 0x6e, 0x12, 0x3a, 0x0a,
	0x19, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x5f, 0x64, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x17, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x30, 0x0a, 0x14, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x5f, 0x6f, 0x73, 0x5f, 0x64, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x12, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x4f, 0x73, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x36, 0x0a, 0x17, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x5f, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x52, 0x15, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x53, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x09,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x22, 0x26, 0x0a,
	0x12, 0x47, 0x65, 0x74, 0x53, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0x56, 0x0a, 0x0e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x32, 0x8b, 0x01,
	0x0a, 0x0e, 0x53, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x3c, 0x0a, 0x09, 0x53, 0x74, 0x61, 0x72, 0x74, 0x53, 0x63, 0x61, 0x6e, 0x12, 0x18, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x74, 0x53, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3b,
	0x0a, 0x07, 0x47, 0x65, 0x74, 0x53, 0x63, 0x61, 0x6e, 0x12, 0x19, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x1a, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x08, 0x5a, 0x06, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_service_proto_rawDescOnce sync.Once
	file_proto_service_proto_rawDescData = file_proto_service_proto_rawDesc
)

func file_proto_service_proto_rawDescGZIP() []byte {
	file_proto_service_proto_rawDescOnce.Do(func() {
		file_proto_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_service_proto_rawDescData)
	})
	return file_proto_service_proto_rawDescData
}

var file_proto_service_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_proto_service_proto_goTypes = []interface{}{
	(*Host)(nil),               // 0: proto.Host
	(*PortVersion)(nil),        // 1: proto.PortVersion
	(*Port)(nil),               // 2: proto.Port
	(*HostResult)(nil),         // 3: proto.HostResult
	(*ScannerResponse)(nil),    // 4: proto.ScannerResponse
	(*SetScannerRequest)(nil),  // 5: proto.SetScannerRequest
	(*GetScannerResponse)(nil), // 6: proto.GetScannerResponse
	(*ServerResponse)(nil),     // 7: proto.ServerResponse
}
var file_proto_service_proto_depIdxs = []int32{
	1, // 0: proto.Port.version:type_name -> proto.PortVersion
	0, // 1: proto.HostResult.host:type_name -> proto.Host
	2, // 2: proto.HostResult.ports:type_name -> proto.Port
	3, // 3: proto.ScannerResponse.host_result:type_name -> proto.HostResult
	5, // 4: proto.ScannerService.StartScan:input_type -> proto.SetScannerRequest
	6, // 5: proto.ScannerService.GetScan:input_type -> proto.GetScannerResponse
	7, // 6: proto.ScannerService.StartScan:output_type -> proto.ServerResponse
	7, // 7: proto.ScannerService.GetScan:output_type -> proto.ServerResponse
	6, // [6:8] is the sub-list for method output_type
	4, // [4:6] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proto_service_proto_init() }
func file_proto_service_proto_init() {
	if File_proto_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_proto_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
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
		file_proto_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_proto_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
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
		file_proto_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ScannerResponse); i {
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
		file_proto_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetScannerRequest); i {
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
		file_proto_service_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetScannerResponse); i {
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
		file_proto_service_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerResponse); i {
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
	file_proto_service_proto_msgTypes[0].OneofWrappers = []interface{}{}
	file_proto_service_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_service_proto_goTypes,
		DependencyIndexes: file_proto_service_proto_depIdxs,
		MessageInfos:      file_proto_service_proto_msgTypes,
	}.Build()
	File_proto_service_proto = out.File
	file_proto_service_proto_rawDesc = nil
	file_proto_service_proto_goTypes = nil
	file_proto_service_proto_depIdxs = nil
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
	StartScan(ctx context.Context, in *SetScannerRequest, opts ...grpc.CallOption) (*ServerResponse, error)
	// Fetch a previous scan result
	// We are trying to fix this implem
	GetScan(ctx context.Context, in *GetScannerResponse, opts ...grpc.CallOption) (*ServerResponse, error)
}

type scannerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewScannerServiceClient(cc grpc.ClientConnInterface) ScannerServiceClient {
	return &scannerServiceClient{cc}
}

func (c *scannerServiceClient) StartScan(ctx context.Context, in *SetScannerRequest, opts ...grpc.CallOption) (*ServerResponse, error) {
	out := new(ServerResponse)
	err := c.cc.Invoke(ctx, "/proto.ScannerService/StartScan", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *scannerServiceClient) GetScan(ctx context.Context, in *GetScannerResponse, opts ...grpc.CallOption) (*ServerResponse, error) {
	out := new(ServerResponse)
	err := c.cc.Invoke(ctx, "/proto.ScannerService/GetScan", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ScannerServiceServer is the server API for ScannerService service.
type ScannerServiceServer interface {
	// Create new scanner
	StartScan(context.Context, *SetScannerRequest) (*ServerResponse, error)
	// Fetch a previous scan result
	// We are trying to fix this implem
	GetScan(context.Context, *GetScannerResponse) (*ServerResponse, error)
}

// UnimplementedScannerServiceServer can be embedded to have forward compatible implementations.
type UnimplementedScannerServiceServer struct {
}

func (*UnimplementedScannerServiceServer) StartScan(context.Context, *SetScannerRequest) (*ServerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartScan not implemented")
}
func (*UnimplementedScannerServiceServer) GetScan(context.Context, *GetScannerResponse) (*ServerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetScan not implemented")
}

func RegisterScannerServiceServer(s *grpc.Server, srv ScannerServiceServer) {
	s.RegisterService(&_ScannerService_serviceDesc, srv)
}

func _ScannerService_StartScan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetScannerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScannerServiceServer).StartScan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ScannerService/StartScan",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScannerServiceServer).StartScan(ctx, req.(*SetScannerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ScannerService_GetScan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetScannerResponse)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScannerServiceServer).GetScan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ScannerService/GetScan",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScannerServiceServer).GetScan(ctx, req.(*GetScannerResponse))
	}
	return interceptor(ctx, in, info, handler)
}

var _ScannerService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ScannerService",
	HandlerType: (*ScannerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "StartScan",
			Handler:    _ScannerService_StartScan_Handler,
		},
		{
			MethodName: "GetScan",
			Handler:    _ScannerService_GetScan_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/service.proto",
}