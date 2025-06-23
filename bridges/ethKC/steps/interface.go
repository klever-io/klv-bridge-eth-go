package steps

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/batchProcessor"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// Executor defines a generic bridge interface able to handle both halves of the bridge
type Executor interface {
	PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{})
	MyTurnAsLeader() bool

	GetBatchFromKC(ctx context.Context) (*bridgeCore.TransferBatch, error)
	StoreBatchFromKC(batch *bridgeCore.TransferBatch) error
	GetStoredBatch() *bridgeCore.TransferBatch

	GetLastExecutedEthBatchIDFromKC(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error

	GetAndStoreActionIDForProposeTransferOnKC(ctx context.Context) (uint64, error)
	GetAndStoreActionIDForProposeSetStatusFromKC(ctx context.Context) (uint64, error)
	GetStoredActionID() uint64

	WasTransferProposedOnKC(ctx context.Context) (bool, error)
	ProposeTransferOnKC(ctx context.Context) error
	ProcessMaxRetriesOnWasTransferProposedOnKC() bool
	ResetRetriesOnWasTransferProposedOnKC()

	WasSetStatusProposedOnKC(ctx context.Context) (bool, error)
	ProposeSetStatusOnKC(ctx context.Context) error

	WasActionSignedOnKC(ctx context.Context) (bool, error)
	SignActionOnKC(ctx context.Context) error

	ProcessQuorumReachedOnKC(ctx context.Context) (bool, error)
	WasActionPerformedOnKC(ctx context.Context) (bool, error)
	PerformActionOnKC(ctx context.Context) error
	ResolveNewDepositsStatuses(numDeposits uint64)

	ProcessMaxQuorumRetriesOnKC() bool
	ResetRetriesCountOnKC()

	GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnEthereum(ctx context.Context) (bool, error)
	SignTransferOnEthereum() error
	PerformTransferOnEthereum(ctx context.Context) error
	ProcessQuorumReachedOnEthereum(ctx context.Context) (bool, error)
	WaitForTransferConfirmation(ctx context.Context)
	WaitAndReturnFinalBatchStatuses(ctx context.Context) []byte
	GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error)

	ProcessMaxQuorumRetriesOnEthereum() bool
	ResetRetriesCountOnEthereum()
	ClearStoredP2PSignaturesForEthereum()

	CheckKCClientAvailability(ctx context.Context) error
	CheckEthereumClientAvailability(ctx context.Context) error
	CheckAvailableTokens(ctx context.Context, ethTokens []common.Address, kdaTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error

	IsInterfaceNil() bool
}
