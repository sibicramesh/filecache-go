// Code generated by MockGen. DO NOT EDIT.
// Source: internal/pkg/filecache/interfaces.go

// Package mockfilecache is a generated GoMock package.
package mockfilecache

import (
	context "context"
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockQueue is a mock of Queue interface.
// nolint
type MockQueue struct {
	ctrl     *gomock.Controller
	recorder *MockQueueMockRecorder
}

// MockQueueMockRecorder is the mock recorder for MockQueue.
// nolint
type MockQueueMockRecorder struct {
	mock *MockQueue
}

// NewMockQueue creates a new mock instance.
// nolint
func NewMockQueue(ctrl *gomock.Controller) *MockQueue {
	mock := &MockQueue{ctrl: ctrl}
	mock.recorder = &MockQueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
// nolint
func (m *MockQueue) EXPECT() *MockQueueMockRecorder {
	return m.recorder
}

// Close mocks base method.
// nolint
func (m *MockQueue) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
// nolint
func (mr *MockQueueMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockQueue)(nil).Close))
}

// Decode mocks base method.
// nolint
func (m *MockQueue) Decode(data []byte) ([]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Decode", data)
	ret0, _ := ret[0].([]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Decode indicates an expected call of Decode.
// nolint
func (mr *MockQueueMockRecorder) Decode(data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Decode", reflect.TypeOf((*MockQueue)(nil).Decode), data)
}

// Dequeue mocks base method.
// nolint
func (m *MockQueue) Dequeue(element interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dequeue", element)
	ret0, _ := ret[0].(error)
	return ret0
}

// Dequeue indicates an expected call of Dequeue.
// nolint
func (mr *MockQueueMockRecorder) Dequeue(element interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dequeue", reflect.TypeOf((*MockQueue)(nil).Dequeue), element)
}

// Destroy mocks base method.
// nolint
func (m *MockQueue) Destroy() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Destroy")
	ret0, _ := ret[0].(error)
	return ret0
}

// Destroy indicates an expected call of Destroy.
// nolint
func (mr *MockQueueMockRecorder) Destroy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Destroy", reflect.TypeOf((*MockQueue)(nil).Destroy))
}

// Encode mocks base method.
// nolint
func (m *MockQueue) Encode() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Encode")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Encode indicates an expected call of Encode.
// nolint
func (mr *MockQueueMockRecorder) Encode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Encode", reflect.TypeOf((*MockQueue)(nil).Encode))
}

// EncodeToReader mocks base method.
// nolint
func (m *MockQueue) EncodeToReader() (io.Reader, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EncodeToReader")
	ret0, _ := ret[0].(io.Reader)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// EncodeToReader indicates an expected call of EncodeToReader.
// nolint
func (mr *MockQueueMockRecorder) EncodeToReader() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EncodeToReader", reflect.TypeOf((*MockQueue)(nil).EncodeToReader))
}

// Enqueue mocks base method.
// nolint
func (m *MockQueue) Enqueue(element interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enqueue", element)
	ret0, _ := ret[0].(error)
	return ret0
}

// Enqueue indicates an expected call of Enqueue.
// nolint
func (mr *MockQueueMockRecorder) Enqueue(element interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enqueue", reflect.TypeOf((*MockQueue)(nil).Enqueue), element)
}

// Errors mocks base method.
// nolint
func (m *MockQueue) Errors() chan error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Errors")
	ret0, _ := ret[0].(chan error)
	return ret0
}

// Errors indicates an expected call of Errors.
// nolint
func (mr *MockQueueMockRecorder) Errors() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Errors", reflect.TypeOf((*MockQueue)(nil).Errors))
}

// Name mocks base method.
// nolint
func (m *MockQueue) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
// nolint
func (mr *MockQueueMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockQueue)(nil).Name))
}

// Run mocks base method.
// nolint
func (m *MockQueue) Run(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
// nolint
func (mr *MockQueueMockRecorder) Run(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockQueue)(nil).Run), ctx)
}
