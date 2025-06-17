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

	GetBatchFromKc(ctx context.Context) (*bridgeCore.TransferBatch, error)
	StoreBatchFromKc(batch *bridgeCore.TransferBatch) error
	GetStoredBatch() *bridgeCore.TransferBatch

	GetLastExecutedEthBatchIDFromKc(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error

	GetAndStoreActionIDForProposeTransferOnKc(ctx context.Context) (uint64, error)
	GetAndStoreActionIDForProposeSetStatusFromKc(ctx context.Context) (uint64, error)
	GetStoredActionID() uint64

	WasTransferProposedOnKc(ctx context.Context) (bool, error)
	ProposeTransferOnKc(ctx context.Context) error
	ProcessMaxRetriesOnWasTransferProposedOnKc() bool
	ResetRetriesOnWasTransferProposedOnKc()

	WasSetStatusProposedOnKc(ctx context.Context) (bool, error)
	ProposeSetStatusOnKc(ctx context.Context) error

	WasActionSignedOnKc(ctx context.Context) (bool, error)
	SignActionOnKc(ctx context.Context) error

	ProcessQuorumReachedOnKc(ctx context.Context) (bool, error)
	WasActionPerformedOnKc(ctx context.Context) (bool, error)
	PerformActionOnKc(ctx context.Context) error
	ResolveNewDepositsStatuses(numDeposits uint64)

	ProcessMaxQuorumRetriesOnKc() bool
	ResetRetriesCountOnKc()

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

	CheckKcClientAvailability(ctx context.Context) error
	CheckEthereumClientAvailability(ctx context.Context) error
	CheckAvailableTokens(ctx context.Context, ethTokens []common.Address, kdaTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error

	IsInterfaceNil() bool
}
