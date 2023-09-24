// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/irvankadhafi/go-point-of-sales/internal/model (interfaces: RBACRepository)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "github.com/irvankadhafi/go-point-of-sales/internal/model"
	rbac "github.com/irvankadhafi/go-point-of-sales/rbac"
)

// MockRBACRepository is a mock of RBACRepository interface.
type MockRBACRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRBACRepositoryMockRecorder
}

// MockRBACRepositoryMockRecorder is the mock recorder for MockRBACRepository.
type MockRBACRepositoryMockRecorder struct {
	mock *MockRBACRepository
}

// NewMockRBACRepository creates a new mock instance.
func NewMockRBACRepository(ctrl *gomock.Controller) *MockRBACRepository {
	mock := &MockRBACRepository{ctrl: ctrl}
	mock.recorder = &MockRBACRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRBACRepository) EXPECT() *MockRBACRepositoryMockRecorder {
	return m.recorder
}

// CreateRoleResourceAction mocks base method.
func (m *MockRBACRepository) CreateRoleResourceAction(arg0 context.Context, arg1 *model.RoleResourceAction) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRoleResourceAction", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateRoleResourceAction indicates an expected call of CreateRoleResourceAction.
func (mr *MockRBACRepositoryMockRecorder) CreateRoleResourceAction(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRoleResourceAction", reflect.TypeOf((*MockRBACRepository)(nil).CreateRoleResourceAction), arg0, arg1)
}

// LoadPermission mocks base method.
func (m *MockRBACRepository) LoadPermission(arg0 context.Context) (*rbac.Permission, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadPermission", arg0)
	ret0, _ := ret[0].(*rbac.Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadPermission indicates an expected call of LoadPermission.
func (mr *MockRBACRepositoryMockRecorder) LoadPermission(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadPermission", reflect.TypeOf((*MockRBACRepository)(nil).LoadPermission), arg0)
}
