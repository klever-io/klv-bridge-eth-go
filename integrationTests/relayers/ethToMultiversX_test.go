package relayers

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/chain"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/factory"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/status"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	chainConfig "github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayersShouldExecuteTransferFromEthToMultiversX(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	token1Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker1 := "tck-000001"

	token2Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker2 := "tck-000002"

	value1 := big.NewInt(111111111)
	destination1 := testsCommon.CreateRandomMultiversXAddress()
	depositor1 := testsCommon.CreateRandomEthereumAddress()

	value2 := big.NewInt(222222222)
	destination2 := testsCommon.CreateRandomMultiversXAddress()
	depositor2 := testsCommon.CreateRandomEthereumAddress()

	tokens := []common.Address{token1Erc20, token2Erc20}
	availableBalances := []*big.Int{value1, value2}

	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	batchNonceOnEthereum := uint64(345)
	txNonceOnEthereum := uint64(772634)
	batch := contract.Batch{
		Nonce:                  big.NewInt(int64(batchNonceOnEthereum) + 1),
		BlockNumber:            0,
		LastUpdatedBlockNumber: 0,
		DepositsCount:          2,
	}

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 1),
		TokenAddress: token1Erc20,
		Amount:       value1,
		Depositor:    depositor1,
		Recipient:    destination1.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 2),
		TokenAddress: token2Erc20,
		Amount:       value2,
		Depositor:    depositor2,
		Recipient:    destination2.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.SetQuorum(numRelayers)

	multiversXChainMock := mock.NewMultiversXChainMock(testsCommon.CreateRandomEthereumAddress().Bytes())
	multiversXChainMock.AddTokensPair(token1Erc20, ticker1)
	multiversXChainMock.AddTokensPair(token2Erc20, ticker2)
	multiversXChainMock.SetLastExecutedEthBatchID(batchNonceOnEthereum)
	multiversXChainMock.SetLastExecutedEthTxId(txNonceOnEthereum)
	multiversXChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return []byte{clients.Executed, clients.Rejected}
	}
	multiversXChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer func() {
		for _, r := range relayers {
			_ = r.Close()
		}
	}()

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	multiversXChainMock.ProcessFinishedHandler = func() {
		log.Info("multiversXChainMock.ProcessFinishedHandler called")
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], multiversXChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		relayer, err := factory.NewEthMultiversXBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		multiversXChainMock.AddRelayer(relayer.MultiversXRelayerAddress())
		ethereumChainMock.AddRelayer(relayer.EthereumRelayerAddress())

		go func() {
			err = relayer.Start()
			integrationTests.Log.LogIfError(err)
			require.Nil(t, err)
		}()

		relayers = append(relayers, relayer)
	}

	<-ctx.Done()
	time.Sleep(time.Second * 5)

	assert.NotNil(t, multiversXChainMock.PerformedActionID())
	transfer := multiversXChainMock.ProposedTransfer()
	require.NotNil(t, transfer)
	require.Equal(t, 2, len(transfer.Transfers))
	assert.Equal(t, batchNonceOnEthereum+1, transfer.BatchId.Uint64())

	assert.Equal(t, destination1.AddressBytes(), transfer.Transfers[0].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker1)), transfer.Transfers[0].Token)
	assert.Equal(t, value1, transfer.Transfers[0].Amount)
	assert.Equal(t, depositor1, common.BytesToAddress(transfer.Transfers[0].From))
	assert.Equal(t, txNonceOnEthereum+1, transfer.Transfers[0].Nonce.Uint64())
	assert.Empty(t, transfer.Transfers[0].Data)
	assert.Zero(t, transfer.Transfers[0].ExtraGas)

	assert.Equal(t, destination2.AddressBytes(), transfer.Transfers[1].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker2)), transfer.Transfers[1].Token)
	assert.Equal(t, value2, transfer.Transfers[1].Amount)
	assert.Equal(t, depositor2, common.BytesToAddress(transfer.Transfers[1].From))
	assert.Equal(t, txNonceOnEthereum+2, transfer.Transfers[1].Nonce.Uint64())
	assert.Empty(t, transfer.Transfers[1].Data)
	assert.Zero(t, transfer.Transfers[1].ExtraGas)
}

func TestRelayersShouldExecuteTransferFromEthToMultiversXHavingTxsWithSCcalls(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	scExecProxyEthAddress := testsCommon.CreateRandomEthereumAddress()

	token1Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker1 := "tck-000001"

	token2Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker2 := "tck-000002"

	token3Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker3 := "tck-000003"

	value1 := big.NewInt(111111111)
	destination1 := testsCommon.CreateRandomMultiversXAddress()
	depositor1 := testsCommon.CreateRandomEthereumAddress()

	value2 := big.NewInt(222222222)
	destination2 := testsCommon.CreateRandomMultiversXAddress()
	depositor2 := testsCommon.CreateRandomEthereumAddress()

	value3 := big.NewInt(333333333)
	destination3 := testsCommon.CreateRandomMultiversXAddress()

	tokens := []common.Address{token1Erc20, token2Erc20, token3Erc20}
	availableBalances := []*big.Int{value1, value2, value3}

	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	batchNonceOnEthereum := uint64(345)
	txNonceOnEthereum := uint64(772634)
	batch := contract.Batch{
		Nonce:                  big.NewInt(int64(batchNonceOnEthereum) + 1),
		BlockNumber:            0,
		LastUpdatedBlockNumber: 0,
		DepositsCount:          3,
	}

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 1),
		TokenAddress: token1Erc20,
		Amount:       value1,
		Depositor:    depositor1,
		Recipient:    destination1.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 2),
		TokenAddress: token2Erc20,
		Amount:       value2,
		Depositor:    depositor2,
		Recipient:    destination2.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 3),
		TokenAddress: token3Erc20,
		Amount:       value3,
		Depositor:    scExecProxyEthAddress,
		Recipient:    destination3.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.SetQuorum(numRelayers)

	mvxGasLimit := uint64(3737373)
	callData := "doSomething@aabbcc@001122"

	ethereumChainMock.FilterLogsCalled = func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
		expectedBatchNonceHash := []common.Hash{
			common.BytesToHash(big.NewInt(int64(batchNonceOnEthereum + 1)).Bytes()),
		}
		require.Equal(t, 2, len(q.Topics))
		assert.Equal(t, expectedBatchNonceHash, q.Topics[1])

		scExecAbi, err := contract.SCExecProxyMetaData.GetAbi()
		require.Nil(t, err)

		eventInputs := scExecAbi.Events["ERC20SCDeposit"].Inputs.NonIndexed()
		packedArgs, err := eventInputs.Pack(txNonceOnEthereum+3, mvxGasLimit, callData)
		require.Nil(t, err)

		scLog := types.Log{
			Data: packedArgs,
		}

		return []types.Log{scLog}, nil
	}

	multiversXChainMock := mock.NewMultiversXChainMock(scExecProxyEthAddress.Bytes())
	multiversXChainMock.AddTokensPair(token1Erc20, ticker1)
	multiversXChainMock.AddTokensPair(token2Erc20, ticker2)
	multiversXChainMock.AddTokensPair(token3Erc20, ticker3)
	multiversXChainMock.SetLastExecutedEthBatchID(batchNonceOnEthereum)
	multiversXChainMock.SetLastExecutedEthTxId(txNonceOnEthereum)
	multiversXChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return []byte{clients.Executed, clients.Rejected, clients.Executed}
	}
	multiversXChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer func() {
		for _, r := range relayers {
			_ = r.Close()
		}
	}()

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	multiversXChainMock.ProcessFinishedHandler = func() {
		log.Info("multiversXChainMock.ProcessFinishedHandler called")
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], multiversXChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Configs.GeneralConfig.Eth.SCExecProxyAddress = scExecProxyEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		relayer, err := factory.NewEthMultiversXBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		multiversXChainMock.AddRelayer(relayer.MultiversXRelayerAddress())
		ethereumChainMock.AddRelayer(relayer.EthereumRelayerAddress())

		go func() {
			err = relayer.Start()
			integrationTests.Log.LogIfError(err)
			require.Nil(t, err)
		}()

		relayers = append(relayers, relayer)
	}

	<-ctx.Done()
	time.Sleep(time.Second * 5)

	assert.NotNil(t, multiversXChainMock.PerformedActionID())
	transfer := multiversXChainMock.ProposedTransfer()
	require.NotNil(t, transfer)
	require.Equal(t, 3, len(transfer.Transfers))
	assert.Equal(t, batchNonceOnEthereum+1, transfer.BatchId.Uint64())

	assert.Equal(t, destination1.AddressBytes(), transfer.Transfers[0].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker1)), transfer.Transfers[0].Token)
	assert.Equal(t, value1, transfer.Transfers[0].Amount)
	assert.Equal(t, depositor1, common.BytesToAddress(transfer.Transfers[0].From))
	assert.Equal(t, txNonceOnEthereum+1, transfer.Transfers[0].Nonce.Uint64())
	assert.Empty(t, transfer.Transfers[0].Data)
	assert.Zero(t, transfer.Transfers[0].ExtraGas)

	assert.Equal(t, destination2.AddressBytes(), transfer.Transfers[1].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker2)), transfer.Transfers[1].Token)
	assert.Equal(t, value2, transfer.Transfers[1].Amount)
	assert.Equal(t, depositor2, common.BytesToAddress(transfer.Transfers[1].From))
	assert.Equal(t, txNonceOnEthereum+2, transfer.Transfers[1].Nonce.Uint64())
	assert.Empty(t, transfer.Transfers[1].Data)
	assert.Zero(t, transfer.Transfers[1].ExtraGas)

	assert.Equal(t, destination3.AddressBytes(), transfer.Transfers[2].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker3)), transfer.Transfers[2].Token)
	assert.Equal(t, value3, transfer.Transfers[2].Amount)
	assert.Equal(t, scExecProxyEthAddress, common.BytesToAddress(transfer.Transfers[2].From))
	assert.Equal(t, txNonceOnEthereum+3, transfer.Transfers[2].Nonce.Uint64())
	assert.Equal(t, callData, string(transfer.Transfers[2].Data))
	assert.Equal(t, mvxGasLimit, transfer.Transfers[2].ExtraGas)
}

func createMockBridgeComponentsArgs(
	index int,
	messenger p2p.Messenger,
	multiversXChainMock *mock.MultiversXChainMock,
	ethereumChainMock *mock.EthereumChainMock,
) factory.ArgsEthereumToMultiversXBridge {

	generalConfigs := createBridgeComponentsConfig(index)
	return factory.ArgsEthereumToMultiversXBridge{
		Configs: config.Configs{
			GeneralConfig:   generalConfigs,
			ApiRoutesConfig: config.ApiRoutesConfig{},
			FlagsConfig: config.ContextFlagsConfig{
				RestApiInterface: core.WebServerOffString,
			},
		},
		Proxy:                         multiversXChainMock,
		ClientWrapper:                 ethereumChainMock,
		Messenger:                     messenger,
		StatusStorer:                  testsCommon.NewStorerMock(),
		TimeForBootstrap:              time.Second * 5,
		TimeBeforeRepeatJoin:          time.Second * 30,
		MetricsHolder:                 status.NewMetricsHolder(),
		AppStatusHandler:              &statusHandler.AppStatusHandlerStub{},
		MultiversXClientStatusHandler: &testsCommon.StatusHandlerStub{},
	}
}

func createBridgeComponentsConfig(index int) config.Config {
	stateMachineConfig := config.ConfigStateMachine{
		StepDurationInMillis:       1000,
		IntervalForLeaderInSeconds: 60,
	}

	return config.Config{
		Eth: config.EthereumConfig{
			Chain:                        chain.Ethereum,
			NetworkAddress:               "mock",
			MultisigContractAddress:      "3009d97FfeD62E57d444e552A9eDF9Ee6Bc8644c",
			PrivateKeyFile:               fmt.Sprintf("testdata/ethereum%d.sk", index),
			IntervalToResendTxsInSeconds: 10,
			GasLimitBase:                 200000,
			GasLimitForEach:              30000,
			GasStation: config.GasStationConfig{
				Enabled: false,
			},
			MaxRetriesOnQuorumReached:          1,
			IntervalToWaitForTransferInSeconds: 1,
			MaxBlocksDelta:                     5,
		},
		MultiversX: config.MultiversXConfig{
			NetworkAddress:                  "mock",
			MultisigContractAddress:         "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf",
			PrivateKeyFile:                  fmt.Sprintf("testdata/multiversx%d.pem", index),
			IntervalToResendTxsInSeconds:    10,
			GasMap:                          testsCommon.CreateTestMultiversXGasMap(),
			MaxRetriesOnQuorumReached:       1,
			MaxRetriesOnWasTransferProposed: 3,
			ProxyMaxNoncesDelta:             5,
		},
		P2P: config.ConfigP2P{},
		StateMachine: map[string]config.ConfigStateMachine{
			"EthereumToMultiversX": stateMachineConfig,
			"MultiversXToEthereum": stateMachineConfig,
		},
		Relayer: config.ConfigRelayer{
			Marshalizer: chainConfig.MarshalizerConfig{
				Type:           "json",
				SizeCheckDelta: 10,
			},
			RoleProvider: config.RoleProviderConfig{
				PollingIntervalInMillis: 1000,
			},
		},
	}
}
