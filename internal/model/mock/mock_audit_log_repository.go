// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/irvankadhafi/go-point-of-sales/internal/model (interfaces: AuditRepository)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "github.com/irvankadhafi/go-point-of-sales/internal/model"
	gorm "gorm.io/gorm"
)

// MockAuditRepository is a mock of AuditRepository interface.
type MockAuditRepository struct {
	ctrl     *gomock.Controller
	recorder *MockAuditRepositoryMockRecorder
}

// MockAuditRepositoryMockRecorder is the mock recorder for MockAuditRepository.
type MockAuditRepositoryMockRecorder struct {
	mock *MockAuditRepository
}

// NewMockAuditRepository creates a new mock instance.
func NewMockAuditRepository(ctrl *gomock.Controller) *MockAuditRepository {
	mock := &MockAuditRepository{ctrl: ctrl}
	mock.recorder = &MockAuditRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuditRepository) EXPECT() *MockAuditRepositoryMockRecorder {
	return m.recorder
}

// Audit mocks base method.
func (m *MockAuditRepository) Audit(arg0 *gorm.DB, arg1 interface{}, arg2 *model.Audit) (*gorm.DB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Audit", arg0, arg1, arg2)
	ret0, _ := ret[0].(*gorm.DB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Audit indicates an expected call of Audit.
func (mr *MockAuditRepositoryMockRecorder) Audit(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Audit", reflect.TypeOf((*MockAuditRepository)(nil).Audit), arg0, arg1, arg2)
}