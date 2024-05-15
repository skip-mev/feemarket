// Code generated by mockery v2.43.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper is an autogenerated mock type for the AccountKeeper type
type AccountKeeper struct {
	mock.Mock
}

// GetAccount provides a mock function with given fields: ctx, addr
func (_m *AccountKeeper) GetAccount(ctx context.Context, addr types.AccAddress) types.AccountI {
	ret := _m.Called(ctx, addr)

	if len(ret) == 0 {
		panic("no return value specified for GetAccount")
	}

	var r0 types.AccountI
	if rf, ok := ret.Get(0).(func(context.Context, types.AccAddress) types.AccountI); ok {
		r0 = rf(ctx, addr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.AccountI)
		}
	}

	return r0
}

// GetModuleAccount provides a mock function with given fields: ctx, name
func (_m *AccountKeeper) GetModuleAccount(ctx context.Context, name string) types.ModuleAccountI {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for GetModuleAccount")
	}

	var r0 types.ModuleAccountI
	if rf, ok := ret.Get(0).(func(context.Context, string) types.ModuleAccountI); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.ModuleAccountI)
		}
	}

	return r0
}

// GetModuleAddress provides a mock function with given fields: name
func (_m *AccountKeeper) GetModuleAddress(name string) types.AccAddress {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for GetModuleAddress")
	}

	var r0 types.AccAddress
	if rf, ok := ret.Get(0).(func(string) types.AccAddress); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.AccAddress)
		}
	}

	return r0
}

// NewAccountKeeper creates a new instance of AccountKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAccountKeeper(t interface {
	mock.TestingT
	Cleanup(func())
},
) *AccountKeeper {
	mock := &AccountKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
