// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: nibiru/inflation/v1/tx.proto

package inflationv1

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

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	// ToggleInflation defines a method to enable or disable inflation.
	ToggleInflation(ctx context.Context, in *MsgToggleInflation, opts ...grpc.CallOption) (*MsgToggleInflationResponse, error)
	// EditInflationParams defines a method to edit the inflation params.
	EditInflationParams(ctx context.Context, in *MsgEditInflationParams, opts ...grpc.CallOption) (*MsgEditInflationParamsResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) ToggleInflation(ctx context.Context, in *MsgToggleInflation, opts ...grpc.CallOption) (*MsgToggleInflationResponse, error) {
	out := new(MsgToggleInflationResponse)
	err := c.cc.Invoke(ctx, "/nibiru.inflation.v1.Msg/ToggleInflation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) EditInflationParams(ctx context.Context, in *MsgEditInflationParams, opts ...grpc.CallOption) (*MsgEditInflationParamsResponse, error) {
	out := new(MsgEditInflationParamsResponse)
	err := c.cc.Invoke(ctx, "/nibiru.inflation.v1.Msg/EditInflationParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
type MsgServer interface {
	// ToggleInflation defines a method to enable or disable inflation.
	ToggleInflation(context.Context, *MsgToggleInflation) (*MsgToggleInflationResponse, error)
	// EditInflationParams defines a method to edit the inflation params.
	EditInflationParams(context.Context, *MsgEditInflationParams) (*MsgEditInflationParamsResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) ToggleInflation(context.Context, *MsgToggleInflation) (*MsgToggleInflationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleInflation not implemented")
}
func (UnimplementedMsgServer) EditInflationParams(context.Context, *MsgEditInflationParams) (*MsgEditInflationParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EditInflationParams not implemented")
}
func (UnimplementedMsgServer) mustEmbedUnimplementedMsgServer() {}

// UnsafeMsgServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MsgServer will
// result in compilation errors.
type UnsafeMsgServer interface {
	mustEmbedUnimplementedMsgServer()
}

func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&Msg_ServiceDesc, srv)
}

func _Msg_ToggleInflation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgToggleInflation)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ToggleInflation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/nibiru.inflation.v1.Msg/ToggleInflation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ToggleInflation(ctx, req.(*MsgToggleInflation))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_EditInflationParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgEditInflationParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).EditInflationParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/nibiru.inflation.v1.Msg/EditInflationParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).EditInflationParams(ctx, req.(*MsgEditInflationParams))
	}
	return interceptor(ctx, in, info, handler)
}

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "nibiru.inflation.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ToggleInflation",
			Handler:    _Msg_ToggleInflation_Handler,
		},
		{
			MethodName: "EditInflationParams",
			Handler:    _Msg_EditInflationParams_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "nibiru/inflation/v1/tx.proto",
}
