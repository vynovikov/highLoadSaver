// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.1
// source: internal/adapters/driven/recorder/proto/svc.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// LoggerClient is the client API for Logger service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LoggerClient interface {
	Log(ctx context.Context, in *LogReq, opts ...grpc.CallOption) (*LogRes, error)
}

type loggerClient struct {
	cc grpc.ClientConnInterface
}

func NewLoggerClient(cc grpc.ClientConnInterface) LoggerClient {
	return &loggerClient{cc}
}

func (c *loggerClient) Log(ctx context.Context, in *LogReq, opts ...grpc.CallOption) (*LogRes, error) {
	out := new(LogRes)
	err := c.cc.Invoke(ctx, "/recorder.Logger/Log", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LoggerServer is the server API for Logger service.
// All implementations must embed UnimplementedLoggerServer
// for forward compatibility
type LoggerServer interface {
	Log(context.Context, *LogReq) (*LogRes, error)
	mustEmbedUnimplementedLoggerServer()
}

// UnimplementedLoggerServer must be embedded to have forward compatible implementations.
type UnimplementedLoggerServer struct {
}

func (UnimplementedLoggerServer) Log(context.Context, *LogReq) (*LogRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Log not implemented")
}
func (UnimplementedLoggerServer) mustEmbedUnimplementedLoggerServer() {}

// UnsafeLoggerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LoggerServer will
// result in compilation errors.
type UnsafeLoggerServer interface {
	mustEmbedUnimplementedLoggerServer()
}

func RegisterLoggerServer(s grpc.ServiceRegistrar, srv LoggerServer) {
	s.RegisterService(&Logger_ServiceDesc, srv)
}

func _Logger_Log_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LoggerServer).Log(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/recorder.Logger/Log",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LoggerServer).Log(ctx, req.(*LogReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Logger_ServiceDesc is the grpc.ServiceDesc for Logger service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Logger_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "recorder.Logger",
	HandlerType: (*LoggerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Log",
			Handler:    _Logger_Log_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/adapters/driven/recorder/proto/svc.proto",
}
