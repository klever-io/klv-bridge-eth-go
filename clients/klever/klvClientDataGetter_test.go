package klever

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/klever-io/klever-go/data/vm"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/builders"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	bridgeErrors "github.com/klever-io/klv-bridge-eth-go/errors"
	bridgeTests "github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon/interactors"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

const (
	returnCode     = "return code"
	returnMessage  = "return message"
	calledFunction = "called function"
)

var calledArgs = []string{"args1", "args2"}

func createMockArgsKLVClientDataGetter() ArgsKLVClientDataGetter {
	args := ArgsKLVClientDataGetter{
		Log:   logger.GetOrCreate("test"),
		Proxy: &interactors.ProxyStub{},
	}

	args.MultisigContractAddress, _ = address.NewAddress("klv1qqqqqqqqqqqqqpgqh46r9zh78lry2py8tq723fpjdr4pp0zgsg8syf6mq0")
	args.SafeContractAddress, _ = address.NewAddress("klv1qqqqqqqqqqqqqpgqxjgmvqe9kvvr4xvvxflue3a7cjjeyvx9sg8snh0ljc")
	args.RelayerAddress, _ = address.NewAddress("klv12e0kqcvqsrayj8j0c4dqjyvnv4ep253m5anx4rfj4jeq34lxsg8s84ec9j")

	return args
}

func getBech32Address(addressHandler address.Address) string {
	return addressHandler.Bech32()
}

func createMockProxy(returningBytes [][]byte) *interactors.ProxyStub {
	return &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: returningBytes,
				},
			}, nil
		},
	}
}

func createMockBatch() *bridgeCore.TransferBatch {
	return &bridgeCore.TransferBatch{
		ID: 112233,
		Deposits: []*bridgeCore.DepositTransfer{
			{
				Nonce:                 1,
				ToBytes:               []byte("to1"),
				DisplayableTo:         "to1",
				FromBytes:             []byte("from1"),
				DisplayableFrom:       "from1",
				SourceTokenBytes:      []byte("token1"),
				DestinationTokenBytes: []byte("converted_token1"),
				DisplayableToken:      "token1",
				Amount:                big.NewInt(2),
				Data:                  []byte{0x00},
				DisplayableData:       "00",
			},
			{
				Nonce:                 3,
				ToBytes:               []byte("to2"),
				DisplayableTo:         "to2",
				FromBytes:             []byte("from2"),
				DisplayableFrom:       "from2",
				SourceTokenBytes:      []byte("token2"),
				DestinationTokenBytes: []byte("converted_token2"),
				DisplayableToken:      "token2",
				Amount:                big.NewInt(4),
				Data:                  []byte{0x00},
				DisplayableData:       "00",
			},
		},
		Statuses: []byte{bridgeCore.Rejected, bridgeCore.Executed},
	}
}

func TestNewKLVClientDataGetter(t *testing.T) {
	t.Parallel()

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.Log = nil

		dg, err := NewKLVClientDataGetter(args)
		assert.Equal(t, errNilLogger, err)
		assert.True(t, check.IfNil(dg))
	})
	t.Run("nil proxy", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.Proxy = nil

		dg, err := NewKLVClientDataGetter(args)
		assert.Equal(t, errNilProxy, err)
		assert.True(t, check.IfNil(dg))
	})
	t.Run("nil multisig contact address", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.MultisigContractAddress = nil

		dg, err := NewKLVClientDataGetter(args)
		assert.True(t, errors.Is(err, errNilAddressHandler))
		assert.True(t, strings.Contains(err.Error(), "MultisigContractAddress"))
		assert.True(t, check.IfNil(dg))
	})
	t.Run("nil safe contact address", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.SafeContractAddress = nil

		dg, err := NewKLVClientDataGetter(args)
		assert.True(t, errors.Is(err, errNilAddressHandler))
		assert.True(t, strings.Contains(err.Error(), "SafeContractAddress"))
		assert.True(t, check.IfNil(dg))
	})
	t.Run("nil relayer address", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.RelayerAddress = nil

		dg, err := NewKLVClientDataGetter(args)
		assert.True(t, errors.Is(err, errNilAddressHandler))
		assert.True(t, strings.Contains(err.Error(), "RelayerAddress"))
		assert.True(t, check.IfNil(dg))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()

		dg, err := NewKLVClientDataGetter(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(dg))
	})
}

func TestKlvClientDataGetter_ExecuteQueryReturningBytes(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	t.Run("nil vm ", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.ExecuteQueryReturningBytes(context.Background(), nil)
		assert.Nil(t, result)
		assert.Equal(t, errNilRequest, err)
	})
	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}

		result, err := dg.ExecuteQueryReturningBytes(context.Background(), &models.VmValueRequest{})
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("return code not ok", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)

		expectedErr := bridgeErrors.NewQueryResponseError(returnCode, returnMessage, calledFunction, getBech32Address(dg.multisigContractAddress), calledArgs...)
		dg.proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				return &models.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData:      nil,
						ReturnCode:      returnCode,
						ReturnMessage:   returnMessage,
						GasRemaining:    0,
						OutputAccounts:  nil,
						DeletedAccounts: nil,
						Logs:            nil,
					},
				}, nil
			},
		}

		request := &models.VmValueRequest{
			Address:    getBech32Address(dg.multisigContractAddress),
			FuncName:   calledFunction,
			CallerAddr: getBech32Address(dg.relayerAddress),
			CallValue:  map[string]int64{},
			Args:       calledArgs,
		}

		result, err := dg.ExecuteQueryReturningBytes(context.Background(), request)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)

		retData := [][]byte{[]byte("response 1"), []byte("response 2")}
		dg.proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				return &models.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData:      retData,
						ReturnCode:      okCodeAfterExecution,
						ReturnMessage:   returnMessage,
						GasRemaining:    0,
						OutputAccounts:  nil,
						DeletedAccounts: nil,
						Logs:            nil,
					},
				}, nil
			},
		}

		request := &models.VmValueRequest{
			Address:    getBech32Address(dg.multisigContractAddress),
			FuncName:   calledFunction,
			CallerAddr: getBech32Address(dg.relayerAddress),
			CallValue:  map[string]int64{},
			Args:       calledArgs,
		}

		result, err := dg.ExecuteQueryReturningBytes(context.Background(), request)
		assert.Nil(t, err)
		assert.Equal(t, retData, result)
	})
}

func TestKlvClientDataGetter_ExecuteQueryReturningBool(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.ExecuteQueryReturningBool(context.Background(), nil)
		assert.False(t, result)
		assert.Equal(t, errNilRequest, err)
	})
	t.Run("empty response", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		dg.proxy = createMockProxy(make([][]byte, 0))

		result, err := dg.ExecuteQueryReturningBool(context.Background(), &models.VmValueRequest{})
		assert.False(t, result)
		assert.Nil(t, err)
	})
	t.Run("empty byte slice on first element", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		dg.proxy = createMockProxy([][]byte{make([]byte, 0)})

		result, err := dg.ExecuteQueryReturningBool(context.Background(), &models.VmValueRequest{})
		assert.False(t, result)
		assert.Nil(t, err)
	})
	t.Run("not a bool result", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		dg.proxy = createMockProxy([][]byte{[]byte("random bytes")})

		expectedError := bridgeErrors.NewQueryResponseError(
			internalError,
			`error converting the received bytes to bool, strconv.ParseBool: parsing "114": invalid syntax`,
			"",
			"",
		)

		result, err := dg.ExecuteQueryReturningBool(context.Background(), &models.VmValueRequest{})
		assert.False(t, result)
		assert.Equal(t, expectedError, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		dg.proxy = createMockProxy([][]byte{{1}})

		result, err := dg.ExecuteQueryReturningBool(context.Background(), &models.VmValueRequest{})
		assert.True(t, result)
		assert.Nil(t, err)

		dg.proxy = createMockProxy([][]byte{{0}})

		result, err = dg.ExecuteQueryReturningBool(context.Background(), &models.VmValueRequest{})
		assert.False(t, result)
		assert.Nil(t, err)
	})
}

func TestKlvClientDataGetter_ExecuteQueryReturningUint64(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.ExecuteQueryReturningUint64(context.Background(), nil)
		assert.Zero(t, result)
		assert.Equal(t, errNilRequest, err)
	})
	t.Run("empty response", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		dg.proxy = createMockProxy(make([][]byte, 0))

		result, err := dg.ExecuteQueryReturningUint64(context.Background(), &models.VmValueRequest{})
		assert.Zero(t, result)
		assert.Nil(t, err)
	})
	t.Run("empty byte slice on first element", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		dg.proxy = createMockProxy([][]byte{make([]byte, 0)})

		result, err := dg.ExecuteQueryReturningUint64(context.Background(), &models.VmValueRequest{})
		assert.Zero(t, result)
		assert.Nil(t, err)
	})
	t.Run("large buffer", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		dg.proxy = createMockProxy([][]byte{[]byte("random bytes")})

		expectedError := bridgeErrors.NewQueryResponseError(
			internalError,
			errNotUint64Bytes.Error(),
			"",
			"",
		)

		result, err := dg.ExecuteQueryReturningUint64(context.Background(), &models.VmValueRequest{})
		assert.Zero(t, result)
		assert.Equal(t, expectedError, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		dg.proxy = createMockProxy([][]byte{{1}})

		result, err := dg.ExecuteQueryReturningUint64(context.Background(), &models.VmValueRequest{})
		assert.Equal(t, uint64(1), result)
		assert.Nil(t, err)

		dg.proxy = createMockProxy([][]byte{{0xFF, 0xFF}})

		result, err = dg.ExecuteQueryReturningUint64(context.Background(), &models.VmValueRequest{})
		assert.Equal(t, uint64(65535), result)
		assert.Nil(t, err)
	})
}

func TestKlvClientDataGetter_ExecuteQueryReturningBigInt(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.ExecuteQueryReturningBigInt(context.Background(), nil)
		assert.Nil(t, result)
		assert.Equal(t, errNilRequest, err)
	})
	t.Run("empty response", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		dg.proxy = createMockProxy(make([][]byte, 0))

		result, err := dg.ExecuteQueryReturningBigInt(context.Background(), &models.VmValueRequest{})
		assert.Equal(t, big.NewInt(0), result)
		assert.Nil(t, err)
	})
	t.Run("empty byte slice on first element", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		dg.proxy = createMockProxy([][]byte{make([]byte, 0)})

		result, err := dg.ExecuteQueryReturningBigInt(context.Background(), &models.VmValueRequest{})
		assert.Equal(t, big.NewInt(0), result)
		assert.Nil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewKLVClientDataGetter(args)
		largeNumber := new(big.Int)
		largeNumber.SetString("18446744073709551616", 10)
		dg.proxy = createMockProxy([][]byte{largeNumber.Bytes()})

		result, err := dg.ExecuteQueryReturningBigInt(context.Background(), &models.VmValueRequest{})
		assert.Equal(t, largeNumber, result)
		assert.Nil(t, err)

		dg.proxy = createMockProxy([][]byte{{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}})

		result, err = dg.ExecuteQueryReturningBigInt(context.Background(), &models.VmValueRequest{})
		largeNumber.SetString("79228162514264337593543950335", 10)
		assert.Equal(t, largeNumber, result)
		assert.Nil(t, err)
	})
}

func TestKlvClientDataGetter_GetCurrentBatchAsDataBytes(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	returningBytes := [][]byte{[]byte("buff0"), []byte("buff1"), []byte("buff2")}
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, getCurrentTxBatchFuncName, vmRequest.FuncName)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: returningBytes,
				},
			}, nil
		},
	}
	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.GetCurrentBatchAsDataBytes(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, returningBytes, result)
}

func TestExecuteQueryFromBuilderReturnErr(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	expectedError := errors.New("expected error")
	erc20Address := "erc20Address"
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: internalError,
					ReturnData: [][]byte{},
				},
			}, expectedError
		},
	}
	dg, _ := NewKLVClientDataGetter(args)

	_, err := dg.GetTokenIdForErc20Address(context.Background(), []byte(erc20Address))
	assert.Equal(t, expectedError, err)
}

func TestKlvClientDataGetter_GetTokenIdForErc20Address(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	klvAddress := "klvAddress"
	erc20Address := "erc20Address"
	returningBytes := [][]byte{[]byte(klvAddress)}
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, []string{hex.EncodeToString([]byte(erc20Address))}, vmRequest.Args)
			assert.Equal(t, getTokenIdForErc20AddressFuncName, vmRequest.FuncName)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: returningBytes,
				},
			}, nil
		},
	}
	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.GetTokenIdForErc20Address(context.Background(), []byte(erc20Address))

	assert.Nil(t, err)
	assert.Equal(t, returningBytes, result)
}

func TestKlvClientDataGetter_GetERC20AddressForTokenId(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	klvAddress := "klvAddress"
	erc20Address := "erc20Address"
	returningBytes := [][]byte{[]byte(erc20Address)}
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, []string{hex.EncodeToString([]byte(klvAddress))}, vmRequest.Args)
			assert.Equal(t, getErc20AddressForTokenIdFuncName, vmRequest.FuncName)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: returningBytes,
				},
			}, nil
		},
	}
	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.GetERC20AddressForTokenId(context.Background(), []byte(klvAddress))

	assert.Nil(t, err)
	assert.Equal(t, returningBytes, result)
}

func TestKlvClientDataGetter_WasProposedTransfer(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.WasProposedTransfer(context.Background(), nil)
		assert.False(t, result)
		assert.Equal(t, clients.ErrNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		proxyCalled := false
		batch := createMockBatch()

		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				proxyCalled = true
				assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
				assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
				assert.Equal(t, 0, len(vmRequest.CallValue))
				assert.Equal(t, wasTransferActionProposedFuncName, vmRequest.FuncName)

				expectedArgs := []string{
					hex.EncodeToString(big.NewInt(112233).Bytes()),

					hex.EncodeToString([]byte("from1")),
					hex.EncodeToString([]byte("to1")),
					hex.EncodeToString([]byte("converted_token1")),
					hex.EncodeToString(big.NewInt(2).Bytes()),
					hex.EncodeToString(big.NewInt(1).Bytes()),
					hex.EncodeToString([]byte{bridgeCore.MissingDataProtocolMarker}),

					hex.EncodeToString([]byte("from2")),
					hex.EncodeToString([]byte("to2")),
					hex.EncodeToString([]byte("converted_token2")),
					hex.EncodeToString(big.NewInt(4).Bytes()),
					hex.EncodeToString(big.NewInt(3).Bytes()),
					hex.EncodeToString([]byte{bridgeCore.MissingDataProtocolMarker}),
				}

				assert.Equal(t, expectedArgs, vmRequest.Args)

				return &models.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{{1}},
					},
				}, nil
			},
		}

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.WasProposedTransfer(context.Background(), batch)
		assert.True(t, result)
		assert.Nil(t, err)
		assert.True(t, proxyCalled)
	})
	t.Run("should work with SC calls", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		proxyCalled := false
		batch := createMockBatch()
		batch.Deposits[0].Data = bridgeTests.CallDataMock

		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				proxyCalled = true
				assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
				assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
				assert.Equal(t, 0, len(vmRequest.CallValue))
				assert.Equal(t, wasTransferActionProposedFuncName, vmRequest.FuncName)

				expectedArgs := []string{
					hex.EncodeToString(big.NewInt(112233).Bytes()),

					hex.EncodeToString([]byte("from1")),
					hex.EncodeToString([]byte("to1")),
					hex.EncodeToString([]byte("converted_token1")),
					hex.EncodeToString(big.NewInt(2).Bytes()),
					hex.EncodeToString(big.NewInt(1).Bytes()),
					hex.EncodeToString(bridgeTests.CallDataMock),

					hex.EncodeToString([]byte("from2")),
					hex.EncodeToString([]byte("to2")),
					hex.EncodeToString([]byte("converted_token2")),
					hex.EncodeToString(big.NewInt(4).Bytes()),
					hex.EncodeToString(big.NewInt(3).Bytes()),
					hex.EncodeToString([]byte{bridgeCore.MissingDataProtocolMarker}),
				}

				assert.Equal(t, expectedArgs, vmRequest.Args)

				return &models.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{{1}},
					},
				}, nil
			},
		}

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.WasProposedTransfer(context.Background(), batch)
		assert.True(t, result)
		assert.Nil(t, err)
		assert.True(t, proxyCalled)
	})
}

func TestKlvClientDataGetter_WasExecuted(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, wasActionExecutedFuncName, vmRequest.FuncName)

			expectedArgs := []string{hex.EncodeToString(big.NewInt(112233).Bytes())}
			assert.Equal(t, expectedArgs, vmRequest.Args)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{{1}},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.WasExecuted(context.Background(), 112233)
	assert.Nil(t, err)
	assert.True(t, proxyCalled)
	assert.True(t, result)
}

func TestKlvClientDataGetter_executeQueryWithErroredBuilder(t *testing.T) {
	t.Parallel()

	builder := builders.NewVMQueryBuilder().ArgBytes(nil)

	args := createMockArgsKLVClientDataGetter()
	dg, _ := NewKLVClientDataGetter(args)

	resultBytes, err := dg.executeQueryFromBuilder(context.Background(), builder)
	assert.Nil(t, resultBytes)
	assert.True(t, errors.Is(err, builders.ErrInvalidValue))
	assert.True(t, strings.Contains(err.Error(), "builder.ArgBytes"))

	resultUint64, err := dg.executeQueryUint64FromBuilder(context.Background(), builder)
	assert.Zero(t, resultUint64)
	assert.True(t, errors.Is(err, builders.ErrInvalidValue))
	assert.True(t, strings.Contains(err.Error(), "builder.ArgBytes"))

	resultBool, err := dg.executeQueryBoolFromBuilder(context.Background(), builder)
	assert.False(t, resultBool)
	assert.True(t, errors.Is(err, builders.ErrInvalidValue))
	assert.True(t, strings.Contains(err.Error(), "builder.ArgBytes"))
}

func TestKlvClientDataGetter_GetActionIDForProposeTransfer(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetActionIDForProposeTransfer(context.Background(), nil)
		assert.Zero(t, result)
		assert.Equal(t, clients.ErrNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		proxyCalled := false
		batch := createMockBatch()
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				proxyCalled = true
				assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
				assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
				assert.Equal(t, 0, len(vmRequest.CallValue))
				assert.Equal(t, getActionIdForTransferBatchFuncName, vmRequest.FuncName)

				expectedArgs := []string{
					hex.EncodeToString(big.NewInt(112233).Bytes()),

					hex.EncodeToString([]byte("from1")),
					hex.EncodeToString([]byte("to1")),
					hex.EncodeToString([]byte("converted_token1")),
					hex.EncodeToString(big.NewInt(2).Bytes()),
					hex.EncodeToString(big.NewInt(1).Bytes()),
					hex.EncodeToString([]byte{bridgeCore.MissingDataProtocolMarker}),

					hex.EncodeToString([]byte("from2")),
					hex.EncodeToString([]byte("to2")),
					hex.EncodeToString([]byte("converted_token2")),
					hex.EncodeToString(big.NewInt(4).Bytes()),
					hex.EncodeToString(big.NewInt(3).Bytes()),
					hex.EncodeToString([]byte{bridgeCore.MissingDataProtocolMarker}),
				}

				assert.Equal(t, expectedArgs, vmRequest.Args)

				return &models.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{big.NewInt(1234).Bytes()},
					},
				}, nil
			},
		}

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetActionIDForProposeTransfer(context.Background(), batch)
		assert.Equal(t, uint64(1234), result)
		assert.Nil(t, err)
		assert.True(t, proxyCalled)
	})
	t.Run("should work with SC calls", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		proxyCalled := false
		batch := createMockBatch()
		batch.Deposits[0].Data = bridgeTests.CallDataMock
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				proxyCalled = true
				assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
				assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
				assert.Equal(t, 0, len(vmRequest.CallValue))
				assert.Equal(t, getActionIdForTransferBatchFuncName, vmRequest.FuncName)

				expectedArgs := []string{
					hex.EncodeToString(big.NewInt(112233).Bytes()),

					hex.EncodeToString([]byte("from1")),
					hex.EncodeToString([]byte("to1")),
					hex.EncodeToString([]byte("converted_token1")),
					hex.EncodeToString(big.NewInt(2).Bytes()),
					hex.EncodeToString(big.NewInt(1).Bytes()),
					hex.EncodeToString(bridgeTests.CallDataMock),

					hex.EncodeToString([]byte("from2")),
					hex.EncodeToString([]byte("to2")),
					hex.EncodeToString([]byte("converted_token2")),
					hex.EncodeToString(big.NewInt(4).Bytes()),
					hex.EncodeToString(big.NewInt(3).Bytes()),
					hex.EncodeToString([]byte{bridgeCore.MissingDataProtocolMarker}),
				}

				assert.Equal(t, expectedArgs, vmRequest.Args)

				return &models.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{big.NewInt(1234).Bytes()},
					},
				}, nil
			},
		}

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetActionIDForProposeTransfer(context.Background(), batch)
		assert.Equal(t, uint64(1234), result)
		assert.Nil(t, err)
		assert.True(t, proxyCalled)
	})
}

func TestKlvClientDataGetter_WasProposedSetStatus(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.WasProposedSetStatus(context.Background(), nil)
		assert.False(t, result)
		assert.Equal(t, clients.ErrNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		proxyCalled := false
		batch := createMockBatch()
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				proxyCalled = true
				assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
				assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
				assert.Equal(t, 0, len(vmRequest.CallValue))
				assert.Equal(t, wasSetCurrentTransactionBatchStatusActionProposedFuncName, vmRequest.FuncName)

				expectedArgs := []string{
					hex.EncodeToString(big.NewInt(112233).Bytes()),
				}
				for _, stat := range batch.Statuses {
					expectedArgs = append(expectedArgs, hex.EncodeToString([]byte{stat}))
				}

				assert.Equal(t, expectedArgs, vmRequest.Args)

				return &models.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{{1}},
					},
				}, nil
			},
		}

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.WasProposedSetStatus(context.Background(), batch)
		assert.True(t, result)
		assert.Nil(t, err)
		assert.True(t, proxyCalled)
	})
}

func TestKlvClientDataGetter_GetTransactionsStatuses(t *testing.T) {
	t.Parallel()

	batchID := uint64(112233)
	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		expectedErr := errors.New("expected error")
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetTransactionsStatuses(context.Background(), batchID)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("empty response", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.Proxy = createMockProxy(make([][]byte, 0))

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetTransactionsStatuses(context.Background(), batchID)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, errNoStatusForBatchID))
		assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("for batch ID %d", batchID)))
	})
	t.Run("malformed batch finished status", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.Proxy = createMockProxy([][]byte{{56}})

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetTransactionsStatuses(context.Background(), batchID)
		assert.Nil(t, result)
		expectedErr := bridgeErrors.NewQueryResponseError(internalError, `error converting the received bytes to bool, strconv.ParseBool: parsing "56": invalid syntax`,
			"getStatusesAfterExecution", "klv1qqqqqqqqqqqqqpgqh46r9zh78lry2py8tq723fpjdr4pp0zgsg8syf6mq0")
		assert.Equal(t, expectedErr, err)
	})
	t.Run("batch not finished", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.Proxy = createMockProxy([][]byte{{0}})

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetTransactionsStatuses(context.Background(), batchID)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, errBatchNotFinished))
	})
	t.Run("missing status", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.Proxy = createMockProxy([][]byte{{1}, {}})

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetTransactionsStatuses(context.Background(), batchID)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, errMalformedBatchResponse))
		assert.True(t, strings.Contains(err.Error(), "for result index 0"))
	})
	t.Run("batch finished without response", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.Proxy = createMockProxy([][]byte{{1}})

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetTransactionsStatuses(context.Background(), batchID)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, errMalformedBatchResponse))
		assert.True(t, strings.Contains(err.Error(), "status is finished, no results are given"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		proxyCalled := false
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				proxyCalled = true
				assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
				assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
				assert.Equal(t, 0, len(vmRequest.CallValue))
				assert.Equal(t, getStatusesAfterExecutionFuncName, vmRequest.FuncName)

				expectedArgs := []string{
					hex.EncodeToString(big.NewInt(int64(batchID)).Bytes()),
				}

				assert.Equal(t, expectedArgs, vmRequest.Args)

				return &models.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{{1}, {2}, {3}, {4}},
					},
				}, nil
			},
		}

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetTransactionsStatuses(context.Background(), batchID)
		assert.Equal(t, []byte{2, 3, 4}, result)
		assert.Nil(t, err)
		assert.True(t, proxyCalled)
	})

}

func TestKlvClientDataGetter_GetActionIDForSetStatusOnPendingTransfer(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetActionIDForSetStatusOnPendingTransfer(context.Background(), nil)
		assert.Zero(t, result)
		assert.Equal(t, clients.ErrNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		proxyCalled := false
		batch := createMockBatch()
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
				proxyCalled = true
				assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
				assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
				assert.Equal(t, 0, len(vmRequest.CallValue))
				assert.Equal(t, getActionIdForSetCurrentTransactionBatchStatusFuncName, vmRequest.FuncName)

				expectedArgs := []string{
					hex.EncodeToString(big.NewInt(112233).Bytes()),
				}
				for _, stat := range batch.Statuses {
					expectedArgs = append(expectedArgs, hex.EncodeToString([]byte{stat}))
				}

				assert.Equal(t, expectedArgs, vmRequest.Args)

				return &models.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{big.NewInt(1132).Bytes()},
					},
				}, nil
			},
		}

		dg, _ := NewKLVClientDataGetter(args)

		result, err := dg.GetActionIDForSetStatusOnPendingTransfer(context.Background(), batch)
		assert.Equal(t, uint64(1132), result)
		assert.Nil(t, err)
		assert.True(t, proxyCalled)
	})
}

func TestKlvClientDataGetter_QuorumReached(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	actionID := big.NewInt(112233)
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, quorumReachedFuncName, vmRequest.FuncName)

			expectedArgs := []string{hex.EncodeToString(actionID.Bytes())}
			assert.Equal(t, expectedArgs, vmRequest.Args)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{{1}},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.QuorumReached(context.Background(), actionID.Uint64())
	assert.Nil(t, err)
	assert.True(t, proxyCalled)
	assert.True(t, result)
}

func TestKlvClientDataGetter_GetLastExecutedEthBatchID(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	val := big.NewInt(45372)
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, getLastExecutedEthBatchIdFuncName, vmRequest.FuncName)
			assert.Nil(t, vmRequest.Args)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{val.Bytes()},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.GetLastExecutedEthBatchID(context.Background())
	assert.Nil(t, err)
	assert.True(t, proxyCalled)
	assert.Equal(t, val.Uint64(), result)
}

func TestKlvClientDataGetter_GetLastExecutedEthTxID(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	val := big.NewInt(45372)
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, getLastExecutedEthTxId, vmRequest.FuncName)
			assert.Nil(t, vmRequest.Args)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{val.Bytes()},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.GetLastExecutedEthTxID(context.Background())
	assert.Nil(t, err)
	assert.True(t, proxyCalled)
	assert.Equal(t, val.Uint64(), result)
}

func TestKlvClientDataGetter_WasSigned(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	actionID := big.NewInt(112233)
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, signedFuncName, vmRequest.FuncName)

			expectedArgs := []string{
				hex.EncodeToString(args.RelayerAddress.Bytes()),
				hex.EncodeToString(actionID.Bytes()),
			}
			assert.Equal(t, expectedArgs, vmRequest.Args)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{{1}},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.WasSigned(context.Background(), actionID.Uint64())
	assert.Nil(t, err)
	assert.True(t, proxyCalled)
	assert.True(t, result)
}

func TestKlvClientDataGetter_GetAllStakedRelayers(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	providedRelayers := [][]byte{[]byte("relayer1"), []byte("relayer2")}
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, getAllStakedRelayersFuncName, vmRequest.FuncName)

			assert.Nil(t, vmRequest.Args)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: providedRelayers,
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.GetAllStakedRelayers(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, providedRelayers, result)
}

func TestKlvClientDataGetter_GetAllKnownTokens(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	providedTokens := [][]byte{[]byte("tkn1"), []byte("tkn2")}
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			assert.Equal(t, getBech32Address(args.SafeContractAddress), vmRequest.Address)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, getAllKnownTokens, vmRequest.FuncName)

			assert.Nil(t, vmRequest.Args)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: providedTokens,
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.GetAllKnownTokens(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, providedTokens, result)
}

func TestKcClientDataGetter_GetShardCurrentNonce(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	expectedNonce := uint64(33443)
	t.Run("GetNetworkStatus errors", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.Proxy = &interactors.ProxyStub{
			GetShardOfAddressCalled: func(ctx context.Context, bech32Address string) (uint32, error) {
				return 0, nil
			},
			GetNetworkStatusCalled: func(ctx context.Context) (*models.NodeOverview, error) {
				return nil, expectedErr
			},
		}
		dg, _ := NewKLVClientDataGetter(args)

		nonce, err := dg.GetCurrentNonce(context.Background())
		assert.Equal(t, uint64(0), nonce)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("GetNetworkStatus returns nil, nil", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.Proxy = &interactors.ProxyStub{
			GetShardOfAddressCalled: func(ctx context.Context, bech32Address string) (uint32, error) {
				return 0, nil
			},
			GetNetworkStatusCalled: func(ctx context.Context) (*models.NodeOverview, error) {
				return nil, nil
			},
		}
		dg, _ := NewKLVClientDataGetter(args)

		nonce, err := dg.GetCurrentNonce(context.Background())
		assert.Equal(t, uint64(0), nonce)
		assert.Equal(t, errNilNodeStatusResponse, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsKLVClientDataGetter()
		args.Proxy = &interactors.ProxyStub{
			GetShardOfAddressCalled: func(ctx context.Context, bech32Address string) (uint32, error) {
				return 0, nil
			},
			GetNetworkStatusCalled: func(ctx context.Context) (*models.NodeOverview, error) {
				return &models.NodeOverview{
					Nonce: expectedNonce,
				}, nil
			},
		}
		dg, _ := NewKLVClientDataGetter(args)

		nonce, err := dg.GetCurrentNonce(context.Background())
		assert.Equal(t, expectedNonce, nonce)
		assert.Nil(t, err)
	})
}

func TestKcClientDataGetter_IsPaused(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, getBech32Address(args.MultisigContractAddress), vmRequest.Address)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, isPausedFuncName, vmRequest.FuncName)
			assert.Empty(t, vmRequest.Args)

			strResponse := "AQ=="
			response, _ := base64.StdEncoding.DecodeString(strResponse)
			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{response},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.IsPaused(context.Background())
	assert.Nil(t, err)
	assert.True(t, result)
	assert.True(t, proxyCalled)
}

func TestKcClientDataGetter_isMintBurnToken(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.SafeContractAddress), vmRequest.Address)
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, isMintBurnTokenFuncName, vmRequest.FuncName)
			assert.Equal(t, []string{"746f6b656e"}, vmRequest.Args)

			strResponse := "AQ=="
			response, _ := base64.StdEncoding.DecodeString(strResponse)
			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{response},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.isMintBurnToken(context.Background(), []byte("token"))
	assert.Nil(t, err)
	assert.True(t, result)
	assert.True(t, proxyCalled)
}

func TestKcClientDataGetter_isNativeToken(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.SafeContractAddress), vmRequest.Address)
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, isNativeTokenFuncName, vmRequest.FuncName)
			assert.Equal(t, []string{"746f6b656e"}, vmRequest.Args)

			strResponse := "AQ=="
			response, _ := base64.StdEncoding.DecodeString(strResponse)
			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{response},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.isNativeToken(context.Background(), []byte("token"))
	assert.Nil(t, err)
	assert.True(t, result)
	assert.True(t, proxyCalled)
}

func TestKcClientDataGetter_getTotalBalances(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	expectedAccumulatedBurnedTokens := big.NewInt(100)
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.SafeContractAddress), vmRequest.Address)
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, getTotalBalances, vmRequest.FuncName)
			assert.Equal(t, []string{"746f6b656e"}, vmRequest.Args)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{expectedAccumulatedBurnedTokens.Bytes()},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.getTotalBalances(context.Background(), []byte("token"))
	assert.Nil(t, err)
	assert.Equal(t, result, expectedAccumulatedBurnedTokens)
	assert.True(t, proxyCalled)
}

func TestKcClientDataGetter_getMintBalances(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	expectedAccumulatedMintedTokens := big.NewInt(100)
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.SafeContractAddress), vmRequest.Address)
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, getMintBalances, vmRequest.FuncName)
			assert.Equal(t, []string{"746f6b656e"}, vmRequest.Args)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{expectedAccumulatedMintedTokens.Bytes()},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.getMintBalances(context.Background(), []byte("token"))
	assert.Nil(t, err)
	assert.Equal(t, result, expectedAccumulatedMintedTokens)
	assert.True(t, proxyCalled)
}

func TestKcClientDataGetter_getBurnBalances(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	expectedAccumulatedBurnedTokens := big.NewInt(100)
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.SafeContractAddress), vmRequest.Address)
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, getBurnBalances, vmRequest.FuncName)
			assert.Equal(t, []string{"746f6b656e"}, vmRequest.Args)

			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{expectedAccumulatedBurnedTokens.Bytes()},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.getBurnBalances(context.Background(), []byte("token"))
	assert.Nil(t, err)
	assert.Equal(t, result, expectedAccumulatedBurnedTokens)
	assert.True(t, proxyCalled)
}

func TestKcClientDataGetter_GetLastKcBatchID(t *testing.T) {
	t.Parallel()

	args := createMockArgsKLVClientDataGetter()
	proxyCalled := false
	args.Proxy = &interactors.ProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
			proxyCalled = true
			assert.Equal(t, getBech32Address(args.SafeContractAddress), vmRequest.Address)
			assert.Equal(t, getBech32Address(args.RelayerAddress), vmRequest.CallerAddr)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, getLastBatchId, vmRequest.FuncName)
			assert.Empty(t, vmRequest.Args)

			strResponse := "Dpk="
			response, _ := base64.StdEncoding.DecodeString(strResponse)
			return &models.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: [][]byte{response},
				},
			}, nil
		},
	}

	dg, _ := NewKLVClientDataGetter(args)

	result, err := dg.GetLastKcBatchID(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(3737), result)
	assert.True(t, proxyCalled)
}
