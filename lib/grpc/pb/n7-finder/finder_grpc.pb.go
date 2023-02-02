// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.3
// source: finder.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FinderClient is the client API for Finder service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FinderClient interface {
	LookupTransaction(ctx context.Context, in *wrapperspb.StringValue, opts ...grpc.CallOption) (*Stock, error)
}

type finderClient struct {
	cc grpc.ClientConnInterface
}

func NewFinderClient(cc grpc.ClientConnInterface) FinderClient {
	return &finderClient{cc}
}

func (c *finderClient) LookupTransaction(ctx context.Context, in *wrapperspb.StringValue, opts ...grpc.CallOption) (*Stock, error) {
	out := new(Stock)
	err := c.cc.Invoke(ctx, "/finder.Finder/LookupTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FinderServer is the server API for Finder service.
// All implementations must embed UnimplementedFinderServer
// for forward compatibility
type FinderServer interface {
	LookupTransaction(context.Context, *wrapperspb.StringValue) (*Stock, error)
	mustEmbedUnimplementedFinderServer()
}

// UnimplementedFinderServer must be embedded to have forward compatible implementations.
type UnimplementedFinderServer struct {
}

func (UnimplementedFinderServer) LookupTransaction(context.Context, *wrapperspb.StringValue) (*Stock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LookupTransaction not implemented")
}
func (UnimplementedFinderServer) mustEmbedUnimplementedFinderServer() {}

// UnsafeFinderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FinderServer will
// result in compilation errors.
type UnsafeFinderServer interface {
	mustEmbedUnimplementedFinderServer()
}

func RegisterFinderServer(s grpc.ServiceRegistrar, srv FinderServer) {
	s.RegisterService(&Finder_ServiceDesc, srv)
}

func _Finder_LookupTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(wrapperspb.StringValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FinderServer).LookupTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/finder.Finder/LookupTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FinderServer).LookupTransaction(ctx, req.(*wrapperspb.StringValue))
	}
	return interceptor(ctx, in, info, handler)
}

// Finder_ServiceDesc is the grpc.ServiceDesc for Finder service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Finder_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "finder.Finder",
	HandlerType: (*FinderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LookupTransaction",
			Handler:    _Finder_LookupTransaction_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "finder.proto",
}