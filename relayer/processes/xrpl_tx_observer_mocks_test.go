// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/CoreumFoundation/coreumbridge-xrpl/relayer/processes (interfaces: EvidencesConsumer,XRPLAccountTxScanner)

// Package processes_test is a generated GoMock package.
package processes_test

import (
	context "context"
	reflect "reflect"

	types "github.com/cosmos/cosmos-sdk/types"
	gomock "github.com/golang/mock/gomock"
	data "github.com/rubblelabs/ripple/data"

	coreum "github.com/CoreumFoundation/coreumbridge-xrpl/relayer/coreum"
)

// MockEvidencesConsumer is a mock of EvidencesConsumer interface.
type MockEvidencesConsumer struct {
	ctrl     *gomock.Controller
	recorder *MockEvidencesConsumerMockRecorder
}

// MockEvidencesConsumerMockRecorder is the mock recorder for MockEvidencesConsumer.
type MockEvidencesConsumerMockRecorder struct {
	mock *MockEvidencesConsumer
}

// NewMockEvidencesConsumer creates a new mock instance.
func NewMockEvidencesConsumer(ctrl *gomock.Controller) *MockEvidencesConsumer {
	mock := &MockEvidencesConsumer{ctrl: ctrl}
	mock.recorder = &MockEvidencesConsumerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEvidencesConsumer) EXPECT() *MockEvidencesConsumerMockRecorder {
	return m.recorder
}

// AcceptXRPLToCoreumEvidence mocks base method.
func (m *MockEvidencesConsumer) AcceptXRPLToCoreumEvidence(arg0 context.Context, arg1 types.AccAddress, arg2 coreum.XRPLToCoreumEvidence) (*types.TxResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AcceptXRPLToCoreumEvidence", arg0, arg1, arg2)
	ret0, _ := ret[0].(*types.TxResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AcceptXRPLToCoreumEvidence indicates an expected call of AcceptXRPLToCoreumEvidence.
func (mr *MockEvidencesConsumerMockRecorder) AcceptXRPLToCoreumEvidence(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AcceptXRPLToCoreumEvidence", reflect.TypeOf((*MockEvidencesConsumer)(nil).AcceptXRPLToCoreumEvidence), arg0, arg1, arg2)
}

// IsInitialized mocks base method.
func (m *MockEvidencesConsumer) IsInitialized() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsInitialized")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsInitialized indicates an expected call of IsInitialized.
func (mr *MockEvidencesConsumerMockRecorder) IsInitialized() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsInitialized", reflect.TypeOf((*MockEvidencesConsumer)(nil).IsInitialized))
}

// MockXRPLAccountTxScanner is a mock of XRPLAccountTxScanner interface.
type MockXRPLAccountTxScanner struct {
	ctrl     *gomock.Controller
	recorder *MockXRPLAccountTxScannerMockRecorder
}

// MockXRPLAccountTxScannerMockRecorder is the mock recorder for MockXRPLAccountTxScanner.
type MockXRPLAccountTxScannerMockRecorder struct {
	mock *MockXRPLAccountTxScanner
}

// NewMockXRPLAccountTxScanner creates a new mock instance.
func NewMockXRPLAccountTxScanner(ctrl *gomock.Controller) *MockXRPLAccountTxScanner {
	mock := &MockXRPLAccountTxScanner{ctrl: ctrl}
	mock.recorder = &MockXRPLAccountTxScannerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockXRPLAccountTxScanner) EXPECT() *MockXRPLAccountTxScannerMockRecorder {
	return m.recorder
}

// ScanTxs mocks base method.
func (m *MockXRPLAccountTxScanner) ScanTxs(arg0 context.Context, arg1 chan<- data.TransactionWithMetaData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScanTxs", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ScanTxs indicates an expected call of ScanTxs.
func (mr *MockXRPLAccountTxScannerMockRecorder) ScanTxs(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScanTxs", reflect.TypeOf((*MockXRPLAccountTxScanner)(nil).ScanTxs), arg0, arg1)
}
