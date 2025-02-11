package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	kleverAddress "github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/multiversx/mx-chain-core-go/core/check"
	sdkHttp "github.com/multiversx/mx-sdk-go/core/http"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testHttpURL = "https://test.org"
const networkConfigEndpoint = "network/config"
const getNetworkStatusEndpoint = "network/status/%d"
const getNodeStatusEndpoint = "node/status"

type mockHTTPClient struct {
	doCalled func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doCalled != nil {
		return m.doCalled(req)
	}

	return nil, errors.New("not implemented")
}

func createMockClientRespondingBytes(responseBytes []byte) *mockHTTPClient {
	return &mockHTTPClient{
		doCalled: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				Body:       io.NopCloser(bytes.NewReader(responseBytes)),
				StatusCode: http.StatusOK,
			}, nil
		},
	}
}

func createMockClientRespondingBytesWithStatus(responseBytes []byte, status int) *mockHTTPClient {
	return &mockHTTPClient{
		doCalled: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				Body:       io.NopCloser(bytes.NewReader(responseBytes)),
				StatusCode: status,
			}, nil
		},
	}
}

func createMockClientRespondingError(err error) *mockHTTPClient {
	return &mockHTTPClient{
		doCalled: func(req *http.Request) (*http.Response, error) {
			return nil, err
		},
	}
}

func createMockArgsProxy(httpClient sdkHttp.Client) ArgsProxy {
	return ArgsProxy{
		ProxyURL:            testHttpURL,
		Client:              httpClient,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		AllowedDeltaToFinal: 1,
		CacheExpirationTime: time.Second,
		EntityType:          models.ObserverNode,
	}
}

func handleRequestNetworkConfigAndStatus(
	req *http.Request,
	numShards uint32,
	currentNonce uint64,
	highestNonce uint64,
) (*http.Response, bool, error) {

	handled := false
	url := req.URL.String()
	var response interface{}
	switch url {
	case fmt.Sprintf("%s/%s", testHttpURL, networkConfigEndpoint):
		handled = true
		response = data.NetworkConfigResponse{
			Data: struct {
				Config *data.NetworkConfig `json:"config"`
			}{
				Config: &data.NetworkConfig{
					NumShardsWithoutMeta: numShards,
				},
			},
		}

	case fmt.Sprintf("%s/%s", testHttpURL, getNodeStatusEndpoint):
		handled = true
		response = data.NodeStatusResponse{
			Data: struct {
				Status *data.NetworkStatus `json:"metrics"`
			}{
				Status: &data.NetworkStatus{
					Nonce:                currentNonce,
					HighestNonce:         highestNonce,
					ProbableHighestNonce: currentNonce,
					ShardID:              2,
				},
			},
			Error: "",
			Code:  "",
		}
	case fmt.Sprintf("%s/%s", testHttpURL, fmt.Sprintf(getNetworkStatusEndpoint, 2)):
		handled = true
		response = data.NetworkStatusResponse{
			Data: struct {
				Status *data.NetworkStatus `json:"status"`
			}{
				Status: &data.NetworkStatus{
					Nonce:                currentNonce,
					HighestNonce:         highestNonce,
					ProbableHighestNonce: currentNonce,
					ShardID:              2,
				},
			},
			Error: "",
			Code:  "",
		}
	}

	if !handled {
		return nil, handled, nil
	}

	buff, _ := json.Marshal(response)
	return &http.Response{
		Body:       io.NopCloser(bytes.NewReader(buff)),
		StatusCode: http.StatusOK,
	}, handled, nil
}

func TestNewProxy(t *testing.T) {
	t.Parallel()

	t.Run("invalid time cache should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsProxy(nil)
		args.CacheExpirationTime = time.Second - time.Nanosecond
		proxyInstance, err := NewProxy(args)

		assert.True(t, check.IfNil(proxyInstance))
		assert.True(t, errors.Is(err, ErrInvalidCacherDuration))
	})
	t.Run("invalid nonce delta should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsProxy(nil)
		args.FinalityCheck = true
		args.AllowedDeltaToFinal = 0
		proxyInstance, err := NewProxy(args)

		assert.True(t, check.IfNil(proxyInstance))
		assert.True(t, errors.Is(err, ErrInvalidAllowedDeltaToFinal))
	})
	t.Run("should work with finality check", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsProxy(nil)
		args.FinalityCheck = true
		proxyInstance, err := NewProxy(args)

		assert.False(t, check.IfNil(proxyInstance))
		assert.Nil(t, err)
	})
	t.Run("should work without finality check", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsProxy(nil)
		proxyInstance, err := NewProxy(args)

		assert.False(t, check.IfNil(proxyInstance))
		assert.Nil(t, err)
	})
}

func TestGetAccount(t *testing.T) {
	t.Parallel()

	numAccountQueries := uint32(0)
	httpClient := &mockHTTPClient{
		doCalled: func(req *http.Request) (*http.Response, error) {
			response, handled, err := handleRequestNetworkConfigAndStatus(req, 3, 9170526, 9170526)
			if handled {
				return response, err
			}

			account := models.AccountApiResponse{
				Data: models.ResponseAccount{
					AccountData: models.Account{
						AccountInfo: &models.AccountInfo{
							Nonce:   37,
							Balance: 10,
						},
					},
				},
			}

			accountBytes, _ := json.Marshal(account)
			atomic.AddUint32(&numAccountQueries, 1)
			return &http.Response{
				Body:       io.NopCloser(bytes.NewReader(accountBytes)),
				StatusCode: http.StatusOK,
			}, nil
		},
	}
	args := createMockArgsProxy(httpClient)
	args.FinalityCheck = true
	proxyInstance, _ := NewProxy(args)

	address, err := kleverAddress.NewAddress("klv1qqqqqqqqqqqqqpgqh46r9zh78lry2py8tq723fpjdr4pp0zgsg8syf6mq0")
	require.NoError(t, err)

	t.Run("nil address should error", func(t *testing.T) {
		t.Parallel()

		response, errGet := proxyInstance.GetAccount(context.Background(), nil)
		require.Equal(t, ErrNilAddress, errGet)
		require.Nil(t, response)
	})

	t.Run("should work and return account", func(t *testing.T) {
		account, errGet := proxyInstance.GetAccount(context.Background(), address)
		assert.NotNil(t, account)
		assert.Equal(t, uint64(37), account.Nonce)
		assert.Nil(t, errGet)
		assert.Equal(t, uint32(1), atomic.LoadUint32(&numAccountQueries))
	})
}

// func TestProxy_RequestTransactionCost(t *testing.T) {
// 	t.Parallel()

// 	responseBytes := []byte(`{"data":{"txGasUnits":24273810,"returnMessage":""},"error":"","code":"successful"}`)
// 	httpClient := createMockClientRespondingBytes(responseBytes)
// 	args := createMockArgsProxy(httpClient)
// 	ep, _ := NewProxy(args)

// 	tx := &transaction.FrontendTransaction{
// 		Nonce:    1,
// 		Value:    "50",
// 		Receiver: "erd1rh5ws22jxm9pe7dtvhfy6j3uttuupkepferdwtmslms5fydtrh5sx3xr8r",
// 		Sender:   "erd1rh5ws22jxm9pe7dtvhfy6j3uttuupkepferdwtmslms5fydtrh5sx3xr8r",
// 		Data:     []byte("hello"),
// 		ChainID:  "1",
// 		Version:  1,
// 		Options:  0,
// 	}
// 	txCost, err := ep.RequestTransactionCost(context.Background(), tx)
// 	require.Nil(t, err)
// 	require.Equal(t, &data.TxCostResponseData{
// 		TxCost:     24273810,
// 		RetMessage: "",
// 	}, txCost)
// }

func TestProxy_GetTransactionInfoWithResults(t *testing.T) {
	t.Parallel()

	txHash := "824933e032df87f25da6886d78186e306b2e31062a1b01c8918da10fe69b1c2f"
	responseBytes := []byte(`{"data":{"transaction":{"hash":"824933e032df87f25da6886d78186e306b2e31062a1b01c8918da10fe69b1c2f","blockNum":61,"sender":"klv12e0kqcvqsrayj8j0c4dqjyvnv4ep253m5anx4rfj4jeq34lxsg8s84ec9j","nonce":1,"timestamp":1739292140,"kAppFee":500000,"bandwidthFee":1000000,"status":"success","resultCode":"Ok","version":1,"chainID":"420420","signature":["6f22d23cfd70337cc97cd0153551ccd53bb72e16533723bff3831ddb4c73139f3683db203d2f8a095a845e77bab4319b69afc36ee7b3cfee84005c389de76203"],"searchOrder":0,"receipts":[{"cID":255,"signer":"klv12e0kqcvqsrayj8j0c4dqjyvnv4ep253m5anx4rfj4jeq34lxsg8s84ec9j","type":19,"typeString":"SignedBy","weight":"1"},{"assetId":"KLV","assetType":"Fungible","cID":0,"from":"klv12e0kqcvqsrayj8j0c4dqjyvnv4ep253m5anx4rfj4jeq34lxsg8s84ec9j","marketplaceId":"","orderId":"","to":"klv1mge94r8n3q44hcwu2tk9afgjcxcawmutycu0cwkap7m6jnktjlvq58355l","type":0,"typeString":"Transfer","value":10000000}],"contract":[{"type":0,"typeString":"TransferContractType","parameter":{"amount":10000000,"assetId":"KLV","assetType":{"collection":"KLV","type":"Fungible"},"toAddress":"klv1mge94r8n3q44hcwu2tk9afgjcxcawmutycu0cwkap7m6jnktjlvq58355l"}}]}}, "error":"","code":"successful"}`)
	httpClient := createMockClientRespondingBytes(responseBytes)
	args := createMockArgsProxy(httpClient)
	ep, _ := NewProxy(args)

	tx, err := ep.GetTransactionInfoWithResults(context.Background(), txHash)
	require.Nil(t, err)

	txBytes, _ := json.MarshalIndent(tx, "", " ")
	fmt.Println(string(txBytes))

	require.Equal(t, txHash, tx.Data.Transaction.Hash)
}

func TestProxy_ExecuteVmQuery(t *testing.T) {
	t.Parallel()

	responseBytes := []byte(`{"data":{"data":{"returnData":["MC41LjU="],"returnCode":"ok","returnMessage":"","gasRemaining":18446744073685949187,"gasRefund":0,"outputAccounts":{"0000000000000000050033bb65a91ee17ab84c6f8a01846ef8644e15fb76696a":{"address":"erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt","nonce":0,"balance":null,"balanceDelta":0,"storageUpdates":{},"code":null,"codeMetaData":null,"outputTransfers":[],"callType":0}},"deletedAccounts":[],"touchedAccounts":[],"logs":[]}},"error":"","code":"successful"}`)
	t.Run("no finality check", func(t *testing.T) {
		httpClient := createMockClientRespondingBytes(responseBytes)
		args := createMockArgsProxy(httpClient)
		ep, _ := NewProxy(args)

		response, err := ep.ExecuteVMQuery(context.Background(), &models.VmValueRequest{
			ScAddress:  "erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt",
			FuncName:   "version",
			CallerAddr: "erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt",
		})
		require.Nil(t, err)
		require.Equal(t, "0.5.5", string(response.Data.ReturnData[0]))
	})
	t.Run("with finality check, chain is stuck", func(t *testing.T) {
		httpClient := &mockHTTPClient{
			doCalled: func(req *http.Request) (*http.Response, error) {
				response, handled, err := handleRequestNetworkConfigAndStatus(req, 3, 9170528, 9170526)
				if handled {
					return response, err
				}

				assert.Fail(t, "should have not reached this point in which the VM query is actually requested")
				return nil, nil
			},
		}
		args := createMockArgsProxy(httpClient)
		args.FinalityCheck = true
		ep, _ := NewProxy(args)

		response, err := ep.ExecuteVMQuery(context.Background(), &models.VmValueRequest{
			ScAddress:  "erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt",
			FuncName:   "version",
			CallerAddr: "erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt",
		})

		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "shardID 2 is stuck"))
		assert.Nil(t, response)
	})
	t.Run("with finality check, invalid address", func(t *testing.T) {
		httpClient := &mockHTTPClient{
			doCalled: func(req *http.Request) (*http.Response, error) {
				response, handled, err := handleRequestNetworkConfigAndStatus(req, 3, 9170526, 9170526)
				if handled {
					return response, err
				}

				assert.Fail(t, "should have not reached this point in which the VM query is actually requested")
				return nil, nil
			},
		}
		args := createMockArgsProxy(httpClient)
		args.FinalityCheck = true
		ep, _ := NewProxy(args)

		response, err := ep.ExecuteVMQuery(context.Background(), &models.VmValueRequest{
			ScAddress:  "invalid",
			FuncName:   "version",
			CallerAddr: "erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt",
		})

		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "invalid bech32 string length 7"))
		assert.Nil(t, response)
	})
	t.Run("with finality check, should work", func(t *testing.T) {
		wasHandled := false
		httpClient := &mockHTTPClient{
			doCalled: func(req *http.Request) (*http.Response, error) {
				response, handled, err := handleRequestNetworkConfigAndStatus(req, 3, 9170526, 9170525)
				if handled {
					wasHandled = true
					return response, err
				}

				return &http.Response{
					Body:       io.NopCloser(bytes.NewReader(responseBytes)),
					StatusCode: http.StatusOK,
				}, nil
			},
		}
		args := createMockArgsProxy(httpClient)
		args.FinalityCheck = true
		ep, _ := NewProxy(args)

		response, err := ep.ExecuteVMQuery(context.Background(), &models.VmValueRequest{
			ScAddress:  "erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt",
			FuncName:   "version",
			CallerAddr: "erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt",
		})

		assert.True(t, wasHandled)
		require.Nil(t, err)
		require.Equal(t, "0.5.5", string(response.Data.ReturnData[0]))
	})
}

func TestElrondProxy_GetESDTTokenData(t *testing.T) {
	t.Parallel()

	token := "TKN-001122"
	expectedErr := errors.New("expected error")
	validAddress, err := kleverAddress.NewAddressFromBytes(bytes.Repeat([]byte("1"), 32))
	require.NoError(t, err)

	t.Run("nil address, should error", func(t *testing.T) {
		t.Parallel()

		httpClient := createMockClientRespondingBytes(make([]byte, 0))
		args := createMockArgsProxy(httpClient)
		ep, _ := NewProxy(args)

		tokenData, err := ep.GetESDTTokenData(context.Background(), nil, token)
		assert.Nil(t, tokenData)
		assert.Equal(t, ErrNilAddress, err)
	})
	t.Run("invalid address, should error", func(t *testing.T) {
		t.Parallel()

		httpClient := createMockClientRespondingBytes(make([]byte, 0))
		args := createMockArgsProxy(httpClient)
		ep, _ := NewProxy(args)

		address, err := kleverAddress.NewAddressFromBytes([]byte("invalid"))
		require.NoError(t, err)

		tokenData, err := ep.GetESDTTokenData(context.Background(), address, token)
		assert.Nil(t, tokenData)
		assert.Equal(t, ErrInvalidAddress, err)
	})
	t.Run("http client errors, should error", func(t *testing.T) {
		t.Parallel()

		httpClient := createMockClientRespondingError(expectedErr)
		args := createMockArgsProxy(httpClient)
		ep, _ := NewProxy(args)

		tokenData, err := ep.GetESDTTokenData(context.Background(), validAddress, token)
		assert.Nil(t, tokenData)
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("invalid status, should error", func(t *testing.T) {
		t.Parallel()

		httpClient := createMockClientRespondingBytesWithStatus(make([]byte, 0), http.StatusNotFound)
		args := createMockArgsProxy(httpClient)
		ep, _ := NewProxy(args)

		tokenData, err := ep.GetESDTTokenData(context.Background(), validAddress, token)
		assert.Nil(t, tokenData)
		assert.ErrorIs(t, err, ErrHTTPStatusCodeIsNotOK)
	})
	t.Run("invalid response bytes, should error", func(t *testing.T) {
		t.Parallel()

		httpClient := createMockClientRespondingBytes([]byte("invalid json"))
		args := createMockArgsProxy(httpClient)
		ep, _ := NewProxy(args)

		tokenData, err := ep.GetESDTTokenData(context.Background(), validAddress, token)
		assert.Nil(t, tokenData)
		assert.NotNil(t, err)
	})
	t.Run("response returned error, should error", func(t *testing.T) {
		t.Parallel()

		response := &data.ESDTFungibleResponse{
			Error: expectedErr.Error(),
		}
		responseBytes, _ := json.Marshal(response)

		httpClient := createMockClientRespondingBytes(responseBytes)
		args := createMockArgsProxy(httpClient)
		ep, _ := NewProxy(args)

		tokenData, err := ep.GetESDTTokenData(context.Background(), validAddress, token)
		assert.Nil(t, tokenData)
		assert.NotNil(t, err)
		assert.Equal(t, expectedErr.Error(), err.Error())
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		responseTokenData := &data.ESDTFungibleTokenData{
			TokenIdentifier: "identifier",
			Balance:         "balance",
			Properties:      "properties",
		}
		response := &data.ESDTFungibleResponse{
			Data: struct {
				TokenData *data.ESDTFungibleTokenData `json:"tokenData"`
			}{
				TokenData: responseTokenData,
			},
		}
		responseBytes, _ := json.Marshal(response)

		httpClient := createMockClientRespondingBytes(responseBytes)
		args := createMockArgsProxy(httpClient)
		ep, _ := NewProxy(args)

		tokenData, err := ep.GetESDTTokenData(context.Background(), validAddress, token)
		assert.NotNil(t, tokenData)
		assert.Nil(t, err)
		assert.Equal(t, responseTokenData, tokenData)
		assert.False(t, responseTokenData == tokenData) // pointer testing
	})
	t.Run("should work with query options", func(t *testing.T) {
		t.Parallel()

		responseTokenData := &data.ESDTFungibleTokenData{
			TokenIdentifier: "identifier",
			Balance:         "balance",
			Properties:      "properties",
		}
		response := &data.ESDTFungibleResponse{
			Data: struct {
				TokenData *data.ESDTFungibleTokenData `json:"tokenData"`
			}{
				TokenData: responseTokenData,
			},
		}
		responseBytes, _ := json.Marshal(response)
		expectedSuffix := "?blockHash=626c6f636b2068617368&blockNonce=3838&blockRootHash=626c6f636b20726f6f742068617368&hintEpoch=3939&onFinalBlock=true&onStartOfEpoch=3737"

		httpClient := &mockHTTPClient{
			doCalled: func(req *http.Request) (*http.Response, error) {
				assert.True(t, strings.HasSuffix(req.URL.String(), expectedSuffix))

				return &http.Response{
					Body:       io.NopCloser(bytes.NewReader(responseBytes)),
					StatusCode: http.StatusOK,
				}, nil
			},
		}
		args := createMockArgsProxy(httpClient)
		ep, _ := NewProxy(args)

		tokenData, err := ep.GetESDTTokenData(context.Background(), validAddress, token)
		assert.NotNil(t, tokenData)
		assert.Nil(t, err)
		assert.Equal(t, responseTokenData, tokenData)
		assert.False(t, responseTokenData == tokenData) // pointer testing
	})
}
