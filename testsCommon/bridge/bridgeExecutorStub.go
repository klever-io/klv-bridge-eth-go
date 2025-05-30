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

	PrintInfoCalled                                             func(logLevel logger.LogLevel, message string, extras ...interface{})
	MyTurnAsLeaderCalled                                        func() bool
	GetBatchFromKleverchainCalled                               func(ctx context.Context) (*bridgeCore.TransferBatch, error)
	StoreBatchFromKleverchainCalled                             func(batch *bridgeCore.TransferBatch) error
	GetStoredBatchCalled                                        func() *bridgeCore.TransferBatch
	GetLastExecutedEthBatchIDFromKleverchainCalled              func(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatchCalled         func(ctx context.Context) error
	GetAndStoreActionIDForProposeTransferOnKleverchainCalled    func(ctx context.Context) (uint64, error)
	GetAndStoreActionIDForProposeSetStatusFromKleverchainCalled func(ctx context.Context) (uint64, error)
	GetStoredActionIDCalled                                     func() uint64
	WasTransferProposedOnKleverchainCalled                      func(ctx context.Context) (bool, error)
	ProposeTransferOnKleverchainCalled                          func(ctx context.Context) error
	ProcessMaxRetriesOnWasTransferProposedOnKleverchainCalled   func() bool
	ResetRetriesOnWasTransferProposedOnKleverchainCalled        func()
	WasSetStatusProposedOnKleverchainCalled                     func(ctx context.Context) (bool, error)
	ProposeSetStatusOnKleverchainCalled                         func(ctx context.Context) error
	WasActionSignedOnKleverchainCalled                          func(ctx context.Context) (bool, error)
	SignActionOnKleverchainCalled                               func(ctx context.Context) error
	ProcessQuorumReachedOnKleverchainCalled                     func(ctx context.Context) (bool, error)
	WasActionPerformedOnKleverchainCalled                       func(ctx context.Context) (bool, error)
	PerformActionOnKleverchainCalled                            func(ctx context.Context) error
	ResolveNewDepositsStatusesCalled                            func(numDeposits uint64)
	ProcessMaxQuorumRetriesOnKleverchainCalled                  func() bool
	ResetRetriesCountOnKleverchainCalled                        func()
	GetAndStoreBatchFromEthereumCalled                          func(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnEthereumCalled                        func(ctx context.Context) (bool, error)
	SignTransferOnEthereumCalled                                func() error
	PerformTransferOnEthereumCalled                             func(ctx context.Context) error
	ProcessQuorumReachedOnEthereumCalled                        func(ctx context.Context) (bool, error)
	WaitForTransferConfirmationCalled                           func(ctx context.Context)
	WaitAndReturnFinalBatchStatusesCalled                       func(ctx context.Context) []byte
	GetBatchStatusesFromEthereumCalled                          func(ctx context.Context) ([]byte, error)
	ProcessMaxQuorumRetriesOnEthereumCalled                     func() bool
	ResetRetriesCountOnEthereumCalled                           func()
	ClearStoredP2PSignaturesForEthereumCalled                   func()
	CheckKleverchainClientAvailabilityCalled                    func(ctx context.Context) error
	CheckEthereumClientAvailabilityCalled                       func(ctx context.Context) error
	CheckAvailableTokensCalled                                  func(ctx context.Context, ethTokens []common.Address, kdaTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error
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

// GetBatchFromKleverchain -
func (stub *BridgeExecutorStub) GetBatchFromKleverchain(ctx context.Context) (*bridgeCore.TransferBatch, error) {
	stub.incrementFunctionCounter()
	if stub.GetBatchFromKleverchainCalled != nil {
		return stub.GetBatchFromKleverchainCalled(ctx)
	}
	return nil, notImplemented
}

// StoreBatchFromKleverchain -
func (stub *BridgeExecutorStub) StoreBatchFromKleverchain(batch *bridgeCore.TransferBatch) error {
	stub.incrementFunctionCounter()
	if stub.StoreBatchFromKleverchainCalled != nil {
		return stub.StoreBatchFromKleverchainCalled(batch)
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

// GetLastExecutedEthBatchIDFromKleverchain -
func (stub *BridgeExecutorStub) GetLastExecutedEthBatchIDFromKleverchain(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetLastExecutedEthBatchIDFromKleverchainCalled != nil {
		return stub.GetLastExecutedEthBatchIDFromKleverchainCalled(ctx)
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

// GetAndStoreActionIDForProposeTransferOnKleverchain -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeTransferOnKleverchain(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeTransferOnKleverchainCalled != nil {
		return stub.GetAndStoreActionIDForProposeTransferOnKleverchainCalled(ctx)
	}
	return 0, notImplemented
}

// GetAndStoreActionIDForProposeSetStatusFromKleverchain -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeSetStatusFromKleverchain(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeSetStatusFromKleverchainCalled != nil {
		return stub.GetAndStoreActionIDForProposeSetStatusFromKleverchainCalled(ctx)
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

// WasTransferProposedOnKleverchain -
func (stub *BridgeExecutorStub) WasTransferProposedOnKleverchain(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferProposedOnKleverchainCalled != nil {
		return stub.WasTransferProposedOnKleverchainCalled(ctx)
	}
	return false, notImplemented
}

// ProposeTransferOnKleverchain -
func (stub *BridgeExecutorStub) ProposeTransferOnKleverchain(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeTransferOnKleverchainCalled != nil {
		return stub.ProposeTransferOnKleverchainCalled(ctx)
	}
	return notImplemented
}

// ProcessMaxRetriesOnWasTransferProposedOnKleverchain -
func (stub *BridgeExecutorStub) ProcessMaxRetriesOnWasTransferProposedOnKleverchain() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxRetriesOnWasTransferProposedOnKleverchainCalled != nil {
		return stub.ProcessMaxRetriesOnWasTransferProposedOnKleverchainCalled()
	}
	return false
}

// ResetRetriesOnWasTransferProposedOnKleverchain -
func (stub *BridgeExecutorStub) ResetRetriesOnWasTransferProposedOnKleverchain() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesOnWasTransferProposedOnKleverchainCalled != nil {
		stub.ResetRetriesOnWasTransferProposedOnKleverchainCalled()
	}
}

// WasSetStatusProposedOnKleverchain -
func (stub *BridgeExecutorStub) WasSetStatusProposedOnKleverchain(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasSetStatusProposedOnKleverchainCalled != nil {
		return stub.WasSetStatusProposedOnKleverchainCalled(ctx)
	}
	return false, notImplemented
}

// ProposeSetStatusOnKleverchain -
func (stub *BridgeExecutorStub) ProposeSetStatusOnKleverchain(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeSetStatusOnKleverchainCalled != nil {
		return stub.ProposeSetStatusOnKleverchainCalled(ctx)
	}
	return notImplemented
}

// WasActionSignedOnKleverchain -
func (stub *BridgeExecutorStub) WasActionSignedOnKleverchain(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionSignedOnKleverchainCalled != nil {
		return stub.WasActionSignedOnKleverchainCalled(ctx)
	}
	return false, notImplemented
}

// SignActionOnKleverchain -
func (stub *BridgeExecutorStub) SignActionOnKleverchain(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.SignActionOnKleverchainCalled != nil {
		return stub.SignActionOnKleverchainCalled(ctx)
	}
	return notImplemented
}

// ProcessQuorumReachedOnKleverchain -
func (stub *BridgeExecutorStub) ProcessQuorumReachedOnKleverchain(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.ProcessQuorumReachedOnKleverchainCalled != nil {
		return stub.ProcessQuorumReachedOnKleverchainCalled(ctx)
	}
	return false, notImplemented
}

// WasActionPerformedOnKleverchain -
func (stub *BridgeExecutorStub) WasActionPerformedOnKleverchain(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionPerformedOnKleverchainCalled != nil {
		return stub.WasActionPerformedOnKleverchainCalled(ctx)
	}
	return false, notImplemented
}

// PerformActionOnKleverchain -
func (stub *BridgeExecutorStub) PerformActionOnKleverchain(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformActionOnKleverchainCalled != nil {
		return stub.PerformActionOnKleverchainCalled(ctx)
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

// ProcessMaxQuorumRetriesOnKleverchain -
func (stub *BridgeExecutorStub) ProcessMaxQuorumRetriesOnKleverchain() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxQuorumRetriesOnKleverchainCalled != nil {
		return stub.ProcessMaxQuorumRetriesOnKleverchainCalled()
	}
	return false
}

// ResetRetriesCountOnKleverchain -
func (stub *BridgeExecutorStub) ResetRetriesCountOnKleverchain() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnKleverchainCalled != nil {
		stub.ResetRetriesCountOnKleverchainCalled()
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

// CheckKleverchainClientAvailability -
func (stub *BridgeExecutorStub) CheckKleverchainClientAvailability(ctx context.Context) error {
	if stub.CheckKleverchainClientAvailabilityCalled != nil {
		return stub.CheckKleverchainClientAvailabilityCalled(ctx)
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
