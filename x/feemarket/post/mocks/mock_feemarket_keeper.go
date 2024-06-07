// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	cosmos_sdktypes "github.com/cosmos/cosmos-sdk/types"
	mock "github.com/stretchr/testify/mock"

	types "github.com/skip-mev/feemarket/x/feemarket/types"
)

// FeeMarketKeeper is an autogenerated mock type for the FeeMarketKeeper type
type FeeMarketKeeper struct {
	mock.Mock
}

// GetDenomResolver provides a mock function with given fields:
func (_m *FeeMarketKeeper) GetDenomResolver() types.DenomResolver {
	ret := _m.Called()

	var r0 types.DenomResolver
	if rf, ok := ret.Get(0).(func() types.DenomResolver); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.DenomResolver)
		}
	}

	return r0
}

// GetMinGasPrice provides a mock function with given fields: ctx, denom
func (_m *FeeMarketKeeper) GetMinGasPrice(ctx cosmos_sdktypes.Context, denom string) (cosmos_sdktypes.DecCoin, error) {
	ret := _m.Called(ctx, denom)

	var r0 cosmos_sdktypes.DecCoin
	var r1 error
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, string) (cosmos_sdktypes.DecCoin, error)); ok {
		return rf(ctx, denom)
	}
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, string) cosmos_sdktypes.DecCoin); ok {
		r0 = rf(ctx, denom)
	} else {
		r0 = ret.Get(0).(cosmos_sdktypes.DecCoin)
	}

	if rf, ok := ret.Get(1).(func(cosmos_sdktypes.Context, string) error); ok {
		r1 = rf(ctx, denom)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetParams provides a mock function with given fields: ctx
func (_m *FeeMarketKeeper) GetParams(ctx cosmos_sdktypes.Context) (types.Params, error) {
	ret := _m.Called(ctx)

	var r0 types.Params
	var r1 error
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context) (types.Params, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context) types.Params); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(types.Params)
	}

	if rf, ok := ret.Get(1).(func(cosmos_sdktypes.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetState provides a mock function with given fields: ctx
func (_m *FeeMarketKeeper) GetState(ctx cosmos_sdktypes.Context) (types.State, error) {
	ret := _m.Called(ctx)

	var r0 types.State
	var r1 error
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context) (types.State, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context) types.State); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(types.State)
	}

	if rf, ok := ret.Get(1).(func(cosmos_sdktypes.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetParams provides a mock function with given fields: ctx, params
func (_m *FeeMarketKeeper) SetParams(ctx cosmos_sdktypes.Context, params types.Params) error {
	ret := _m.Called(ctx, params)

	var r0 error
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, types.Params) error); ok {
		r0 = rf(ctx, params)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetState provides a mock function with given fields: ctx, state
func (_m *FeeMarketKeeper) SetState(ctx cosmos_sdktypes.Context, state types.State) error {
	ret := _m.Called(ctx, state)

	var r0 error
	if rf, ok := ret.Get(0).(func(cosmos_sdktypes.Context, types.State) error); ok {
		r0 = rf(ctx, state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewFeeMarketKeeper interface {
	mock.TestingT
	Cleanup(func())
}

// NewFeeMarketKeeper creates a new instance of FeeMarketKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFeeMarketKeeper(t mockConstructorTestingTNewFeeMarketKeeper) *FeeMarketKeeper {
	mock := &FeeMarketKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
