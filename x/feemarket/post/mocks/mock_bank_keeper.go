// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper is an autogenerated mock type for the BankKeeper type
type BankKeeper struct {
	mock.Mock
}

// IsSendEnabledCoins provides a mock function with given fields: ctx, coins
func (_m *BankKeeper) IsSendEnabledCoins(ctx context.Context, coins ...types.Coin) error {
	_va := make([]interface{}, len(coins))
	for _i := range coins {
		_va[_i] = coins[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for IsSendEnabledCoins")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...types.Coin) error); ok {
		r0 = rf(ctx, coins...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendCoins provides a mock function with given fields: ctx, from, to, amt
func (_m *BankKeeper) SendCoins(ctx context.Context, from types.AccAddress, to types.AccAddress, amt types.Coins) error {
	ret := _m.Called(ctx, from, to, amt)

	if len(ret) == 0 {
		panic("no return value specified for SendCoins")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.AccAddress, types.AccAddress, types.Coins) error); ok {
		r0 = rf(ctx, from, to, amt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendCoinsFromAccountToModule provides a mock function with given fields: ctx, senderAddr, recipientModule, amt
func (_m *BankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr types.AccAddress, recipientModule string, amt types.Coins) error {
	ret := _m.Called(ctx, senderAddr, recipientModule, amt)

	if len(ret) == 0 {
		panic("no return value specified for SendCoinsFromAccountToModule")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.AccAddress, string, types.Coins) error); ok {
		r0 = rf(ctx, senderAddr, recipientModule, amt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewBankKeeper creates a new instance of BankKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBankKeeper(t interface {
	mock.TestingT
	Cleanup(func())
},
) *BankKeeper {
	mock := &BankKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
