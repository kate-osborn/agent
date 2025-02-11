//  Copyright (c) F5, Inc.
//
//  This source code is licensed under the Apache License, Version 2.0 license found in the
//  LICENSE file in the root directory of this source tree.
//

// Code generated by MockGen. DO NOT EDIT.
// Source: ingester.go
//
// Generated by this command:
//
//	mockgen -source ingester.go -destination mocks/ingester_mock.go -package mocks -copyright_file=../../../../COPYRIGHT
//
// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	tables "github.com/nginx/agent/v2/src/extensions/advanced-metrics/tables"
	gomock "go.uber.org/mock/gomock"
)

// MockStagingTable is a mock of StagingTable interface.
type MockStagingTable struct {
	ctrl     *gomock.Controller
	recorder *MockStagingTableMockRecorder
}

// MockStagingTableMockRecorder is the mock recorder for MockStagingTable.
type MockStagingTableMockRecorder struct {
	mock *MockStagingTable
}

// NewMockStagingTable creates a new mock instance.
func NewMockStagingTable(ctrl *gomock.Controller) *MockStagingTable {
	mock := &MockStagingTable{ctrl: ctrl}
	mock.recorder = &MockStagingTableMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStagingTable) EXPECT() *MockStagingTableMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockStagingTable) Add(arg0 tables.FieldIterator) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockStagingTableMockRecorder) Add(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockStagingTable)(nil).Add), arg0)
}
