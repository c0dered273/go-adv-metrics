// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/c0dered273/go-adv-metrics/internal/storage (interfaces: Repository)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	metric "github.com/c0dered273/go-adv-metrics/internal/metric"
	gomock "github.com/golang/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockRepository) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockRepositoryMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockRepository)(nil).Close))
}

// FindAll mocks base method.
func (m *MockRepository) FindAll(arg0 context.Context) ([]metric.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindAll", arg0)
	ret0, _ := ret[0].([]metric.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindAll indicates an expected call of FindAll.
func (mr *MockRepositoryMockRecorder) FindAll(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindAll", reflect.TypeOf((*MockRepository)(nil).FindAll), arg0)
}

// FindByID mocks base method.
func (m *MockRepository) FindByID(arg0 context.Context, arg1 metric.Metric) (metric.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", arg0, arg1)
	ret0, _ := ret[0].(metric.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID.
func (mr *MockRepositoryMockRecorder) FindByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockRepository)(nil).FindByID), arg0, arg1)
}

// Ping mocks base method.
func (m *MockRepository) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockRepositoryMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockRepository)(nil).Ping))
}

// Save mocks base method.
func (m *MockRepository) Save(arg0 context.Context, arg1 metric.Metric) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockRepositoryMockRecorder) Save(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockRepository)(nil).Save), arg0, arg1)
}

// SaveAll mocks base method.
func (m *MockRepository) SaveAll(arg0 context.Context, arg1 []metric.Metric) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveAll", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveAll indicates an expected call of SaveAll.
func (mr *MockRepositoryMockRecorder) SaveAll(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveAll", reflect.TypeOf((*MockRepository)(nil).SaveAll), arg0, arg1)
}
