package ethKC

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/klever-io/klv-bridge-eth-go/clients/ethereum/contract"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/batchProcessor"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")
var providedBatch = &bridgeCore.TransferBatch{}
var expectedMaxRetries = uint64(3)

func createMockExecutorArgs() ArgsBridgeExecutor {
	return ArgsBridgeExecutor{
		Log:                        logger.GetOrCreate("test"),
		KCClient:                   &bridgeTests.KCClientStub{},
		EthereumClient:             &bridgeTests.EthereumClientStub{},
		TopologyProvider:           &bridgeTests.TopologyProviderStub{},
		StatusHandler:              testsCommon.NewStatusHandlerMock("test"),
		TimeForWaitOnEthereum:      time.Second,
		SignaturesHolder:           &testsCommon.SignaturesHolderStub{},
		BalanceValidator:           &testsCommon.BalanceValidatorStub{},
		MaxQuorumRetriesOnEthereum: minRetries,
		MaxQuorumRetriesOnKC:       minRetries,
		MaxRetriesOnWasProposed:    minRetries,
	}
}

func TestNewBridgeExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.Log = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("nil klever blockchain client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.KCClient = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilKCClient, err)
	})
	t.Run("nil ethereum client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilEthereumClient, err)
	})
	t.Run("nil topology provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TopologyProvider = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilTopologyProvider, err)
	})
	t.Run("nil status handler", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.StatusHandler = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilStatusHandler, err)
	})
	t.Run("invalid time", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 0
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrInvalidDuration, err)
	})
	t.Run("nil signatures holder", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.SignaturesHolder = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilSignaturesHolder, err)
	})
	t.Run("nil balance validator", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.BalanceValidator = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilBalanceValidator, err)
	})
	t.Run("invalid MaxQuorumRetriesOnEthereum value", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MaxQuorumRetriesOnEthereum = 0
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for args.MaxQuorumRetriesOnEthereum"))
	})
	t.Run("invalid MaxQuorumRetriesOnKC value", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MaxQuorumRetriesOnKC = 0
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for args.MaxQuorumRetriesOnKC"))
	})
	t.Run("invalid MaxRetriesOnWasProposed value", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MaxRetriesOnWasProposed = 0
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for args.MaxRetriesOnWasProposed"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, err := NewBridgeExecutor(args)

		assert.False(t, check.IfNil(executor))
		assert.Nil(t, err)
	})
}

func TestEthToKCBridgeExecutor_PrintInfo(t *testing.T) {
	t.Parallel()

	logLevels := []logger.LogLevel{logger.LogTrace, logger.LogDebug, logger.LogInfo, logger.LogWarning, logger.LogError, logger.LogNone}
	for _, logLevel := range logLevels {
		shouldOutputToStatusHandler := logLevel == logger.LogError || logLevel == logger.LogWarning
		testPrintInfo(t, logLevel, shouldOutputToStatusHandler)
	}
}

func testPrintInfo(t *testing.T, logLevel logger.LogLevel, shouldOutputToStatusHandler bool) {
	providedLogLevel := logLevel
	providedMessage := "message"
	providedArgs := []interface{}{"string", 1, []byte("aaa")}
	wasCalled := false

	args := createMockExecutorArgs()
	statusHandler := testsCommon.NewStatusHandlerMock("test")
	args.StatusHandler = statusHandler
	args.Log = &testsCommon.LoggerStub{
		LogCalled: func(logLevel logger.LogLevel, message string, args ...interface{}) {
			wasCalled = true
			assert.Equal(t, providedLogLevel, logLevel)
			assert.Equal(t, providedMessage, message)
			assert.Equal(t, providedArgs, args)
		},
	}
	executor, _ := NewBridgeExecutor(args)
	executor.PrintInfo(providedLogLevel, providedMessage, providedArgs...)

	assert.True(t, wasCalled)

	if shouldOutputToStatusHandler {
		assert.True(t, len(statusHandler.GetStringMetric(bridgeCore.MetricLastError)) > 0)
	}
}

func TestEthToKCBridgeExecutor_MyTurnAsLeader(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	wasCalled := false
	args.TopologyProvider = &bridgeTests.TopologyProviderStub{
		MyTurnAsLeaderCalled: func() bool {
			wasCalled = true
			return true
		},
	}

	executor, _ := NewBridgeExecutor(args)
	assert.True(t, executor.MyTurnAsLeader())
	assert.True(t, wasCalled)
}

func TestEthToKCBridgeExecutor_GetAndStoreActionIDForProposeTransferOnKC(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		actionID, err := executor.GetAndStoreActionIDForProposeTransferOnKC(context.Background())
		assert.Zero(t, actionID)
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("klever blockchain client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			GetActionIDForProposeTransferCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (uint64, error) {
				assert.True(t, providedBatch == batch)
				return 0, expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		actionID, err := executor.GetAndStoreActionIDForProposeTransferOnKC(context.Background())
		assert.Zero(t, actionID)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(48939)

		args.KCClient = &bridgeTests.KCClientStub{
			GetActionIDForProposeTransferCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (uint64, error) {
				assert.True(t, providedBatch == batch)
				return providedActionID, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		assert.NotEqual(t, providedActionID, executor.actionID)

		actionID, err := executor.GetAndStoreActionIDForProposeTransferOnKC(context.Background())
		assert.Equal(t, providedActionID, actionID)
		assert.Nil(t, err)
		assert.Equal(t, providedActionID, executor.GetStoredActionID())
		assert.Equal(t, providedActionID, executor.actionID)
	})
}

func TestEthToKCBridgeExecutor_GetAndStoreBatchFromEthereum(t *testing.T) {
	t.Parallel()

	t.Run("ethereum client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
				assert.Equal(t, providedNonce, nonce)
				return nil, false, expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Equal(t, expectedErr, err)
	})
	t.Run("batch nonce mismatch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		expectedBatch := &bridgeCore.TransferBatch{
			ID: 0,
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, true, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.True(t, errors.Is(err, ErrFinalBatchNotFound))
		assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", providedNonce)))
		assert.Nil(t, executor.GetStoredBatch())
		assert.Nil(t, executor.batch)
	})
	t.Run("no deposits should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		expectedBatch := &bridgeCore.TransferBatch{
			ID: providedNonce,
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, true, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.True(t, errors.Is(err, ErrFinalBatchNotFound))
		assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", providedNonce)))
		assert.Nil(t, executor.GetStoredBatch())
		assert.Nil(t, executor.batch)
	})
	t.Run("not a final batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		expectedBatch := &bridgeCore.TransferBatch{
			ID: providedNonce,
			Deposits: []*bridgeCore.DepositTransfer{
				{},
			},
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, false, nil
			},
			GetBatchSCMetadataCalled: func(ctx context.Context, nonce uint64, blockNumber int64) ([]*contract.ERC20SafeERC20SCDeposit, error) {
				return make([]*contract.ERC20SafeERC20SCDeposit, 0), nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.True(t, errors.Is(err, ErrFinalBatchNotFound))
		assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", providedNonce)))
		assert.Nil(t, executor.GetStoredBatch())
		assert.Nil(t, executor.batch)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		expectedBatch := &bridgeCore.TransferBatch{
			ID: providedNonce,
			Deposits: []*bridgeCore.DepositTransfer{
				{},
			},
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, true, nil
			},
			GetBatchSCMetadataCalled: func(ctx context.Context, nonce uint64, blockNumber int64) ([]*contract.ERC20SafeERC20SCDeposit, error) {
				return make([]*contract.ERC20SafeERC20SCDeposit, 0), nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Nil(t, err)
		assert.True(t, expectedBatch == executor.GetStoredBatch()) // pointer testing
		assert.True(t, expectedBatch == executor.batch)
	})
	t.Run("should add deposits metadata for sc calls", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		depositNonce := uint64(100)
		depositData := []byte("testData")
		expectedBatch := &bridgeCore.TransferBatch{
			ID: providedNonce,
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce: depositNonce,
				},
			},
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, true, nil
			},
			GetBatchSCMetadataCalled: func(ctx context.Context, nonce uint64, blockNumber int64) ([]*contract.ERC20SafeERC20SCDeposit, error) {
				return []*contract.ERC20SafeERC20SCDeposit{{
					DepositNonce: big.NewInt(0).SetUint64(depositNonce),
					CallData:     depositData,
				}}, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Nil(t, err)
		assert.True(t, expectedBatch == executor.GetStoredBatch()) // pointer testing
		expectedDepositData := []byte{bridgeCore.DataPresentProtocolMarker, 0, 0, 0, byte(len(depositData))}
		expectedDepositData = append(expectedDepositData, []byte(depositData)...)
		assert.Equal(t, string(expectedDepositData), string(executor.batch.Deposits[0].Data))
	})
	t.Run("should add deposits metadata for sc calls with a data starting with missing data marker", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		depositNonce := uint64(100)
		depositData := append([]byte{bridgeCore.MissingDataProtocolMarker}, "testData"...)
		expectedBatch := &bridgeCore.TransferBatch{
			ID: providedNonce,
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce: depositNonce,
				},
			},
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, true, nil
			},
			GetBatchSCMetadataCalled: func(ctx context.Context, nonce uint64, blockNumber int64) ([]*contract.ERC20SafeERC20SCDeposit, error) {
				return []*contract.ERC20SafeERC20SCDeposit{{
					DepositNonce: big.NewInt(0).SetUint64(depositNonce),
					CallData:     depositData,
				}}, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Nil(t, err)
		assert.True(t, expectedBatch == executor.GetStoredBatch()) // pointer testing
		expectedDepositData := []byte{bridgeCore.DataPresentProtocolMarker, 0, 0, 0, byte(len(depositData))}
		expectedDepositData = append(expectedDepositData, []byte(depositData)...)
		assert.Equal(t, string(expectedDepositData), string(executor.batch.Deposits[0].Data))
	})
	t.Run("should add deposits metadata for sc calls even if with no data", func(t *testing.T) {
		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		depositNonce := uint64(100)
		depositData := make([]byte, 0)
		expectedBatch := &bridgeCore.TransferBatch{
			ID: providedNonce,
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce: depositNonce,
				},
			},
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, true, nil
			},
			GetBatchSCMetadataCalled: func(ctx context.Context, nonce uint64, blockNumber int64) ([]*contract.ERC20SafeERC20SCDeposit, error) {
				return []*contract.ERC20SafeERC20SCDeposit{{
					DepositNonce: big.NewInt(0).SetUint64(depositNonce),
					CallData:     depositData,
				}}, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Nil(t, err)
		assert.True(t, expectedBatch == executor.GetStoredBatch()) // pointer testing
		assert.Equal(t, string([]byte{bridgeCore.MissingDataProtocolMarker}), string(executor.batch.Deposits[0].Data))
	})
	t.Run("should bypass data if the data is the missing marker", func(t *testing.T) {
		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		depositNonce := uint64(100)
		depositData := []byte{bridgeCore.MissingDataProtocolMarker}
		expectedBatch := &bridgeCore.TransferBatch{
			ID: providedNonce,
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce: depositNonce,
				},
			},
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, true, nil
			},
			GetBatchSCMetadataCalled: func(ctx context.Context, nonce uint64, blockNumber int64) ([]*contract.ERC20SafeERC20SCDeposit, error) {
				return []*contract.ERC20SafeERC20SCDeposit{{
					DepositNonce: big.NewInt(0).SetUint64(depositNonce),
					CallData:     depositData,
				}}, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Nil(t, err)
		assert.True(t, expectedBatch == executor.GetStoredBatch()) // pointer testing
		assert.Equal(t, depositData, executor.batch.Deposits[0].Data)
	})
}

func TestEthToKCBridgeExecutor_GetLastExecutedEthBatchIDFromKC(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedBatchID := uint64(36727)
	args.KCClient = &bridgeTests.KCClientStub{
		GetLastExecutedEthBatchIDCalled: func(ctx context.Context) (uint64, error) {
			return providedBatchID, nil
		},
	}
	setIntCalled := false
	args.StatusHandler = &testsCommon.StatusHandlerStub{
		SetIntMetricCalled: func(metric string, value int) {
			assert.Equal(t, bridgeCore.MetricNumBatches, metric)
			assert.Equal(t, int(providedBatchID), value)
			setIntCalled = true
		},
	}
	executor, _ := NewBridgeExecutor(args)

	batchID, err := executor.GetLastExecutedEthBatchIDFromKC(context.Background())
	assert.Equal(t, providedBatchID, batchID)
	assert.Nil(t, err)
	assert.True(t, setIntCalled)
}

func TestEthToKCBridgeExecutor_VerifyLastDepositNonceExecutedOnEthereumBatch(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("get last executed tx id errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			GetLastExecutedEthTxIDCalled: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &bridgeCore.TransferBatch{}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Equal(t, expectedErr, err)
	})

	args := createMockExecutorArgs()
	txId := uint64(6657)
	args.KCClient = &bridgeTests.KCClientStub{
		GetLastExecutedEthTxIDCalled: func(ctx context.Context) (uint64, error) {
			return txId, nil
		},
	}

	t.Run("first deposit nonce equals last tx nonce should error", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewBridgeExecutor(args)
		executor.batch = &bridgeCore.TransferBatch{
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce: txId,
				},
			},
		}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.True(t, errors.Is(err, ErrInvalidDepositNonce))
		assert.True(t, strings.Contains(err.Error(), "6657"))
	})
	t.Run("first deposit nonce is smaller than the last tx nonce should error", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewBridgeExecutor(args)
		executor.batch = &bridgeCore.TransferBatch{
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce: txId - 1,
				},
			},
		}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.True(t, errors.Is(err, ErrInvalidDepositNonce))
		assert.True(t, strings.Contains(err.Error(), "6656"))
	})
	t.Run("gap found error", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewBridgeExecutor(args)
		executor.batch = &bridgeCore.TransferBatch{
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce: txId + 1,
				},
				{
					Nonce: txId + 3,
				},
			},
		}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.True(t, errors.Is(err, ErrInvalidDepositNonce))
		assert.True(t, strings.Contains(err.Error(), "6660"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewBridgeExecutor(args)
		executor.batch = &bridgeCore.TransferBatch{
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce: txId + 1,
				},
			},
		}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Nil(t, err)

		executor.batch = &bridgeCore.TransferBatch{
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce: txId + 1,
				},
				{
					Nonce: txId + 2,
				},
			},
		}

		err = executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Nil(t, err)
	})
}

func TestEthToKCBridgeExecutor_WasTransferProposedOnKC(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		wasTransfered, err := executor.WasTransferProposedOnKC(context.Background())
		assert.False(t, wasTransfered)
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		args.KCClient = &bridgeTests.KCClientStub{
			WasProposedTransferCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (bool, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		wasProposed, err := executor.WasTransferProposedOnKC(context.Background())
		assert.True(t, wasProposed)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToKCBridgeExecutor_ProposeTransferOnKC(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.ProposeTransferOnKC(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("propose transfer fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			ProposeTransferCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)

				return "", expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeTransferOnKC(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		args.KCClient = &bridgeTests.KCClientStub{
			ProposeTransferCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true

				return "", nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeTransferOnKC(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToKCBridgeExecutor_WasActionSignedOnKC(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.KCClient = &bridgeTests.KCClientStub{
		WasSignedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	executor.actionID = providedActionID

	wasSigned, err := executor.WasActionSignedOnKC(context.Background())
	assert.True(t, wasSigned)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToKCBridgeExecutor_SignActionOnKC(t *testing.T) {
	t.Parallel()

	t.Run("klever blockchain client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(378276)
		args.KCClient = &bridgeTests.KCClientStub{
			SignCalled: func(ctx context.Context, actionID uint64) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				return "", expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.actionID = providedActionID

		err := executor.SignActionOnKC(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(378276)
		wasCalled := false
		args.KCClient = &bridgeTests.KCClientStub{
			SignCalled: func(ctx context.Context, actionID uint64) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				wasCalled = true
				return "", nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.actionID = providedActionID

		err := executor.SignActionOnKC(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToKCBridgeExecutor_IsQuorumReachedOnKC(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.KCClient = &bridgeTests.KCClientStub{
		QuorumReachedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	executor.actionID = providedActionID

	isQuorumReached, err := executor.ProcessQuorumReachedOnKC(context.Background())
	assert.True(t, isQuorumReached)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToKCBridgeExecutor_WasActionPerformedOnKC(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.KCClient = &bridgeTests.KCClientStub{
		WasExecutedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	executor.actionID = providedActionID

	wasPerformed, err := executor.WasActionPerformedOnKC(context.Background())
	assert.True(t, wasPerformed)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToKCBridgeExecutor_PerformActionOnKC(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.PerformActionOnKC(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("klever blockchain client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(7383)
		args.KCClient = &bridgeTests.KCClientStub{
			PerformActionCalled: func(ctx context.Context, actionID uint64, batch *bridgeCore.TransferBatch) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				assert.True(t, providedBatch == batch)
				return "", expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		executor.actionID = providedActionID

		err := executor.PerformActionOnKC(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		providedActionID := uint64(7383)
		args.KCClient = &bridgeTests.KCClientStub{
			PerformActionCalled: func(ctx context.Context, actionID uint64, batch *bridgeCore.TransferBatch) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return "", nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		executor.actionID = providedActionID

		err := executor.PerformActionOnKC(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToKCBridgeExecutor_RetriesCountOnKC(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	args.MaxQuorumRetriesOnKC = expectedMaxRetries
	executor, _ := NewBridgeExecutor(args)
	for i := uint64(0); i < expectedMaxRetries; i++ {
		assert.False(t, executor.ProcessMaxQuorumRetriesOnKC())
	}

	assert.Equal(t, expectedMaxRetries, executor.quorumRetriesOnKC)
	assert.True(t, executor.ProcessMaxQuorumRetriesOnKC())
	executor.ResetRetriesCountOnKC()
	assert.Equal(t, uint64(0), executor.quorumRetriesOnKC)
}

func TestEthToKCBridgeExecutor_RetriesCountOnWasTransferProposedOnKC(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	args.MaxRetriesOnWasProposed = expectedMaxRetries
	executor, _ := NewBridgeExecutor(args)
	for i := uint64(0); i < expectedMaxRetries; i++ {
		assert.False(t, executor.ProcessMaxRetriesOnWasTransferProposedOnKC())
	}

	assert.Equal(t, expectedMaxRetries, executor.retriesOnWasProposed)
	assert.True(t, executor.ProcessMaxRetriesOnWasTransferProposedOnKC())
	executor.ResetRetriesOnWasTransferProposedOnKC()
	assert.Equal(t, uint64(0), executor.retriesOnWasProposed)
}

func TestKCToEthBridgeExecutor_GetAndStoreBatchFromKC(t *testing.T) {
	t.Parallel()

	t.Run("GetBatchFromKC fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			GetPendingBatchCalled: func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
				return nil, expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		_, err := executor.GetBatchFromKC(context.Background())
		assert.Equal(t, expectedErr, err)

		batch := executor.GetStoredBatch()
		assert.Nil(t, batch)
	})
	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{}

		executor, _ := NewBridgeExecutor(args)
		err := executor.StoreBatchFromKC(nil)
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			GetPendingBatchCalled: func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
				wasCalled = true
				return providedBatch, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		batch, err := executor.GetBatchFromKC(context.Background())
		assert.True(t, wasCalled)
		assert.Equal(t, providedBatch, batch)
		assert.Nil(t, err)

		err = executor.StoreBatchFromKC(batch)
		assert.Equal(t, providedBatch, executor.batch)
		assert.Nil(t, err)
	})
}

func TestKCToEthBridgeExecutor_GetAndStoreActionIDForProposeSetStatusFromKC(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		actionId, err := executor.GetAndStoreActionIDForProposeSetStatusFromKC(context.Background())
		assert.Equal(t, ErrNilBatch, err)
		assert.Equal(t, InvalidActionID, actionId)
	})
	t.Run("GetAndStoreActionIDForProposeSetStatusFromKC fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			GetActionIDForSetStatusOnPendingTransferCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (uint64, error) {
				return uint64(0), expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.GetAndStoreActionIDForProposeSetStatusFromKC(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedActionId := uint64(1123)
		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			GetActionIDForSetStatusOnPendingTransferCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (uint64, error) {
				wasCalled = true
				return providedActionId, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		actionId, err := executor.GetAndStoreActionIDForProposeSetStatusFromKC(context.Background())
		assert.True(t, wasCalled)
		assert.Equal(t, providedActionId, actionId)
		assert.Nil(t, err)

		actionId = executor.GetStoredActionID()
		assert.Equal(t, providedActionId, actionId)
	})
}

func TestKCToEthBridgeExecutor_WasSetStatusProposedOnKC(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		wasProposed, err := executor.WasSetStatusProposedOnKC(context.Background())
		assert.Equal(t, ErrNilBatch, err)
		assert.False(t, wasProposed)
	})
	t.Run("WasSetStatusProposedOnKC fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			WasProposedSetStatusCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (bool, error) {
				return false, expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.WasSetStatusProposedOnKC(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			WasProposedSetStatusCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (bool, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		wasProposed, err := executor.WasSetStatusProposedOnKC(context.Background())
		assert.True(t, wasCalled)
		assert.True(t, wasProposed)
		assert.Nil(t, err)
	})
}

func TestEthToKCBridgeExecutor_ProposeSetStatusOnKC(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.ProposeSetStatusOnKC(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("ProposeSetStatusOnKC fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			ProposeSetStatusCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (string, error) {
				return "", expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.ProposeSetStatusOnKC(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockExecutorArgs()
		args.KCClient = &bridgeTests.KCClientStub{
			ProposeSetStatusCalled: func(ctx context.Context, batch *bridgeCore.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true

				return "", nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeSetStatusOnKC(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestKCToEthBridgeExecutor_MyTurnAsLeader(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	wasCalled := false
	args.TopologyProvider = &bridgeTests.TopologyProviderStub{
		MyTurnAsLeaderCalled: func() bool {
			wasCalled = true
			return true
		},
	}

	executor, _ := NewBridgeExecutor(args)
	assert.True(t, executor.MyTurnAsLeader())
	assert.True(t, wasCalled)
}

func TestKCToEthBridgeExecutor_WasTransferPerformedOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		_, err := executor.WasTransferPerformedOnEthereum(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("WasExecuted fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
				return false, expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.WasTransferPerformedOnEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedBatchID := uint64(36727)
		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
				assert.True(t, providedBatchID == batchID)
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		executor.batch.ID = providedBatchID

		_, err := executor.WasTransferPerformedOnEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestKCToEthBridgeExecutor_SignTransferOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.SignTransferOnEthereum()
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("GenerateMessageHash fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GenerateMessageHashCalled: func(batch *batchProcessor.ArgListsBatch, batchID uint64) (common.Hash, error) {
				return common.Hash{}, expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.SignTransferOnEthereum()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalledGenerateMessageHashCalled := false
		wasCalledBroadcastSignatureForMessageHashCalled := false
		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GenerateMessageHashCalled: func(batch *batchProcessor.ArgListsBatch, batchID uint64) (common.Hash, error) {
				wasCalledGenerateMessageHashCalled = true
				return common.Hash{}, nil
			},
			BroadcastSignatureForMessageHashCalled: func(msgHash common.Hash) {
				wasCalledBroadcastSignatureForMessageHashCalled = true
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.SignTransferOnEthereum()
		assert.Nil(t, err)
		assert.True(t, wasCalledGenerateMessageHashCalled)
		assert.True(t, wasCalledBroadcastSignatureForMessageHashCalled)
	})
}

func TestKCToEthBridgeExecutor_PerformTransferOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.PerformTransferOnEthereum(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("GetQuorumSize fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetQuorumSizeCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(0), expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.PerformTransferOnEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ExecuteTransfer fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetQuorumSizeCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(0), nil
			},
			ExecuteTransferCalled: func(ctx context.Context, msgHash common.Hash, batch *batchProcessor.ArgListsBatch, batchId uint64, quorum int) (string, error) {
				return "", expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.PerformTransferOnEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedHash := common.Hash{}
		providedQuorum := 12
		wasCalledGetQuorumSizeCalled := false
		wasCalledExecuteTransferCalled := false
		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetQuorumSizeCalled: func(ctx context.Context) (*big.Int, error) {
				wasCalledGetQuorumSizeCalled = true
				return big.NewInt(int64(providedQuorum)), nil
			},
			ExecuteTransferCalled: func(ctx context.Context, msgHash common.Hash, batch *batchProcessor.ArgListsBatch, batchId uint64, quorum int) (string, error) {
				assert.True(t, providedHash == msgHash)
				assert.True(t, providedBatch.ID == batchId)
				for i := 0; i < len(providedBatch.Deposits); i++ {
					assert.Equal(t, providedBatch.Deposits[i].Amount, batch.Amounts[i])
					assert.Equal(t, providedBatch.Deposits[i].Nonce, batch.Nonces[i].Uint64())
					assert.Equal(t, providedBatch.Deposits[i].ToBytes, batch.Recipients[i].Bytes())
					assert.Equal(t, providedBatch.Deposits[i].SourceTokenBytes, batch.EthTokens[i].Bytes())
					assert.Equal(t, providedBatch.Deposits[i].DestinationTokenBytes, batch.KdaTokenBytes[i])
				}
				assert.True(t, providedQuorum == quorum)

				wasCalledExecuteTransferCalled = true
				return "", nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.msgHash = providedHash
		executor.batch = providedBatch
		err := executor.PerformTransferOnEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalledGetQuorumSizeCalled)
		assert.True(t, wasCalledExecuteTransferCalled)
	})
}

func TestKCToEthBridgeExecutor_IsQuorumReachedOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("ProcessQuorumReachedOnEthereum fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			IsQuorumReachedCalled: func(ctx context.Context, msgHash common.Hash) (bool, error) {
				return false, expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)

		_, err := executor.ProcessQuorumReachedOnEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			IsQuorumReachedCalled: func(ctx context.Context, msgHash common.Hash) (bool, error) {
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)

		isReached, err := executor.ProcessQuorumReachedOnEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
		assert.True(t, isReached)
	})
}

func TestKCToEthBridgeExecutor_RetriesCountOnEthereum(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	args.MaxQuorumRetriesOnEthereum = expectedMaxRetries
	executor, _ := NewBridgeExecutor(args)
	for i := uint64(0); i < expectedMaxRetries; i++ {
		assert.False(t, executor.ProcessMaxQuorumRetriesOnEthereum())
	}

	assert.Equal(t, expectedMaxRetries, executor.quorumRetriesOnEthereum)
	assert.True(t, executor.ProcessMaxQuorumRetriesOnEthereum())
	executor.ResetRetriesCountOnEthereum()
	assert.Equal(t, uint64(0), executor.quorumRetriesOnEthereum)
}

func TestWaitForTransferConfirmation(t *testing.T) {
	t.Parallel()

	t.Run("normal expiration", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 2 * time.Second
		executor, _ := NewBridgeExecutor(args)

		start := time.Now()
		executor.WaitForTransferConfirmation(context.Background())
		elapsed := time.Since(start)

		assert.True(t, elapsed >= args.TimeForWaitOnEthereum)
	})
	t.Run("context expiration", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		executor, _ := NewBridgeExecutor(args)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		start := time.Now()
		executor.WaitForTransferConfirmation(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForWaitOnEthereum)
	})

	t.Run("WasTransferPerformedOnEthereum always returns false/err", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		counter := 0
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
				counter++
				return false, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &bridgeCore.TransferBatch{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		executor.WaitForTransferConfirmation(ctx)

		assert.Equal(t, 10, counter)
	})

	t.Run("WasTransferPerformedOnEthereum always returns true only after 4 checks", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		counter := 0
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
				counter++
				if counter >= 5 {
					return true, nil
				}
				return false, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &bridgeCore.TransferBatch{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		start := time.Now()

		executor.WaitForTransferConfirmation(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForWaitOnEthereum)
		assert.Equal(t, 5, counter)
	})
}

func TestGetBatchStatusesFromEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)
		_, err := executor.GetBatchStatusesFromEthereum(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("GetTransactionsStatuses fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				return nil, expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.GetBatchStatusesFromEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedStatuses := []byte{bridgeCore.Executed, bridgeCore.Rejected}
		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				wasCalled = true
				return providedStatuses, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		statuses, err := executor.GetBatchStatusesFromEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
		assert.Equal(t, providedStatuses, statuses)
	})
}

func TestWaitAndReturnFinalBatchStatuses(t *testing.T) {
	t.Parallel()

	t.Run("normal expiration", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 2 * time.Second
		executor, _ := NewBridgeExecutor(args)

		start := time.Now()
		statuses := executor.WaitAndReturnFinalBatchStatuses(context.Background())
		elapsed := time.Since(start)

		assert.True(t, elapsed >= args.TimeForWaitOnEthereum)
		assert.Nil(t, statuses)
	})
	t.Run("context expiration", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		executor, _ := NewBridgeExecutor(args)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		start := time.Now()
		statuses := executor.WaitAndReturnFinalBatchStatuses(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForWaitOnEthereum)
		assert.Nil(t, statuses)
	})
	t.Run("GetBatchStatusesFromEthereum always returns err", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		counter := 0
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				counter++
				return nil, expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &bridgeCore.TransferBatch{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		statuses := executor.WaitAndReturnFinalBatchStatuses(ctx)

		assert.Equal(t, 10, counter)
		assert.Nil(t, statuses)
	})
	t.Run("GetBatchStatusesFromEthereum always returns success+statuses only after 4 checks", func(t *testing.T) {
		t.Parallel()

		providedStatuses := []byte{bridgeCore.Executed, bridgeCore.Rejected}
		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		counter := 0
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				counter++
				if counter >= 5 {
					return providedStatuses, nil
				}
				return nil, expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &bridgeCore.TransferBatch{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		start := time.Now()

		statuses := executor.WaitAndReturnFinalBatchStatuses(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForWaitOnEthereum)
		assert.Equal(t, 5, counter)
		assert.Equal(t, providedStatuses, statuses)
	})
	t.Run("GetBatchStatusesFromEthereum always returns success+statuses only after 4 checks, otherwise empty slice", func(t *testing.T) {
		t.Parallel()

		providedStatuses := []byte{bridgeCore.Executed, bridgeCore.Rejected}
		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		counter := 0
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				counter++
				if counter >= 5 {
					return providedStatuses, nil
				}
				return make([]byte, 0), nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &bridgeCore.TransferBatch{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		start := time.Now()

		statuses := executor.WaitAndReturnFinalBatchStatuses(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForWaitOnEthereum)
		assert.Equal(t, 5, counter)
		assert.Equal(t, providedStatuses, statuses)
	})
}

func TestResolveNewDepositsStatuses(t *testing.T) {
	t.Parallel()

	providedBatchForResolve := &bridgeCore.TransferBatch{
		Deposits: []*bridgeCore.DepositTransfer{
			{
				DisplayableTo: "to1",
			},
			{
				DisplayableTo: "to2",
			},
		},
		Statuses: make([]byte, 2),
	}

	t.Run("less new deposits", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatchForResolve.Clone()

		executor.ResolveNewDepositsStatuses(uint64(0))
		assert.Equal(t, []byte{bridgeCore.Rejected, bridgeCore.Rejected}, executor.batch.Statuses)

		executor.batch = providedBatchForResolve.Clone()
		executor.batch.ResolveNewDeposits(1)
		assert.Equal(t, []byte{0, bridgeCore.Rejected}, executor.batch.Statuses)
	})
	t.Run("equal new deposits", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatchForResolve.Clone()

		executor.ResolveNewDepositsStatuses(uint64(2))
		assert.Equal(t, []byte{0, 0}, executor.batch.Statuses)
	})
	t.Run("more new deposits", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatchForResolve.Clone()

		executor.ResolveNewDepositsStatuses(uint64(3))
		assert.Equal(t, []byte{0, 0, bridgeCore.Rejected}, executor.batch.Statuses)
	})
}

func TestEthToKCBridgeExecutor_setExecutionMessageInStatusHandler(t *testing.T) {
	t.Parallel()

	expectedString := "DEBUG: message a = 1 b = ff c = str"

	wasCalled := false
	args := createMockExecutorArgs()
	args.StatusHandler = &testsCommon.StatusHandlerStub{
		SetStringMetricCalled: func(metric string, val string) {
			wasCalled = true

			assert.Equal(t, metric, bridgeCore.MetricLastError)
			assert.Equal(t, expectedString, val)
		},
	}
	executor, _ := NewBridgeExecutor(args)
	executor.setExecutionMessageInStatusHandler(logger.LogDebug, "message", "a", 1, "b", []byte{255}, "c", "str")

	assert.True(t, wasCalled)
}

func TestSignaturesHolder_ClearStoredSignatures(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	wasCalled := false
	args.SignaturesHolder = &testsCommon.SignaturesHolderStub{
		ClearStoredSignaturesCalled: func() {
			wasCalled = true
		},
	}

	executor, _ := NewBridgeExecutor(args)
	executor.ClearStoredP2PSignaturesForEthereum()

	assert.True(t, wasCalled)
}

func TestBridgeExecutor_CheckKCClientAvailability(t *testing.T) {
	t.Parallel()

	checkAvailabilityCalled := false
	args := createMockExecutorArgs()
	args.KCClient = &bridgeTests.KCClientStub{
		CheckClientAvailabilityCalled: func(ctx context.Context) error {
			checkAvailabilityCalled = true
			return nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	err := executor.CheckKCClientAvailability(context.Background())

	assert.Nil(t, err)
	assert.True(t, checkAvailabilityCalled)
}

func TestBridgeExecutor_CheckEthereumClientAvailability(t *testing.T) {
	t.Parallel()

	checkAvailabilityCalled := false
	args := createMockExecutorArgs()
	args.EthereumClient = &bridgeTests.EthereumClientStub{
		CheckClientAvailabilityCalled: func(ctx context.Context) error {
			checkAvailabilityCalled = true
			return nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	err := executor.CheckEthereumClientAvailability(context.Background())

	assert.Nil(t, err)
	assert.True(t, checkAvailabilityCalled)
}

func TestBridgeExecutor_CheckAvailableTokens(t *testing.T) {
	t.Parallel()

	ethTokens := []common.Address{
		common.BytesToAddress([]byte("eth token 1")),
		common.BytesToAddress([]byte("eth token 1")),
		common.BytesToAddress([]byte("eth token 2")),
	}

	kdaTokens := [][]byte{
		[]byte("kda token 1"),
		[]byte("kda token 1"),
		[]byte("kda token 2"),
	}

	amounts := []*big.Int{
		big.NewInt(37),
		big.NewInt(38),
		big.NewInt(39),
	}

	testDirection := batchProcessor.FromKC
	checkedEthTokens := make([]common.Address, 0)
	checkedKdaTokens := make([][]byte, 0)
	checkedAmounts := make([]*big.Int, 0)

	args := createMockExecutorArgs()
	var returnedError error
	args.BalanceValidator = &testsCommon.BalanceValidatorStub{
		CheckTokenCalled: func(ctx context.Context, ethToken common.Address, kdaToken []byte, amount *big.Int, direction batchProcessor.Direction) error {
			checkedEthTokens = append(checkedEthTokens, ethToken)
			checkedKdaTokens = append(checkedKdaTokens, kdaToken)
			checkedAmounts = append(checkedAmounts, amount)

			assert.Equal(t, testDirection, direction)

			return returnedError
		},
	}
	executor, _ := NewBridgeExecutor(args)

	// do not run these tests in parallel
	t.Run("check validator does not error", func(t *testing.T) {
		returnedError = nil
		checkedEthTokens = make([]common.Address, 0)
		checkedKdaTokens = make([][]byte, 0)
		checkedAmounts = make([]*big.Int, 0)
		err := executor.CheckAvailableTokens(context.Background(), ethTokens, kdaTokens, amounts, testDirection)

		expectedEthTokens := []common.Address{
			common.BytesToAddress([]byte("eth token 1")),
			common.BytesToAddress([]byte("eth token 2")),
		}
		expectedKdaTokens := [][]byte{
			[]byte("kda token 1"),
			[]byte("kda token 2"),
		}
		expectedAmounts := []*big.Int{
			big.NewInt(75), // 37 + 38
			big.NewInt(39),
		}

		assert.Nil(t, err)
		assert.Equal(t, expectedEthTokens, checkedEthTokens)
		assert.Equal(t, expectedKdaTokens, checkedKdaTokens)
		assert.Equal(t, expectedAmounts, checkedAmounts)
	})
	t.Run("check validator returns error", func(t *testing.T) {
		returnedError = fmt.Errorf("expected error")
		checkedEthTokens = make([]common.Address, 0)
		checkedKdaTokens = make([][]byte, 0)
		checkedAmounts = make([]*big.Int, 0)
		err := executor.CheckAvailableTokens(context.Background(), ethTokens, kdaTokens, amounts, testDirection)

		expectedEthTokens := []common.Address{
			common.BytesToAddress([]byte("eth token 1")), // only the first token is checked
		}
		expectedKdaTokens := [][]byte{
			[]byte("kda token 1"),
		}
		expectedAmounts := []*big.Int{
			big.NewInt(75), // 37 + 38
		}

		assert.Equal(t, returnedError, err)
		assert.Equal(t, expectedEthTokens, checkedEthTokens)
		assert.Equal(t, expectedKdaTokens, checkedKdaTokens)
		assert.Equal(t, expectedAmounts, checkedAmounts)
	})
}
