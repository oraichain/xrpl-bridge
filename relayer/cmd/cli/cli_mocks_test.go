// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/CoreumFoundation/coreumbridge-xrpl/relayer/cmd/cli (interfaces: BridgeClient,Processor)

// Package cli_test is a generated GoMock package.
package cli_test

import (
	context "context"
	reflect "reflect"

	math "cosmossdk.io/math"
	types "github.com/cosmos/cosmos-sdk/types"
	gomock "github.com/golang/mock/gomock"
	data "github.com/rubblelabs/ripple/data"

	client "github.com/CoreumFoundation/coreumbridge-xrpl/relayer/client"
	coreum "github.com/CoreumFoundation/coreumbridge-xrpl/relayer/coreum"
)

// MockBridgeClient is a mock of BridgeClient interface.
type MockBridgeClient struct {
	ctrl     *gomock.Controller
	recorder *MockBridgeClientMockRecorder
}

// MockBridgeClientMockRecorder is the mock recorder for MockBridgeClient.
type MockBridgeClientMockRecorder struct {
	mock *MockBridgeClient
}

// NewMockBridgeClient creates a new mock instance.
func NewMockBridgeClient(ctrl *gomock.Controller) *MockBridgeClient {
	mock := &MockBridgeClient{ctrl: ctrl}
	mock.recorder = &MockBridgeClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBridgeClient) EXPECT() *MockBridgeClientMockRecorder {
	return m.recorder
}

// Bootstrap mocks base method.
func (m *MockBridgeClient) Bootstrap(arg0 context.Context, arg1 types.AccAddress, arg2 string, arg3 client.BootstrappingConfig) (types.AccAddress, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Bootstrap", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(types.AccAddress)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Bootstrap indicates an expected call of Bootstrap.
func (mr *MockBridgeClientMockRecorder) Bootstrap(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Bootstrap", reflect.TypeOf((*MockBridgeClient)(nil).Bootstrap), arg0, arg1, arg2, arg3)
}

// GetAllTokens mocks base method.
func (m *MockBridgeClient) GetAllTokens(arg0 context.Context) ([]coreum.CoreumToken, []coreum.XRPLToken, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllTokens", arg0)
	ret0, _ := ret[0].([]coreum.CoreumToken)
	ret1, _ := ret[1].([]coreum.XRPLToken)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetAllTokens indicates an expected call of GetAllTokens.
func (mr *MockBridgeClientMockRecorder) GetAllTokens(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllTokens", reflect.TypeOf((*MockBridgeClient)(nil).GetAllTokens), arg0)
}

// GetContractConfig mocks base method.
func (m *MockBridgeClient) GetContractConfig(arg0 context.Context) (coreum.ContractConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContractConfig", arg0)
	ret0, _ := ret[0].(coreum.ContractConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContractConfig indicates an expected call of GetContractConfig.
func (mr *MockBridgeClientMockRecorder) GetContractConfig(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContractConfig", reflect.TypeOf((*MockBridgeClient)(nil).GetContractConfig), arg0)
}

// GetCoreumBalances mocks base method.
func (m *MockBridgeClient) GetCoreumBalances(arg0 context.Context, arg1 types.AccAddress) (types.Coins, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCoreumBalances", arg0, arg1)
	ret0, _ := ret[0].(types.Coins)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCoreumBalances indicates an expected call of GetCoreumBalances.
func (mr *MockBridgeClientMockRecorder) GetCoreumBalances(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCoreumBalances", reflect.TypeOf((*MockBridgeClient)(nil).GetCoreumBalances), arg0, arg1)
}

// GetXRPLBalances mocks base method.
func (m *MockBridgeClient) GetXRPLBalances(arg0 context.Context, arg1 data.Account) ([]data.Amount, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetXRPLBalances", arg0, arg1)
	ret0, _ := ret[0].([]data.Amount)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetXRPLBalances indicates an expected call of GetXRPLBalances.
func (mr *MockBridgeClientMockRecorder) GetXRPLBalances(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetXRPLBalances", reflect.TypeOf((*MockBridgeClient)(nil).GetXRPLBalances), arg0, arg1)
}

// RecoverTickets mocks base method.
func (m *MockBridgeClient) RecoverTickets(arg0 context.Context, arg1 types.AccAddress, arg2 uint32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecoverTickets", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecoverTickets indicates an expected call of RecoverTickets.
func (mr *MockBridgeClientMockRecorder) RecoverTickets(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecoverTickets", reflect.TypeOf((*MockBridgeClient)(nil).RecoverTickets), arg0, arg1, arg2)
}

// RegisterCoreumToken mocks base method.
func (m *MockBridgeClient) RegisterCoreumToken(arg0 context.Context, arg1 types.AccAddress, arg2 string, arg3 uint32, arg4 int32, arg5 math.Int) (coreum.CoreumToken, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterCoreumToken", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(coreum.CoreumToken)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterCoreumToken indicates an expected call of RegisterCoreumToken.
func (mr *MockBridgeClientMockRecorder) RegisterCoreumToken(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterCoreumToken", reflect.TypeOf((*MockBridgeClient)(nil).RegisterCoreumToken), arg0, arg1, arg2, arg3, arg4, arg5)
}

// RegisterXRPLToken mocks base method.
func (m *MockBridgeClient) RegisterXRPLToken(arg0 context.Context, arg1 types.AccAddress, arg2 data.Account, arg3 data.Currency, arg4 int32, arg5 math.Int) (coreum.XRPLToken, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterXRPLToken", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(coreum.XRPLToken)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterXRPLToken indicates an expected call of RegisterXRPLToken.
func (mr *MockBridgeClientMockRecorder) RegisterXRPLToken(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterXRPLToken", reflect.TypeOf((*MockBridgeClient)(nil).RegisterXRPLToken), arg0, arg1, arg2, arg3, arg4, arg5)
}

// SendFromCoreumToXRPL mocks base method.
func (m *MockBridgeClient) SendFromCoreumToXRPL(arg0 context.Context, arg1 types.AccAddress, arg2 types.Coin, arg3 data.Account) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendFromCoreumToXRPL", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendFromCoreumToXRPL indicates an expected call of SendFromCoreumToXRPL.
func (mr *MockBridgeClientMockRecorder) SendFromCoreumToXRPL(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendFromCoreumToXRPL", reflect.TypeOf((*MockBridgeClient)(nil).SendFromCoreumToXRPL), arg0, arg1, arg2, arg3)
}

// SendFromXRPLToCoreum mocks base method.
func (m *MockBridgeClient) SendFromXRPLToCoreum(arg0 context.Context, arg1 string, arg2 data.Amount, arg3 types.AccAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendFromXRPLToCoreum", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendFromXRPLToCoreum indicates an expected call of SendFromXRPLToCoreum.
func (mr *MockBridgeClientMockRecorder) SendFromXRPLToCoreum(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendFromXRPLToCoreum", reflect.TypeOf((*MockBridgeClient)(nil).SendFromXRPLToCoreum), arg0, arg1, arg2, arg3)
}

// SetXRPLTrustSet mocks base method.
func (m *MockBridgeClient) SetXRPLTrustSet(arg0 context.Context, arg1 string, arg2 data.Amount) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetXRPLTrustSet", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetXRPLTrustSet indicates an expected call of SetXRPLTrustSet.
func (mr *MockBridgeClientMockRecorder) SetXRPLTrustSet(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetXRPLTrustSet", reflect.TypeOf((*MockBridgeClient)(nil).SetXRPLTrustSet), arg0, arg1, arg2)
}

// UpdateCoreumToken mocks base method.
func (m *MockBridgeClient) UpdateCoreumToken(arg0 context.Context, arg1 types.AccAddress, arg2 string, arg3 *coreum.TokenState, arg4 *int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCoreumToken", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCoreumToken indicates an expected call of UpdateCoreumToken.
func (mr *MockBridgeClientMockRecorder) UpdateCoreumToken(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCoreumToken", reflect.TypeOf((*MockBridgeClient)(nil).UpdateCoreumToken), arg0, arg1, arg2, arg3, arg4)
}

// UpdateXRPLToken mocks base method.
func (m *MockBridgeClient) UpdateXRPLToken(arg0 context.Context, arg1 types.AccAddress, arg2, arg3 string, arg4 *coreum.TokenState, arg5 *int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateXRPLToken", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateXRPLToken indicates an expected call of UpdateXRPLToken.
func (mr *MockBridgeClientMockRecorder) UpdateXRPLToken(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateXRPLToken", reflect.TypeOf((*MockBridgeClient)(nil).UpdateXRPLToken), arg0, arg1, arg2, arg3, arg4, arg5)
}

// MockProcessor is a mock of Processor interface.
type MockProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockProcessorMockRecorder
}

// MockProcessorMockRecorder is the mock recorder for MockProcessor.
type MockProcessorMockRecorder struct {
	mock *MockProcessor
}

// NewMockProcessor creates a new mock instance.
func NewMockProcessor(ctrl *gomock.Controller) *MockProcessor {
	mock := &MockProcessor{ctrl: ctrl}
	mock.recorder = &MockProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProcessor) EXPECT() *MockProcessorMockRecorder {
	return m.recorder
}

// StartAllProcesses mocks base method.
func (m *MockProcessor) StartAllProcesses(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartAllProcesses", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartAllProcesses indicates an expected call of StartAllProcesses.
func (mr *MockProcessorMockRecorder) StartAllProcesses(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartAllProcesses", reflect.TypeOf((*MockProcessor)(nil).StartAllProcesses), arg0)
}