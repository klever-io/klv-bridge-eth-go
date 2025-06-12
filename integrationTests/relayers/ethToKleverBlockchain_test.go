// go:build !slow

package relayers

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/klever-io/klv-bridge-eth-go/clients/ethereum/contract"
	"github.com/klever-io/klv-bridge-eth-go/config"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/factory"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests/mock"
	"github.com/klever-io/klv-bridge-eth-go/status"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const noGasStationURL = ""

type argsForSCCallsTest struct {
	providedScCallData []byte
	expectedScCallData []byte
}

func TestRelayersShouldExecuteTransfersFromEthToKleverchain(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("simple tokens transfers", func(t *testing.T) {
		testRelayersShouldExecuteTransfersFromEthToKleverchain(t, false)
	})
	t.Run("native tokens transfers", func(t *testing.T) {
		testRelayersShouldExecuteTransfersFromEthToKleverchain(t, true)
	})
}

func testRelayersShouldExecuteTransfersFromEthToKleverchain(t *testing.T, withNativeTokens bool) {
	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	token1Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker1 := "tck-000001"

	token2Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker2 := "tck-000002"

	value1 := big.NewInt(111111111)
	destination1 := testsCommon.CreateRandomKleverchainAddress()
	depositor1 := testsCommon.CreateRandomEthereumAddress()

	value2 := big.NewInt(222222222)
	destination2 := testsCommon.CreateRandomKleverchainAddress()
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
	ethereumChainMock.SetFinalNonce(batchNonceOnEthereum + 1)

	kleverchainChainMock := mock.NewKleverBlockchainMock()

	if !withNativeTokens {
		ethereumChainMock.UpdateNativeTokens(token1Erc20, true)
		ethereumChainMock.UpdateMintBurnTokens(token1Erc20, false)
		ethereumChainMock.UpdateTotalBalances(token1Erc20, value1)

		ethereumChainMock.UpdateNativeTokens(token2Erc20, true)
		ethereumChainMock.UpdateMintBurnTokens(token2Erc20, false)
		ethereumChainMock.UpdateTotalBalances(token2Erc20, value2)

		kleverchainChainMock.AddTokensPair(token1Erc20, ticker1, withNativeTokens, true, zero, zero, zero)
		kleverchainChainMock.AddTokensPair(token2Erc20, ticker2, withNativeTokens, true, zero, zero, zero)
	} else {
		ethereumChainMock.UpdateNativeTokens(token1Erc20, false)
		ethereumChainMock.UpdateMintBurnTokens(token1Erc20, true)
		ethereumChainMock.UpdateBurnBalances(token1Erc20, value1)
		ethereumChainMock.UpdateMintBalances(token1Erc20, value1)

		ethereumChainMock.UpdateNativeTokens(token2Erc20, false)
		ethereumChainMock.UpdateMintBurnTokens(token2Erc20, true)
		ethereumChainMock.UpdateBurnBalances(token2Erc20, value2)
		ethereumChainMock.UpdateMintBalances(token2Erc20, value2)

		kleverchainChainMock.AddTokensPair(token1Erc20, ticker1, withNativeTokens, true, zero, zero, value1)
		kleverchainChainMock.AddTokensPair(token2Erc20, ticker2, withNativeTokens, true, zero, zero, value2)
	}

	kleverchainChainMock.SetLastExecutedEthBatchID(batchNonceOnEthereum)
	kleverchainChainMock.SetLastExecutedEthTxId(txNonceOnEthereum)
	kleverchainChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return []byte{bridgeCore.Executed, bridgeCore.Rejected}
	}
	kleverchainChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer closeRelayers(relayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	kleverchainChainMock.ProcessFinishedHandler = func() {
		log.Info("kleverchainChainMock.ProcessFinishedHandler called")
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], kleverchainChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		relayer, err := factory.NewEthKleverBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		kleverchainChainMock.AddRelayer(relayer.KleverRelayerAddress())
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

	assert.NotNil(t, kleverchainChainMock.PerformedActionID())
	transfer := kleverchainChainMock.ProposedTransfer()
	require.NotNil(t, transfer)
	require.Equal(t, 2, len(transfer.Transfers))
	assert.Equal(t, batchNonceOnEthereum+1, transfer.BatchId.Uint64())

	assert.Equal(t, destination1.Bytes(), transfer.Transfers[0].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker1)), transfer.Transfers[0].Token)
	assert.Equal(t, value1, transfer.Transfers[0].Amount)
	assert.Equal(t, depositor1, common.BytesToAddress(transfer.Transfers[0].From))
	assert.Equal(t, txNonceOnEthereum+1, transfer.Transfers[0].Nonce.Uint64())
	assert.Equal(t, []byte{bridgeCore.MissingDataProtocolMarker}, transfer.Transfers[0].Data)

	assert.Equal(t, destination2.Bytes(), transfer.Transfers[1].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker2)), transfer.Transfers[1].Token)
	assert.Equal(t, value2, transfer.Transfers[1].Amount)
	assert.Equal(t, depositor2, common.BytesToAddress(transfer.Transfers[1].From))
	assert.Equal(t, txNonceOnEthereum+2, transfer.Transfers[1].Nonce.Uint64())
	assert.Equal(t, []byte{bridgeCore.MissingDataProtocolMarker}, transfer.Transfers[1].Data)
}

func TestRelayersShouldExecuteTransferFromEthToKleverchainHavingTxsWithSCcalls(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("correct SC call", func(t *testing.T) {
		testArgs := argsForSCCallsTest{
			providedScCallData: bridge.EthCallDataMock,
			expectedScCallData: bridge.CallDataMock,
		}

		testRelayersShouldExecuteTransferFromEthToKleverchainHavingTxsWithSCcalls(t, testArgs)
	})
	t.Run("no SC call", func(t *testing.T) {
		testArgs := argsForSCCallsTest{
			providedScCallData: []byte{bridgeCore.MissingDataProtocolMarker},
			expectedScCallData: []byte{bridgeCore.MissingDataProtocolMarker},
		}

		testRelayersShouldExecuteTransferFromEthToKleverchainHavingTxsWithSCcalls(t, testArgs)
	})
}

func testRelayersShouldExecuteTransferFromEthToKleverchainHavingTxsWithSCcalls(t *testing.T, args argsForSCCallsTest) {
	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()

	token1Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker1 := "tck-000001"

	token2Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker2 := "tck-000002"

	token3Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker3 := "tck-000003"

	value1 := big.NewInt(111111111)
	destination1 := testsCommon.CreateRandomKleverchainAddress()
	depositor1 := testsCommon.CreateRandomEthereumAddress()

	value2 := big.NewInt(222222222)
	destination2 := testsCommon.CreateRandomKleverchainAddress()
	depositor2 := testsCommon.CreateRandomEthereumAddress()

	depositor3 := testsCommon.CreateRandomEthereumAddress()

	value3 := big.NewInt(333333333)
	destination3Sc := testsCommon.CreateRandomKleverchainSCAddress()

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
		Depositor:    depositor3,
		Recipient:    destination3Sc.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.SetQuorum(numRelayers)
	ethereumChainMock.SetFinalNonce(batchNonceOnEthereum + 1)

	ethereumChainMock.UpdateNativeTokens(token1Erc20, true)
	ethereumChainMock.UpdateMintBurnTokens(token1Erc20, false)
	ethereumChainMock.UpdateTotalBalances(token1Erc20, value1)

	ethereumChainMock.UpdateNativeTokens(token2Erc20, true)
	ethereumChainMock.UpdateMintBurnTokens(token2Erc20, false)
	ethereumChainMock.UpdateTotalBalances(token2Erc20, value2)

	ethereumChainMock.UpdateNativeTokens(token3Erc20, true)
	ethereumChainMock.UpdateMintBurnTokens(token3Erc20, false)
	ethereumChainMock.UpdateTotalBalances(token3Erc20, value3)

	ethereumChainMock.FilterLogsCalled = func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
		expectedBatchNonceHash := []common.Hash{
			common.BytesToHash(big.NewInt(int64(batchNonceOnEthereum + 1)).Bytes()),
		}
		require.Equal(t, 2, len(q.Topics))
		assert.Equal(t, expectedBatchNonceHash, q.Topics[1])

		scExecAbi, err := contract.ERC20SafeMetaData.GetAbi()
		require.Nil(t, err)

		eventInputs := scExecAbi.Events["ERC20SCDeposit"].Inputs.NonIndexed()
		packedArgs, err := eventInputs.Pack(big.NewInt(0).SetUint64(txNonceOnEthereum+3), args.providedScCallData)
		require.Nil(t, err)

		scLog := types.Log{
			Data: packedArgs,
		}

		return []types.Log{scLog}, nil
	}

	kleverchainChainMock := mock.NewKleverBlockchainMock()
	kleverchainChainMock.AddTokensPair(token1Erc20, ticker1, false, true, zero, zero, zero)
	kleverchainChainMock.AddTokensPair(token2Erc20, ticker2, false, true, zero, zero, zero)
	kleverchainChainMock.AddTokensPair(token3Erc20, ticker3, false, true, zero, zero, zero)
	kleverchainChainMock.SetLastExecutedEthBatchID(batchNonceOnEthereum)
	kleverchainChainMock.SetLastExecutedEthTxId(txNonceOnEthereum)
	kleverchainChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return []byte{bridgeCore.Executed, bridgeCore.Rejected, bridgeCore.Executed}
	}
	kleverchainChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer func() {
		for _, r := range relayers {
			_ = r.Close()
		}
	}()

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	kleverchainChainMock.ProcessFinishedHandler = func() {
		log.Info("kleverchainChainMock.ProcessFinishedHandler called")
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], kleverchainChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		relayer, err := factory.NewEthKleverBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		kleverchainChainMock.AddRelayer(relayer.KleverRelayerAddress())
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

	assert.NotNil(t, kleverchainChainMock.PerformedActionID())
	transfer := kleverchainChainMock.ProposedTransfer()
	require.NotNil(t, transfer)
	require.Equal(t, 3, len(transfer.Transfers))
	assert.Equal(t, batchNonceOnEthereum+1, transfer.BatchId.Uint64())

	assert.Equal(t, destination1.Bytes(), transfer.Transfers[0].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker1)), transfer.Transfers[0].Token)
	assert.Equal(t, value1, transfer.Transfers[0].Amount)
	assert.Equal(t, depositor1, common.BytesToAddress(transfer.Transfers[0].From))
	assert.Equal(t, txNonceOnEthereum+1, transfer.Transfers[0].Nonce.Uint64())
	assert.Equal(t, []byte{bridgeCore.MissingDataProtocolMarker}, transfer.Transfers[0].Data)

	assert.Equal(t, destination2.Bytes(), transfer.Transfers[1].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker2)), transfer.Transfers[1].Token)
	assert.Equal(t, value2, transfer.Transfers[1].Amount)
	assert.Equal(t, depositor2, common.BytesToAddress(transfer.Transfers[1].From))
	assert.Equal(t, txNonceOnEthereum+2, transfer.Transfers[1].Nonce.Uint64())
	assert.Equal(t, []byte{bridgeCore.MissingDataProtocolMarker}, transfer.Transfers[1].Data)

	assert.Equal(t, destination3Sc.Bytes(), transfer.Transfers[2].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker3)), transfer.Transfers[2].Token)
	assert.Equal(t, value3, transfer.Transfers[2].Amount)
	assert.Equal(t, depositor3, common.BytesToAddress(transfer.Transfers[2].From))
	assert.Equal(t, txNonceOnEthereum+3, transfer.Transfers[2].Nonce.Uint64())
	assert.Equal(t, args.expectedScCallData, transfer.Transfers[2].Data)
}

func createMockBridgeComponentsArgs(
	index int,
	messenger p2p.Messenger,
	kleverChainMock *mock.KleverBlockchainMock,
	ethereumChainMock *mock.EthereumChainMock,
) factory.ArgsEthereumToKleverBridge {

	generalConfigs := CreateBridgeComponentsConfig(index, "testdata", noGasStationURL)
	return factory.ArgsEthereumToKleverBridge{
		Configs: config.Configs{
			GeneralConfig:   generalConfigs,
			ApiRoutesConfig: config.ApiRoutesConfig{},
			FlagsConfig: config.ContextFlagsConfig{
				RestApiInterface: bridgeCore.WebServerOffString,
			},
		},
		Proxy:                     kleverChainMock,
		ClientWrapper:             ethereumChainMock,
		Messenger:                 messenger,
		StatusStorer:              testsCommon.NewStorerMock(),
		TimeForBootstrap:          time.Second * 5,
		TimeBeforeRepeatJoin:      time.Second * 30,
		MetricsHolder:             status.NewMetricsHolder(),
		AppStatusHandler:          &statusHandler.AppStatusHandlerStub{},
		KleverClientStatusHandler: &testsCommon.StatusHandlerStub{},
	}
}
