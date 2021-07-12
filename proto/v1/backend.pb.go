// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: proto/v1/backend.proto

package v1

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

// number of tasks messages by type
type TasksStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success  int64 `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Failed   int64 `protobuf:"varint,2,opt,name=failed,proto3" json:"failed,omitempty"`
	Returned int64 `protobuf:"varint,3,opt,name=returned,proto3" json:"returned,omitempty"`
	Ready    int64 `protobuf:"varint,4,opt,name=ready,proto3" json:"ready,omitempty"`
	Rejected int64 `protobuf:"varint,5,opt,name=rejected,proto3" json:"rejected,omitempty"`
}

func (x *TasksStatus) Reset() {
	*x = TasksStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_v1_backend_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TasksStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TasksStatus) ProtoMessage() {}

func (x *TasksStatus) ProtoReflect() protoreflect.Message {
	mi := &file_proto_v1_backend_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TasksStatus.ProtoReflect.Descriptor instead.
func (*TasksStatus) Descriptor() ([]byte, []int) {
	return file_proto_v1_backend_proto_rawDescGZIP(), []int{0}
}

func (x *TasksStatus) GetSuccess() int64 {
	if x != nil {
		return x.Success
	}
	return 0
}

func (x *TasksStatus) GetFailed() int64 {
	if x != nil {
		return x.Failed
	}
	return 0
}

func (x *TasksStatus) GetReturned() int64 {
	if x != nil {
		return x.Returned
	}
	return 0
}

func (x *TasksStatus) GetReady() int64 {
	if x != nil {
		return x.Ready
	}
	return 0
}

func (x *TasksStatus) GetRejected() int64 {
	if x != nil {
		return x.Rejected
	}
	return 0
}

// prometheus status
type PrometheusStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TasksStatus   *TasksStatus `protobuf:"bytes,1,opt,name=tasks_status,json=tasksStatus,proto3" json:"tasks_status,omitempty"`
	WorkerCount   int64        `protobuf:"varint,2,opt,name=worker_count,json=workerCount,proto3" json:"worker_count,omitempty"`
	ConsumerCount int64        `protobuf:"varint,3,opt,name=consumer_count,json=consumerCount,proto3" json:"consumer_count,omitempty"`
}

func (x *PrometheusStatus) Reset() {
	*x = PrometheusStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_v1_backend_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PrometheusStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrometheusStatus) ProtoMessage() {}

func (x *PrometheusStatus) ProtoReflect() protoreflect.Message {
	mi := &file_proto_v1_backend_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrometheusStatus.ProtoReflect.Descriptor instead.
func (*PrometheusStatus) Descriptor() ([]byte, []int) {
	return file_proto_v1_backend_proto_rawDescGZIP(), []int{1}
}

func (x *PrometheusStatus) GetTasksStatus() *TasksStatus {
	if x != nil {
		return x.TasksStatus
	}
	return nil
}

func (x *PrometheusStatus) GetWorkerCount() int64 {
	if x != nil {
		return x.WorkerCount
	}
	return 0
}

func (x *PrometheusStatus) GetConsumerCount() int64 {
	if x != nil {
		return x.ConsumerCount
	}
	return 0
}

var File_proto_v1_backend_proto protoreflect.FileDescriptor

var file_proto_v1_backend_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x76, 0x31, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65,
	0x6e, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8d, 0x01, 0x0a, 0x0b, 0x54, 0x61, 0x73, 0x6b,
	0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x06, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x74,
	0x75, 0x72, 0x6e, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x72, 0x65, 0x74,
	0x75, 0x72, 0x6e, 0x65, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x65, 0x61, 0x64, 0x79, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x72, 0x65, 0x61, 0x64, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x72,
	0x65, 0x6a, 0x65, 0x63, 0x74, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x72,
	0x65, 0x6a, 0x65, 0x63, 0x74, 0x65, 0x64, 0x22, 0x93, 0x01, 0x0a, 0x10, 0x50, 0x72, 0x6f, 0x6d,
	0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x35, 0x0a, 0x0c,
	0x74, 0x61, 0x73, 0x6b, 0x73, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x73,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x0b, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x5f, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x77, 0x6f, 0x72, 0x6b, 0x65,
	0x72, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x25, 0x0a, 0x0e, 0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6d,
	0x65, 0x72, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d,
	0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65, 0x72, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x32, 0xab, 0x01,
	0x0a, 0x0e, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x4e, 0x0a, 0x14, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x12, 0x19, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x73, 0x1a, 0x19, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x30, 0x01,
	0x12, 0x49, 0x0a, 0x11, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x54, 0x61, 0x73, 0x6b, 0x73, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x50, 0x72,
	0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x1a, 0x17,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x50, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75,
	0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x28, 0x01, 0x30, 0x01, 0x42, 0x0b, 0x5a, 0x09, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_v1_backend_proto_rawDescOnce sync.Once
	file_proto_v1_backend_proto_rawDescData = file_proto_v1_backend_proto_rawDesc
)

func file_proto_v1_backend_proto_rawDescGZIP() []byte {
	file_proto_v1_backend_proto_rawDescOnce.Do(func() {
		file_proto_v1_backend_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_v1_backend_proto_rawDescData)
	})
	return file_proto_v1_backend_proto_rawDescData
}

var file_proto_v1_backend_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_v1_backend_proto_goTypes = []interface{}{
	(*TasksStatus)(nil),        // 0: proto.TasksStatus
	(*PrometheusStatus)(nil),   // 1: proto.PrometheusStatus
	(*ServiceStateValues)(nil), // 2: proto.ServiceStateValues
}
var file_proto_v1_backend_proto_depIdxs = []int32{
	0, // 0: proto.PrometheusStatus.tasks_status:type_name -> proto.TasksStatus
	2, // 1: proto.BackendService.StreamServiceControl:input_type -> proto.ServiceStateValues
	1, // 2: proto.BackendService.StreamTasksStatus:input_type -> proto.PrometheusStatus
	2, // 3: proto.BackendService.StreamServiceControl:output_type -> proto.ServiceStateValues
	1, // 4: proto.BackendService.StreamTasksStatus:output_type -> proto.PrometheusStatus
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proto_v1_backend_proto_init() }
func file_proto_v1_backend_proto_init() {
	if File_proto_v1_backend_proto != nil {
		return
	}
	file_proto_v1_service_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_proto_v1_backend_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TasksStatus); i {
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
		file_proto_v1_backend_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PrometheusStatus); i {
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
			RawDescriptor: file_proto_v1_backend_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_v1_backend_proto_goTypes,
		DependencyIndexes: file_proto_v1_backend_proto_depIdxs,
		MessageInfos:      file_proto_v1_backend_proto_msgTypes,
	}.Build()
	File_proto_v1_backend_proto = out.File
	file_proto_v1_backend_proto_rawDesc = nil
	file_proto_v1_backend_proto_goTypes = nil
	file_proto_v1_backend_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// BackendServiceClient is the client API for BackendService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BackendServiceClient interface {
	// get the worker wanted state from the server
	StreamServiceControl(ctx context.Context, in *ServiceStateValues, opts ...grpc.CallOption) (BackendService_StreamServiceControlClient, error)
	// stream tasks status to prometheus
	StreamTasksStatus(ctx context.Context, opts ...grpc.CallOption) (BackendService_StreamTasksStatusClient, error)
}

type backendServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBackendServiceClient(cc grpc.ClientConnInterface) BackendServiceClient {
	return &backendServiceClient{cc}
}

func (c *backendServiceClient) StreamServiceControl(ctx context.Context, in *ServiceStateValues, opts ...grpc.CallOption) (BackendService_StreamServiceControlClient, error) {
	stream, err := c.cc.NewStream(ctx, &_BackendService_serviceDesc.Streams[0], "/proto.BackendService/StreamServiceControl", opts...)
	if err != nil {
		return nil, err
	}
	x := &backendServiceStreamServiceControlClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type BackendService_StreamServiceControlClient interface {
	Recv() (*ServiceStateValues, error)
	grpc.ClientStream
}

type backendServiceStreamServiceControlClient struct {
	grpc.ClientStream
}

func (x *backendServiceStreamServiceControlClient) Recv() (*ServiceStateValues, error) {
	m := new(ServiceStateValues)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *backendServiceClient) StreamTasksStatus(ctx context.Context, opts ...grpc.CallOption) (BackendService_StreamTasksStatusClient, error) {
	stream, err := c.cc.NewStream(ctx, &_BackendService_serviceDesc.Streams[1], "/proto.BackendService/StreamTasksStatus", opts...)
	if err != nil {
		return nil, err
	}
	x := &backendServiceStreamTasksStatusClient{stream}
	return x, nil
}

type BackendService_StreamTasksStatusClient interface {
	Send(*PrometheusStatus) error
	Recv() (*PrometheusStatus, error)
	grpc.ClientStream
}

type backendServiceStreamTasksStatusClient struct {
	grpc.ClientStream
}

func (x *backendServiceStreamTasksStatusClient) Send(m *PrometheusStatus) error {
	return x.ClientStream.SendMsg(m)
}

func (x *backendServiceStreamTasksStatusClient) Recv() (*PrometheusStatus, error) {
	m := new(PrometheusStatus)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// BackendServiceServer is the server API for BackendService service.
type BackendServiceServer interface {
	// get the worker wanted state from the server
	StreamServiceControl(*ServiceStateValues, BackendService_StreamServiceControlServer) error
	// stream tasks status to prometheus
	StreamTasksStatus(BackendService_StreamTasksStatusServer) error
}

// UnimplementedBackendServiceServer can be embedded to have forward compatible implementations.
type UnimplementedBackendServiceServer struct {
}

func (*UnimplementedBackendServiceServer) StreamServiceControl(*ServiceStateValues, BackendService_StreamServiceControlServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamServiceControl not implemented")
}
func (*UnimplementedBackendServiceServer) StreamTasksStatus(BackendService_StreamTasksStatusServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamTasksStatus not implemented")
}

func RegisterBackendServiceServer(s *grpc.Server, srv BackendServiceServer) {
	s.RegisterService(&_BackendService_serviceDesc, srv)
}

func _BackendService_StreamServiceControl_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ServiceStateValues)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BackendServiceServer).StreamServiceControl(m, &backendServiceStreamServiceControlServer{stream})
}

type BackendService_StreamServiceControlServer interface {
	Send(*ServiceStateValues) error
	grpc.ServerStream
}

type backendServiceStreamServiceControlServer struct {
	grpc.ServerStream
}

func (x *backendServiceStreamServiceControlServer) Send(m *ServiceStateValues) error {
	return x.ServerStream.SendMsg(m)
}

func _BackendService_StreamTasksStatus_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BackendServiceServer).StreamTasksStatus(&backendServiceStreamTasksStatusServer{stream})
}

type BackendService_StreamTasksStatusServer interface {
	Send(*PrometheusStatus) error
	Recv() (*PrometheusStatus, error)
	grpc.ServerStream
}

type backendServiceStreamTasksStatusServer struct {
	grpc.ServerStream
}

func (x *backendServiceStreamTasksStatusServer) Send(m *PrometheusStatus) error {
	return x.ServerStream.SendMsg(m)
}

func (x *backendServiceStreamTasksStatusServer) Recv() (*PrometheusStatus, error) {
	m := new(PrometheusStatus)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _BackendService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.BackendService",
	HandlerType: (*BackendServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamServiceControl",
			Handler:       _BackendService_StreamServiceControl_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "StreamTasksStatus",
			Handler:       _BackendService_StreamTasksStatus_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proto/v1/backend.proto",
}
