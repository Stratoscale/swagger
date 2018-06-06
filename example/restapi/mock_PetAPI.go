// Code generated by mockery v1.0.0
package restapi

import context "context"
import middleware "github.com/go-openapi/runtime/middleware"
import mock "github.com/stretchr/testify/mock"
import pet "github.com/Stratoscale/swagger/example/restapi/operations/pet"

// MockPetAPI is an autogenerated mock type for the PetAPI type
type MockPetAPI struct {
	mock.Mock
}

// PetCreate provides a mock function with given fields: ctx, params
func (_m *MockPetAPI) PetCreate(ctx context.Context, params pet.PetCreateParams) middleware.Responder {
	ret := _m.Called(ctx, params)

	var r0 middleware.Responder
	if rf, ok := ret.Get(0).(func(context.Context, pet.PetCreateParams) middleware.Responder); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(middleware.Responder)
		}
	}

	return r0
}

// PetDelete provides a mock function with given fields: ctx, params
func (_m *MockPetAPI) PetDelete(ctx context.Context, params pet.PetDeleteParams) middleware.Responder {
	ret := _m.Called(ctx, params)

	var r0 middleware.Responder
	if rf, ok := ret.Get(0).(func(context.Context, pet.PetDeleteParams) middleware.Responder); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(middleware.Responder)
		}
	}

	return r0
}

// PetGet provides a mock function with given fields: ctx, params
func (_m *MockPetAPI) PetGet(ctx context.Context, params pet.PetGetParams) middleware.Responder {
	ret := _m.Called(ctx, params)

	var r0 middleware.Responder
	if rf, ok := ret.Get(0).(func(context.Context, pet.PetGetParams) middleware.Responder); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(middleware.Responder)
		}
	}

	return r0
}

// PetList provides a mock function with given fields: ctx, params
func (_m *MockPetAPI) PetList(ctx context.Context, params pet.PetListParams) middleware.Responder {
	ret := _m.Called(ctx, params)

	var r0 middleware.Responder
	if rf, ok := ret.Get(0).(func(context.Context, pet.PetListParams) middleware.Responder); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(middleware.Responder)
		}
	}

	return r0
}

// PetUpdate provides a mock function with given fields: ctx, params
func (_m *MockPetAPI) PetUpdate(ctx context.Context, params pet.PetUpdateParams) middleware.Responder {
	ret := _m.Called(ctx, params)

	var r0 middleware.Responder
	if rf, ok := ret.Get(0).(func(context.Context, pet.PetUpdateParams) middleware.Responder); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(middleware.Responder)
		}
	}

	return r0
}
