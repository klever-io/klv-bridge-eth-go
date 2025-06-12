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
	GetBatchFromKcCalled                                func(ctx context.Context) (*bridgeCore.TransferBatch, error)
	StoreBatchFromKcCalled                              func(batch *bridgeCore.TransferBatch) error
	GetStoredBatchCalled                                func() *bridgeCore.TransferBatch
	GetLastExecutedEthBatchIDFromKcCalled               func(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatchCalled func(ctx context.Context) error
	GetAndStoreActionIDForProposeTransferOnKcCalled     func(ctx context.Context) (uint64, error)
	GetAndStoreActionIDForProposeSetStatusFromKcCalled  func(ctx context.Context) (uint64, error)
	GetStoredActionIDCalled                             func() uint64
	WasTransferProposedOnKcCalled                       func(ctx context.Context) (bool, error)
	ProposeTransferOnKcCalled                           func(ctx context.Context) error
	ProcessMaxRetriesOnWasTransferProposedOnKcCalled    func() bool
	ResetRetriesOnWasTransferProposedOnKcCalled         func()
	WasSetStatusProposedOnKcCalled                      func(ctx context.Context) (bool, error)
	ProposeSetStatusOnKcCalled                          func(ctx context.Context) error
	WasActionSignedOnKcCalled                           func(ctx context.Context) (bool, error)
	SignActionOnKcCalled                                func(ctx context.Context) error
	ProcessQuorumReachedOnKcCalled                      func(ctx context.Context) (bool, error)
	WasActionPerformedOnKcCalled                        func(ctx context.Context) (bool, error)
	PerformActionOnKcCalled                             func(ctx context.Context) error
	ResolveNewDepositsStatusesCalled                    func(numDeposits uint64)
	ProcessMaxQuorumRetriesOnKcCalled                   func() bool
	ResetRetriesCountOnKcCalled                         func()
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
	CheckKcClientAvailabilityCalled                     func(ctx context.Context) error
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

// GetBatchFromKc -
func (stub *BridgeExecutorStub) GetBatchFromKc(ctx context.Context) (*bridgeCore.TransferBatch, error) {
	stub.incrementFunctionCounter()
	if stub.GetBatchFromKcCalled != nil {
		return stub.GetBatchFromKcCalled(ctx)
	}
	return nil, notImplemented
}

// StoreBatchFromKc -
func (stub *BridgeExecutorStub) StoreBatchFromKc(batch *bridgeCore.TransferBatch) error {
	stub.incrementFunctionCounter()
	if stub.StoreBatchFromKcCalled != nil {
		return stub.StoreBatchFromKcCalled(batch)
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

// GetLastExecutedEthBatchIDFromKc -
func (stub *BridgeExecutorStub) GetLastExecutedEthBatchIDFromKc(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetLastExecutedEthBatchIDFromKcCalled != nil {
		return stub.GetLastExecutedEthBatchIDFromKcCalled(ctx)
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

// GetAndStoreActionIDForProposeTransferOnKc -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeTransferOnKc(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeTransferOnKcCalled != nil {
		return stub.GetAndStoreActionIDForProposeTransferOnKcCalled(ctx)
	}
	return 0, notImplemented
}

// GetAndStoreActionIDForProposeSetStatusFromKc -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeSetStatusFromKc(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeSetStatusFromKcCalled != nil {
		return stub.GetAndStoreActionIDForProposeSetStatusFromKcCalled(ctx)
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

// WasTransferProposedOnKc -
func (stub *BridgeExecutorStub) WasTransferProposedOnKc(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferProposedOnKcCalled != nil {
		return stub.WasTransferProposedOnKcCalled(ctx)
	}
	return false, notImplemented
}

// ProposeTransferOnKc -
func (stub *BridgeExecutorStub) ProposeTransferOnKc(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeTransferOnKcCalled != nil {
		return stub.ProposeTransferOnKcCalled(ctx)
	}
	return notImplemented
}

// ProcessMaxRetriesOnWasTransferProposedOnKc -
func (stub *BridgeExecutorStub) ProcessMaxRetriesOnWasTransferProposedOnKc() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxRetriesOnWasTransferProposedOnKcCalled != nil {
		return stub.ProcessMaxRetriesOnWasTransferProposedOnKcCalled()
	}
	return false
}

// ResetRetriesOnWasTransferProposedOnKc -
func (stub *BridgeExecutorStub) ResetRetriesOnWasTransferProposedOnKc() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesOnWasTransferProposedOnKcCalled != nil {
		stub.ResetRetriesOnWasTransferProposedOnKcCalled()
	}
}

// WasSetStatusProposedOnKc -
func (stub *BridgeExecutorStub) WasSetStatusProposedOnKc(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasSetStatusProposedOnKcCalled != nil {
		return stub.WasSetStatusProposedOnKcCalled(ctx)
	}
	return false, notImplemented
}

// ProposeSetStatusOnKc -
func (stub *BridgeExecutorStub) ProposeSetStatusOnKc(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeSetStatusOnKcCalled != nil {
		return stub.ProposeSetStatusOnKcCalled(ctx)
	}
	return notImplemented
}

// WasActionSignedOnKc -
func (stub *BridgeExecutorStub) WasActionSignedOnKc(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionSignedOnKcCalled != nil {
		return stub.WasActionSignedOnKcCalled(ctx)
	}
	return false, notImplemented
}

// SignActionOnKc -
func (stub *BridgeExecutorStub) SignActionOnKc(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.SignActionOnKcCalled != nil {
		return stub.SignActionOnKcCalled(ctx)
	}
	return notImplemented
}

// ProcessQuorumReachedOnKc -
func (stub *BridgeExecutorStub) ProcessQuorumReachedOnKc(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.ProcessQuorumReachedOnKcCalled != nil {
		return stub.ProcessQuorumReachedOnKcCalled(ctx)
	}
	return false, notImplemented
}

// WasActionPerformedOnKc -
func (stub *BridgeExecutorStub) WasActionPerformedOnKc(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionPerformedOnKcCalled != nil {
		return stub.WasActionPerformedOnKcCalled(ctx)
	}
	return false, notImplemented
}

// PerformActionOnKc -
func (stub *BridgeExecutorStub) PerformActionOnKc(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformActionOnKcCalled != nil {
		return stub.PerformActionOnKcCalled(ctx)
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

// ProcessMaxQuorumRetriesOnKc -
func (stub *BridgeExecutorStub) ProcessMaxQuorumRetriesOnKc() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxQuorumRetriesOnKcCalled != nil {
		return stub.ProcessMaxQuorumRetriesOnKcCalled()
	}
	return false
}

// ResetRetriesCountOnKc -
func (stub *BridgeExecutorStub) ResetRetriesCountOnKc() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnKcCalled != nil {
		stub.ResetRetriesCountOnKcCalled()
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

// CheckKcClientAvailability -
func (stub *BridgeExecutorStub) CheckKcClientAvailability(ctx context.Context) error {
	if stub.CheckKcClientAvailabilityCalled != nil {
		return stub.CheckKcClientAvailabilityCalled(ctx)
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
