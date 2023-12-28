// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	types "github.com/cosmos/cosmos-sdk/types"
)

// FeeMarketKeeper is an autogenerated mock type for the FeeMarketKeeper type
type FeeMarketKeeper struct {
	mock.Mock
}

// GetMinGasPrices provides a mock function with given fields: ctx
func (_m *FeeMarketKeeper) GetMinGasPrices(ctx types.Context) (types.Coins, error) {
	ret := _m.Called(ctx)

	var r0 types.Coins
	var r1 error
	if rf, ok := ret.Get(0).(func(types.Context) (types.Coins, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(types.Context) types.Coins); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.Coins)
		}
	}

	if rf, ok := ret.Get(1).(func(types.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetParams provides a mock function with given fields: ctx
func (_m *FeeMarketKeeper) GetParams(ctx types.Context) (feemarkettypes.Params, error) {
	ret := _m.Called(ctx)

	var r0 feemarkettypes.Params
	var r1 error
	if rf, ok := ret.Get(0).(func(types.Context) (feemarkettypes.Params, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(types.Context) feemarkettypes.Params); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(feemarkettypes.Params)
	}

	if rf, ok := ret.Get(1).(func(types.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetState provides a mock function with given fields: ctx
func (_m *FeeMarketKeeper) GetState(ctx types.Context) (feemarkettypes.State, error) {
	ret := _m.Called(ctx)

	var r0 feemarkettypes.State
	var r1 error
	if rf, ok := ret.Get(0).(func(types.Context) (feemarkettypes.State, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(types.Context) feemarkettypes.State); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(feemarkettypes.State)
	}

	if rf, ok := ret.Get(1).(func(types.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetParams provides a mock function with given fields: ctx, params
func (_m *FeeMarketKeeper) SetParams(ctx types.Context, params feemarkettypes.Params) error {
	ret := _m.Called(ctx, params)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.Context, feemarkettypes.Params) error); ok {
		r0 = rf(ctx, params)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetState provides a mock function with given fields: ctx, state
func (_m *FeeMarketKeeper) SetState(ctx types.Context, state feemarkettypes.State) error {
	ret := _m.Called(ctx, state)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.Context, feemarkettypes.State) error); ok {
		r0 = rf(ctx, state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewFeeMarketKeeper creates a new instance of FeeMarketKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFeeMarketKeeper(t interface {
	mock.TestingT
	Cleanup(func())
},
) *FeeMarketKeeper {
	mock := &FeeMarketKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
