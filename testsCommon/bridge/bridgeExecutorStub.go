package bridge

import (
	"context"
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/batchProcessor"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// BridgeExecutorStub -
type BridgeExecutorStub struct {
	functionCalledCounter map[string]int
	mutExecutor           sync.RWMutex
	fullPath              string

	PrintInfoCalled                                     func(logLevel logger.LogLevel, message string, extras ...interface{})
	MyTurnAsLeaderCalled                                func() bool
	GetBatchFromKCCalled                                func(ctx context.Context) (*bridgeCore.TransferBatch, error)
	StoreBatchFromKCCalled                              func(batch *bridgeCore.TransferBatch) error
	GetStoredBatchCalled                                func() *bridgeCore.TransferBatch
	GetLastExecutedEthBatchIDFromKCCalled               func(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatchCalled func(ctx context.Context) error
	GetAndStoreActionIDForProposeTransferOnKCCalled     func(ctx context.Context) (uint64, error)
	GetAndStoreActionIDForProposeSetStatusFromKCCalled  func(ctx context.Context) (uint64, error)
	GetStoredActionIDCalled                             func() uint64
	WasTransferProposedOnKCCalled                       func(ctx context.Context) (bool, error)
	ProposeTransferOnKCCalled                           func(ctx context.Context) error
	ProcessMaxRetriesOnWasTransferProposedOnKCCalled    func() bool
	ResetRetriesOnWasTransferProposedOnKCCalled         func()
	WasSetStatusProposedOnKCCalled                      func(ctx context.Context) (bool, error)
	ProposeSetStatusOnKCCalled                          func(ctx context.Context) error
	WasActionSignedOnKCCalled                           func(ctx context.Context) (bool, error)
	SignActionOnKCCalled                                func(ctx context.Context) error
	ProcessQuorumReachedOnKCCalled                      func(ctx context.Context) (bool, error)
	WasActionPerformedOnKCCalled                        func(ctx context.Context) (bool, error)
	PerformActionOnKCCalled                             func(ctx context.Context) error
	ResolveNewDepositsStatusesCalled                    func(numDeposits uint64)
	ProcessMaxQuorumRetriesOnKCCalled                   func() bool
	ResetRetriesCountOnKCCalled                         func()
	GetAndStoreBatchFromEthereumCalled                  func(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnEthereumCalled                func(ctx context.Context) (bool, error)
	SignTransferOnEthereumCalled                        func() error
	PerformTransferOnEthereumCalled                     func(ctx context.Context) error
	ProcessQuorumReachedOnEthereumCalled                func(ctx context.Context) (bool, error)
	WaitForTransferConfirmationCalled                   func(ctx context.Context)
	WaitAndReturnFinalBatchStatusesCalled               func(ctx context.Context) []byte
	GetBatchStatusesFromEthereumCalled                  func(ctx context.Context) ([]byte, error)
	ProcessMaxQuorumRetriesOnEthereumCalled             func() bool
	ResetRetriesCountOnEthereumCalled                   func()
	ClearStoredP2PSignaturesForEthereumCalled           func()
	CheckKCClientAvailabilityCalled                     func(ctx context.Context) error
	CheckEthereumClientAvailabilityCalled               func(ctx context.Context) error
	CheckAvailableTokensCalled                          func(ctx context.Context, ethTokens []common.Address, kdaTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error
}

// NewBridgeExecutorStub creates a new BridgeExecutorStub instance
func NewBridgeExecutorStub() *BridgeExecutorStub {
	return &BridgeExecutorStub{
		functionCalledCounter: make(map[string]int),
		fullPath:              "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge.(*BridgeExecutorStub).",
	}
}

// PrintInfo -
func (stub *BridgeExecutorStub) PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{}) {
	stub.incrementFunctionCounter()
	if stub.PrintInfoCalled != nil {
		stub.PrintInfoCalled(logLevel, message, extras...)
	}
}

// MyTurnAsLeader -
func (stub *BridgeExecutorStub) MyTurnAsLeader() bool {
	stub.incrementFunctionCounter()
	if stub.MyTurnAsLeaderCalled != nil {
		return stub.MyTurnAsLeaderCalled()
	}
	return false
}

// GetBatchFromKC -
func (stub *BridgeExecutorStub) GetBatchFromKC(ctx context.Context) (*bridgeCore.TransferBatch, error) {
	stub.incrementFunctionCounter()
	if stub.GetBatchFromKCCalled != nil {
		return stub.GetBatchFromKCCalled(ctx)
	}
	return nil, notImplemented
}

// StoreBatchFromKC -
func (stub *BridgeExecutorStub) StoreBatchFromKC(batch *bridgeCore.TransferBatch) error {
	stub.incrementFunctionCounter()
	if stub.StoreBatchFromKCCalled != nil {
		return stub.StoreBatchFromKCCalled(batch)
	}
	return notImplemented
}

// GetStoredBatch -
func (stub *BridgeExecutorStub) GetStoredBatch() *bridgeCore.TransferBatch {
	stub.incrementFunctionCounter()
	if stub.GetStoredBatchCalled != nil {
		return stub.GetStoredBatchCalled()
	}
	return nil
}

// GetLastExecutedEthBatchIDFromKC -
func (stub *BridgeExecutorStub) GetLastExecutedEthBatchIDFromKC(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetLastExecutedEthBatchIDFromKCCalled != nil {
		return stub.GetLastExecutedEthBatchIDFromKCCalled(ctx)
	}
	return 0, notImplemented
}

// VerifyLastDepositNonceExecutedOnEthereumBatch -
func (stub *BridgeExecutorStub) VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled != nil {
		return stub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled(ctx)
	}
	return notImplemented
}

// GetAndStoreActionIDForProposeTransferOnKC -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeTransferOnKC(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeTransferOnKCCalled != nil {
		return stub.GetAndStoreActionIDForProposeTransferOnKCCalled(ctx)
	}
	return 0, notImplemented
}

// GetAndStoreActionIDForProposeSetStatusFromKC -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeSetStatusFromKC(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeSetStatusFromKCCalled != nil {
		return stub.GetAndStoreActionIDForProposeSetStatusFromKCCalled(ctx)
	}
	return 0, notImplemented
}

// GetStoredActionID -
func (stub *BridgeExecutorStub) GetStoredActionID() uint64 {
	stub.incrementFunctionCounter()
	if stub.GetStoredActionIDCalled != nil {
		return stub.GetStoredActionIDCalled()
	}
	return 0
}

// WasTransferProposedOnKC -
func (stub *BridgeExecutorStub) WasTransferProposedOnKC(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferProposedOnKCCalled != nil {
		return stub.WasTransferProposedOnKCCalled(ctx)
	}
	return false, notImplemented
}

// ProposeTransferOnKC -
func (stub *BridgeExecutorStub) ProposeTransferOnKC(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeTransferOnKCCalled != nil {
		return stub.ProposeTransferOnKCCalled(ctx)
	}
	return notImplemented
}

// ProcessMaxRetriesOnWasTransferProposedOnKC -
func (stub *BridgeExecutorStub) ProcessMaxRetriesOnWasTransferProposedOnKC() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxRetriesOnWasTransferProposedOnKCCalled != nil {
		return stub.ProcessMaxRetriesOnWasTransferProposedOnKCCalled()
	}
	return false
}

// ResetRetriesOnWasTransferProposedOnKC -
func (stub *BridgeExecutorStub) ResetRetriesOnWasTransferProposedOnKC() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesOnWasTransferProposedOnKCCalled != nil {
		stub.ResetRetriesOnWasTransferProposedOnKCCalled()
	}
}

// WasSetStatusProposedOnKC -
func (stub *BridgeExecutorStub) WasSetStatusProposedOnKC(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasSetStatusProposedOnKCCalled != nil {
		return stub.WasSetStatusProposedOnKCCalled(ctx)
	}
	return false, notImplemented
}

// ProposeSetStatusOnKC -
func (stub *BridgeExecutorStub) ProposeSetStatusOnKC(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeSetStatusOnKCCalled != nil {
		return stub.ProposeSetStatusOnKCCalled(ctx)
	}
	return notImplemented
}

// WasActionSignedOnKC -
func (stub *BridgeExecutorStub) WasActionSignedOnKC(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionSignedOnKCCalled != nil {
		return stub.WasActionSignedOnKCCalled(ctx)
	}
	return false, notImplemented
}

// SignActionOnKC -
func (stub *BridgeExecutorStub) SignActionOnKC(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.SignActionOnKCCalled != nil {
		return stub.SignActionOnKCCalled(ctx)
	}
	return notImplemented
}

// ProcessQuorumReachedOnKC -
func (stub *BridgeExecutorStub) ProcessQuorumReachedOnKC(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.ProcessQuorumReachedOnKCCalled != nil {
		return stub.ProcessQuorumReachedOnKCCalled(ctx)
	}
	return false, notImplemented
}

// WasActionPerformedOnKC -
func (stub *BridgeExecutorStub) WasActionPerformedOnKC(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionPerformedOnKCCalled != nil {
		return stub.WasActionPerformedOnKCCalled(ctx)
	}
	return false, notImplemented
}

// PerformActionOnKC -
func (stub *BridgeExecutorStub) PerformActionOnKC(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformActionOnKCCalled != nil {
		return stub.PerformActionOnKCCalled(ctx)
	}
	return notImplemented
}

// ResolveNewDepositsStatuses -
func (stub *BridgeExecutorStub) ResolveNewDepositsStatuses(numDeposits uint64) {
	stub.incrementFunctionCounter()
	if stub.ResolveNewDepositsStatusesCalled != nil {
		stub.ResolveNewDepositsStatusesCalled(numDeposits)
	}
}

// ProcessMaxQuorumRetriesOnKC -
func (stub *BridgeExecutorStub) ProcessMaxQuorumRetriesOnKC() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxQuorumRetriesOnKCCalled != nil {
		return stub.ProcessMaxQuorumRetriesOnKCCalled()
	}
	return false
}

// ResetRetriesCountOnKC -
func (stub *BridgeExecutorStub) ResetRetriesCountOnKC() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnKCCalled != nil {
		stub.ResetRetriesCountOnKCCalled()
	}
}

// GetAndStoreBatchFromEthereum -
func (stub *BridgeExecutorStub) GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreBatchFromEthereumCalled != nil {
		return stub.GetAndStoreBatchFromEthereumCalled(ctx, nonce)
	}
	return notImplemented
}

// WasTransferPerformedOnEthereum -
func (stub *BridgeExecutorStub) WasTransferPerformedOnEthereum(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferPerformedOnEthereumCalled != nil {
		return stub.WasTransferPerformedOnEthereumCalled(ctx)
	}
	return false, notImplemented
}

// SignTransferOnEthereum -
func (stub *BridgeExecutorStub) SignTransferOnEthereum() error {
	stub.incrementFunctionCounter()
	if stub.SignTransferOnEthereumCalled != nil {
		return stub.SignTransferOnEthereumCalled()
	}
	return notImplemented
}

// PerformTransferOnEthereum -
func (stub *BridgeExecutorStub) PerformTransferOnEthereum(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformTransferOnEthereumCalled != nil {
		return stub.PerformTransferOnEthereumCalled(ctx)
	}
	return notImplemented
}

// ProcessQuorumReachedOnEthereum -
func (stub *BridgeExecutorStub) ProcessQuorumReachedOnEthereum(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.ProcessQuorumReachedOnEthereumCalled != nil {
		return stub.ProcessQuorumReachedOnEthereumCalled(ctx)
	}
	return false, notImplemented
}

// WaitForTransferConfirmation -
func (stub *BridgeExecutorStub) WaitForTransferConfirmation(ctx context.Context) {
	stub.incrementFunctionCounter()
	if stub.WaitForTransferConfirmationCalled != nil {
		stub.WaitForTransferConfirmationCalled(ctx)
	}
}

// WaitAndReturnFinalBatchStatuses -
func (stub *BridgeExecutorStub) WaitAndReturnFinalBatchStatuses(ctx context.Context) []byte {
	stub.incrementFunctionCounter()
	if stub.WaitAndReturnFinalBatchStatusesCalled != nil {
		return stub.WaitAndReturnFinalBatchStatusesCalled(ctx)
	}
	return nil
}

// GetBatchStatusesFromEthereum -
func (stub *BridgeExecutorStub) GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error) {
	stub.incrementFunctionCounter()
	if stub.GetBatchStatusesFromEthereumCalled != nil {
		return stub.GetBatchStatusesFromEthereumCalled(ctx)
	}
	return nil, notImplemented
}

// ProcessMaxQuorumRetriesOnEthereum -
func (stub *BridgeExecutorStub) ProcessMaxQuorumRetriesOnEthereum() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxQuorumRetriesOnEthereumCalled != nil {
		return stub.ProcessMaxQuorumRetriesOnEthereumCalled()
	}
	return false
}

// ResetRetriesCountOnEthereum -
func (stub *BridgeExecutorStub) ResetRetriesCountOnEthereum() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnEthereumCalled != nil {
		stub.ResetRetriesCountOnEthereumCalled()
	}
}

// ClearStoredP2PSignaturesForEthereum -
func (stub *BridgeExecutorStub) ClearStoredP2PSignaturesForEthereum() {
	stub.incrementFunctionCounter()
	if stub.ClearStoredP2PSignaturesForEthereumCalled != nil {
		stub.ClearStoredP2PSignaturesForEthereumCalled()
	}
}

// CheckKCClientAvailability -
func (stub *BridgeExecutorStub) CheckKCClientAvailability(ctx context.Context) error {
	if stub.CheckKCClientAvailabilityCalled != nil {
		return stub.CheckKCClientAvailabilityCalled(ctx)
	}
	return notImplemented
}

// CheckEthereumClientAvailability -
func (stub *BridgeExecutorStub) CheckEthereumClientAvailability(ctx context.Context) error {
	if stub.CheckEthereumClientAvailabilityCalled != nil {
		return stub.CheckEthereumClientAvailabilityCalled(ctx)
	}
	return notImplemented
}

// IsInterfaceNil -
func (stub *BridgeExecutorStub) IsInterfaceNil() bool {
	return stub == nil
}

// -------- helper functions

// incrementFunctionCounter increments the counter for the function that called it
func (stub *BridgeExecutorStub) incrementFunctionCounter() {
	stub.mutExecutor.Lock()
	defer stub.mutExecutor.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	fmt.Printf("BridgeExecutorMock: called %s\n", runtime.FuncForPC(pc).Name())
	stub.functionCalledCounter[strings.ReplaceAll(runtime.FuncForPC(pc).Name(), stub.fullPath, "")]++
}

// GetFunctionCounter returns the called counter of a given function
func (stub *BridgeExecutorStub) GetFunctionCounter(function string) int {
	stub.mutExecutor.Lock()
	defer stub.mutExecutor.Unlock()

	return stub.functionCalledCounter[function]
}

// CheckAvailableTokens -
func (stub *BridgeExecutorStub) CheckAvailableTokens(ctx context.Context, ethTokens []common.Address, kdaTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
	if stub.CheckAvailableTokensCalled != nil {
		return stub.CheckAvailableTokensCalled(ctx, ethTokens, kdaTokens, amounts, direction)
	}

	return nil
}
