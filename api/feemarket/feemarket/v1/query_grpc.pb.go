// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: feemarket/feemarket/v1/query.proto

package feemarketv1

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

const (
	Query_Params_FullMethodName  = "/feemarket.feemarket.v1.Query/Params"
	Query_State_FullMethodName   = "/feemarket.feemarket.v1.Query/State"
	Query_BaseFee_FullMethodName = "/feemarket.feemarket.v1.Query/BaseGasPrice"
)

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type QueryClient interface {
	// Params returns the current feemarket module parameters.
	Params(ctx context.Context, in *ParamsRequest, opts ...grpc.CallOption) (*ParamsResponse, error)
	// State returns the current feemarket module state.
	State(ctx context.Context, in *StateRequest, opts ...grpc.CallOption) (*StateResponse, error)
	// BaseFee returns the current feemarket module base fee.
	BaseFee(ctx context.Context, in *BaseFeeRequest, opts ...grpc.CallOption) (*BaseFeeResponse, error)
}

type queryClient struct {
	cc grpc.ClientConnInterface
}

func NewQueryClient(cc grpc.ClientConnInterface) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Params(ctx context.Context, in *ParamsRequest, opts ...grpc.CallOption) (*ParamsResponse, error) {
	out := new(ParamsResponse)
	err := c.cc.Invoke(ctx, Query_Params_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) State(ctx context.Context, in *StateRequest, opts ...grpc.CallOption) (*StateResponse, error) {
	out := new(StateResponse)
	err := c.cc.Invoke(ctx, Query_State_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) BaseFee(ctx context.Context, in *BaseFeeRequest, opts ...grpc.CallOption) (*BaseFeeResponse, error) {
	out := new(BaseFeeResponse)
	err := c.cc.Invoke(ctx, Query_BaseFee_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
// All implementations must embed UnimplementedQueryServer
// for forward compatibility
type QueryServer interface {
	// Params returns the current feemarket module parameters.
	Params(context.Context, *ParamsRequest) (*ParamsResponse, error)
	// State returns the current feemarket module state.
	State(context.Context, *StateRequest) (*StateResponse, error)
	// BaseFee returns the current feemarket module base fee.
	BaseFee(context.Context, *BaseFeeRequest) (*BaseFeeResponse, error)
	mustEmbedUnimplementedQueryServer()
}

// UnimplementedQueryServer must be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (UnimplementedQueryServer) Params(context.Context, *ParamsRequest) (*ParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (UnimplementedQueryServer) State(context.Context, *StateRequest) (*StateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method State not implemented")
}
func (UnimplementedQueryServer) BaseFee(context.Context, *BaseFeeRequest) (*BaseFeeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BaseGasPrice not implemented")
}
func (UnimplementedQueryServer) mustEmbedUnimplementedQueryServer() {}

// UnsafeQueryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QueryServer will
// result in compilation errors.
type UnsafeQueryServer interface {
	mustEmbedUnimplementedQueryServer()
}

func RegisterQueryServer(s grpc.ServiceRegistrar, srv QueryServer) {
	s.RegisterService(&Query_ServiceDesc, srv)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_Params_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*ParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_State_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).State(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_State_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).State(ctx, req.(*StateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_BaseFee_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BaseFeeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).BaseFee(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_BaseFee_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).BaseFee(ctx, req.(*BaseFeeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Query_ServiceDesc is the grpc.ServiceDesc for Query service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Query_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "feemarket.feemarket.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Params",
			Handler:    _Query_Params_Handler,
		},
		{
			MethodName: "State",
			Handler:    _Query_State_Handler,
		},
		{
			MethodName: "BaseGasPrice",
			Handler:    _Query_BaseFee_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "feemarket/feemarket/v1/query.proto",
}
