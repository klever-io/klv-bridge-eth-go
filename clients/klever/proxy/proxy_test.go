package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/klever-io/klever-go/data/transaction"
	idata "github.com/klever-io/klever-go/indexer/data"
	"github.com/klever-io/klever-go/tools/check"
	kleverAddress "github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	sdkHttp "github.com/klever-io/klv-bridge-eth-go/core/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Test URLs
	testHttpURL = "https://test.org"
	// Test address
	testAddress     = "klv1qqqqqqqqqqqqqpgqh46r9zh78lry2py8tq723fpjdr4pp0zgsg8syf6mq0"
	contractAddress = "klv1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpgm89z"
)

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

func createMockArgsProxy(httpClient sdkHttp.Client, entity models.RestAPIEntityType) ArgsProxy {
	return ArgsProxy{
		ProxyURL:            testHttpURL,
		Client:              httpClient,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		AllowedDeltaToFinal: 1,
		CacheExpirationTime: time.Second,
		EntityType:          entity,
	}
}

func createMockDoCalled(responseData interface{}, statusCode int, numQueries *uint32) *mockHTTPClient {
	return &mockHTTPClient{
		doCalled: func(req *http.Request) (*http.Response, error) {
			accountBytes, _ := json.Marshal(responseData)
			atomic.AddUint32(numQueries, 1)
			return &http.Response{
				Body:       io.NopCloser(bytes.NewReader(accountBytes)),
				StatusCode: statusCode,
			}, nil
		},
	}
}

func TestNewProxy(t *testing.T) {
	t.Parallel()

	t.Run("invalid time cache should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsProxy(nil, models.ObserverNode)
		args.CacheExpirationTime = time.Second - time.Nanosecond
		proxyInstance, err := NewProxy(args)

		assert.True(t, check.IfNil(proxyInstance))
		assert.True(t, errors.Is(err, ErrInvalidCacherDuration))
	})
	t.Run("invalid nonce delta should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsProxy(nil, models.ObserverNode)
		args.FinalityCheck = true
		args.AllowedDeltaToFinal = 0
		proxyInstance, err := NewProxy(args)

		assert.True(t, check.IfNil(proxyInstance))
		assert.True(t, errors.Is(err, ErrInvalidAllowedDeltaToFinal))
	})
	t.Run("should work with finality check", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsProxy(nil, models.ObserverNode)
		args.FinalityCheck = true
		proxyInstance, err := NewProxy(args)

		assert.False(t, check.IfNil(proxyInstance))
		assert.Nil(t, err)
	})
	t.Run("should work without finality check", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsProxy(nil, models.ObserverNode)
		proxyInstance, err := NewProxy(args)

		assert.False(t, check.IfNil(proxyInstance))
		assert.Nil(t, err)
	})
}

func TestGetAccount_ShouldFailCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		address     string
		entityType  models.RestAPIEntityType
		statusCode  int
		response    interface{}
		expectedErr error
	}{
		{
			name:        "should fail account not found from node",
			entityType:  models.ObserverNode,
			expectedErr: ErrNilAddress,
		},
		{
			name:        "should fail account not found from node",
			entityType:  models.ObserverNode,
			address:     testAddress,
			statusCode:  http.StatusNotFound,
			response:    nil,
			expectedErr: ErrHTTPStatusCodeIsNotOK,
		},
		{
			name:        "should fail account not found from proxy",
			entityType:  models.Proxy,
			address:     testAddress,
			statusCode:  http.StatusNotFound,
			response:    nil,
			expectedErr: ErrHTTPStatusCodeIsNotOK,
		},
		{
			name:        "should fail invalid json from proxy",
			address:     testAddress,
			entityType:  models.Proxy,
			statusCode:  http.StatusOK,
			response:    []byte(`{"data":{}`),
			expectedErr: fmt.Errorf("json: cannot unmarshal string into Go value of type models.AccountApiResponse"),
		},
		{
			name:        "should fail invalid json from node",
			address:     testAddress,
			entityType:  models.ObserverNode,
			statusCode:  http.StatusOK,
			response:    []byte(`{"data":{}`),
			expectedErr: fmt.Errorf("json: cannot unmarshal string into Go value of type models.AccountNodeResponse"),
		},
		{
			name:       "should fail data with error message from node",
			address:    testAddress,
			entityType: models.ObserverNode,
			statusCode: http.StatusOK,
			response:   models.AccountNodeResponse{Error: "expected error"},
		},
		{
			name:       "should fail data with error message from proxy",
			address:    testAddress,
			entityType: models.Proxy,
			statusCode: http.StatusOK,
			response:   models.AccountNodeResponse{Error: "expected error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			numAccountQueries := uint32(0)
			httpClient := createMockDoCalled(tt.response, tt.statusCode, &numAccountQueries)

			args := createMockArgsProxy(httpClient, tt.entityType)
			proxyInstance, _ := NewProxy(args)

			address, _ := kleverAddress.NewAddress(tt.address)

			account, errGet := proxyInstance.GetAccount(context.Background(), address)
			assert.Nil(t, account)
			assert.NotNil(t, errGet)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, errGet, tt.expectedErr.Error())
			}

			switch tt.response.(type) {
			case models.AccountNodeResponse:
				response := tt.response.(models.AccountNodeResponse)
				if response.Error != "" {
					assert.ErrorContains(t, errGet, response.Error)
				}
			case models.AccountApiResponse:
				response := tt.response.(models.AccountApiResponse)
				if response.Error != "" {
					assert.ErrorContains(t, errGet, response.Error)
				}
			}
		})
	}
}

func TestGetAccount_FromNode_ShouldWork(t *testing.T) {
	t.Parallel()

	address, err := kleverAddress.NewAddress(testAddress)
	require.NoError(t, err)

	numAccountQueries := uint32(0)
	httpClient := createMockDoCalled(models.AccountNodeResponse{
		Data: models.ResponseNodeAccount{
			AccountData: models.Account{
				Nonce:   37,
				Balance: 10,
			},
		},
	}, http.StatusOK, &numAccountQueries)

	args := createMockArgsProxy(httpClient, models.ObserverNode)
	proxyInstance, _ := NewProxy(args)

	account, errGet := proxyInstance.GetAccount(context.Background(), address)
	assert.NotNil(t, account)
	assert.Equal(t, uint32(1), atomic.LoadUint32(&numAccountQueries))
	assert.Nil(t, errGet)
	assert.Equal(t, uint64(37), account.Nonce)
}

func TestGetAccount_FromProxy_ShouldWork(t *testing.T) {
	t.Parallel()

	address, err := kleverAddress.NewAddress(testAddress)
	require.NoError(t, err)

	numAccountQueries := uint32(0)
	httpClient := createMockDoCalled(models.AccountApiResponse{
		Data: models.ResponseProxyAccount{
			AccountData: models.ProxyAccountData{
				AccountInfo: &idata.AccountInfo{
					Nonce:   37,
					Balance: 10,
				},
			},
		},
	}, http.StatusOK, &numAccountQueries)

	args := createMockArgsProxy(httpClient, models.Proxy)
	proxyInstance, _ := NewProxy(args)

	account, errGet := proxyInstance.GetAccount(context.Background(), address)
	assert.NotNil(t, account)
	assert.Equal(t, uint32(1), atomic.LoadUint32(&numAccountQueries))
	assert.Nil(t, errGet)
	assert.Equal(t, uint64(37), account.Nonce)
}

func TestProxy_GetTransactionInfoWithResults(t *testing.T) {
	t.Parallel()

	txHash := "824933e032df87f25da6886d78186e306b2e31062a1b01c8918da10fe69b1c2f"
	responseBytes := []byte(`{"data":{"transaction":{"hash":"824933e032df87f25da6886d78186e306b2e31062a1b01c8918da10fe69b1c2f","blockNum":61,"sender":"klv12e0kqcvqsrayj8j0c4dqjyvnv4ep253m5anx4rfj4jeq34lxsg8s84ec9j","nonce":1,"timestamp":1739292140,"kAppFee":500000,"bandwidthFee":1000000,"status":"success","resultCode":"Ok","version":1,"chainID":"420420","signature":["6f22d23cfd70337cc97cd0153551ccd53bb72e16533723bff3831ddb4c73139f3683db203d2f8a095a845e77bab4319b69afc36ee7b3cfee84005c389de76203"],"searchOrder":0,"receipts":[{"cID":255,"signer":"klv12e0kqcvqsrayj8j0c4dqjyvnv4ep253m5anx4rfj4jeq34lxsg8s84ec9j","type":19,"typeString":"SignedBy","weight":"1"},{"assetId":"KLV","assetType":"Fungible","cID":0,"from":"klv12e0kqcvqsrayj8j0c4dqjyvnv4ep253m5anx4rfj4jeq34lxsg8s84ec9j","marketplaceId":"","orderId":"","to":"klv1mge94r8n3q44hcwu2tk9afgjcxcawmutycu0cwkap7m6jnktjlvq58355l","type":0,"typeString":"Transfer","value":10000000}],"contract":[{"type":0,"typeString":"TransferContractType","parameter":{"amount":10000000,"assetId":"KLV","assetType":{"collection":"KLV","type":"Fungible"},"toAddress":"klv1mge94r8n3q44hcwu2tk9afgjcxcawmutycu0cwkap7m6jnktjlvq58355l"}}]}}, "error":"","code":"successful"}`)
	httpClient := createMockClientRespondingBytes(responseBytes)
	args := createMockArgsProxy(httpClient, models.Proxy)
	ep, _ := NewProxy(args)

	tx, err := ep.GetTransactionInfoWithResults(context.Background(), txHash)
	require.Nil(t, err)

	require.Equal(t, txHash, tx.Hash)
	require.Equal(t, uint64(61), tx.BlockNum)
	require.Equal(t, "klv12e0kqcvqsrayj8j0c4dqjyvnv4ep253m5anx4rfj4jeq34lxsg8s84ec9j", tx.Sender)
	require.Equal(t, uint64(1), tx.Nonce)
	require.Equal(t, int64(500000), tx.KAppFee)
	require.Equal(t, int64(1000000), tx.BandwidthFee)
	require.Equal(t, "success", tx.Status)
	require.Equal(t, "Ok", tx.ResultCode)
	require.Equal(t, uint32(1), tx.Version)
	require.Equal(t, "420420", tx.ChainID)
	require.Len(t, tx.Signature, 1)
}

func TestSendTransaction_ShouldWork(t *testing.T) {
	t.Parallel()

	txHash := "824933e032df87f25da6886d78186e306b2e31062a1b01c8918da10fe69b1c2f"
	response := models.SendTransactionResponse{
		Data: &models.SendTransactionData{
			TxHash: txHash,
		},
		Code: "successful",
	}

	responseBytes, _ := json.Marshal(response)
	httpClient := createMockClientRespondingBytes(responseBytes)
	args := createMockArgsProxy(httpClient, models.Proxy)
	ep, _ := NewProxy(args)

	addr, err := kleverAddress.NewAddress(testAddress)
	require.Nil(t, err)

	tx := transaction.NewBaseTransaction(addr.Bytes(), 10, nil, 0, 0)
	tx.SetChainID([]byte("420420"))

	responseHash, err := ep.SendTransaction(context.Background(), tx)
	require.Nil(t, err)
	require.Equal(t, txHash, responseHash)
}

func TestSendTransaction_FailCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		statusCode  int
		txSend      *transaction.Transaction
		response    interface{}
		expectedErr error
	}{
		{
			name:        "should fail with invalid json response",
			statusCode:  http.StatusOK,
			response:    []byte(`{"data":{}`),
			expectedErr: fmt.Errorf("json: cannot unmarshal string into Go value of type models.SendTransactionResponse"),
		},
		{
			name:        "should fail with error message in response",
			statusCode:  http.StatusOK,
			response:    models.SendTransactionResponse{Error: "expected error"},
			expectedErr: errors.New("expected error"),
		},
		{
			name:        "should fail with non-OK status code",
			statusCode:  http.StatusInternalServerError,
			response:    nil,
			expectedErr: ErrHTTPStatusCodeIsNotOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			numAccountQueries := uint32(0)
			httpClient := createMockDoCalled(tt.response, tt.statusCode, &numAccountQueries)
			args := createMockArgsProxy(httpClient, models.Proxy)
			ep, _ := NewProxy(args)

			responseHash, err := ep.SendTransaction(context.Background(), tt.txSend)
			assert.Empty(t, responseHash)
			assert.NotNil(t, err)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			}

			assert.Equal(t, uint32(1), atomic.LoadUint32(&numAccountQueries))
		})
	}
}

func TestSendTransactions_ShouldWork(t *testing.T) {
	t.Parallel()

	txHashes := []string{"824933e032df87f25da6886d78186e306b2e31062a1b01c8918da10fe69b1c2f"}
	response := models.SendBulkTransactionsResponse{
		Data: models.TxHashes{
			Hashes: txHashes,
		},
		Code: "successful",
	}

	responseBytes, _ := json.Marshal(response)
	httpClient := createMockClientRespondingBytes(responseBytes)
	args := createMockArgsProxy(httpClient, models.Proxy)
	ep, _ := NewProxy(args)

	responseHash, err := ep.SendTransactions(context.Background(), nil)
	require.Nil(t, err)
	require.Equal(t, txHashes, responseHash)
}

func TestSendTransactions_FailCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		statusCode  int
		txSend      []*transaction.Transaction
		response    interface{}
		expectedErr error
	}{
		{
			name:        "should fail with invalid json response",
			statusCode:  http.StatusOK,
			response:    []byte(`{"data":{}`),
			expectedErr: fmt.Errorf("json: cannot unmarshal string into Go value of type models.SendBulkTransactionsResponse"),
		},
		{
			name:        "should fail with error message in response",
			statusCode:  http.StatusOK,
			response:    models.SendBulkTransactionsResponse{Error: "expected error"},
			expectedErr: errors.New("expected error"),
		},
		{
			name:        "should fail with non-OK status code",
			statusCode:  http.StatusInternalServerError,
			response:    nil,
			expectedErr: ErrHTTPStatusCodeIsNotOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			numAccountQueries := uint32(0)
			httpClient := createMockDoCalled(tt.response, tt.statusCode, &numAccountQueries)
			args := createMockArgsProxy(httpClient, models.Proxy)
			ep, _ := NewProxy(args)

			responseHashes, err := ep.SendTransactions(context.Background(), tt.txSend)
			assert.Empty(t, responseHashes)
			assert.NotNil(t, err)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			}

			assert.Equal(t, uint32(1), atomic.LoadUint32(&numAccountQueries))
		})
	}
}

func TestProxy_ExecuteVmQuery(t *testing.T) {
	t.Parallel()

	validResponseBytes := []byte(`{"data":{"data":{"returnData":["MC41LjU="],"returnCode":"ok","returnMessage":"","gasRemaining":18446744073685949187,"gasRefund":0,"outputAccounts":{"0000000000000000050033bb65a91ee17ab84c6f8a01846ef8644e15fb76696a":{"address":"klv1qqqqqqqqqqqqqpgqu2jcktadaq8mmytwglc704yfv7rezv5usg8sgzuah3","nonce":0,"balance":null,"balanceDelta":0,"storageUpdates":{},"code":null,"codeMetaData":null,"outputTransfers":[],"callType":0}},"deletedAccounts":[],"touchedAccounts":[],"logs":[]}}}`)
	validProxyBytes := []byte(`{"data":{"returnData":["MC41LjU="],"returnCode":"ok","returnMessage":"","gasRemaining":18446744073685949187,"gasRefund":0,"outputAccounts":{"0000000000000000050033bb65a91ee17ab84c6f8a01846ef8644e15fb76696a":{"address":"klv1qqqqqqqqqqqqqpgqu2jcktadaq8mmytwglc704yfv7rezv5usg8sgzuah3","nonce":0,"balance":null,"balanceDelta":0,"storageUpdates":{},"code":null,"codeMetaData":null,"outputTransfers":[],"callType":0}},"deletedAccounts":[],"touchedAccounts":[],"logs":[]}}`)
	tests := []struct {
		name        string
		entityType  models.RestAPIEntityType
		address     string
		funcName    string
		callerAddr  string
		response    []byte
		expectedErr string
		expectedRes string
	}{
		{
			name:        "should work",
			address:     contractAddress,
			entityType:  models.ObserverNode,
			funcName:    "version",
			callerAddr:  contractAddress,
			response:    validResponseBytes,
			expectedErr: "",
			expectedRes: "0.5.5",
		},
		{
			name:        "should work from proxy",
			address:     contractAddress,
			entityType:  models.Proxy,
			funcName:    "version",
			callerAddr:  contractAddress,
			response:    validProxyBytes,
			expectedErr: "",
			expectedRes: "0.5.5",
		},
		{
			name:        "should fail, invalid address",
			entityType:  models.ObserverNode,
			address:     "invalid",
			funcName:    "version",
			callerAddr:  contractAddress,
			expectedErr: "invalid bech32 string length 7",
			expectedRes: "",
		},
		{
			name:        "should fail with invalid json response",
			entityType:  models.ObserverNode,
			address:     contractAddress,
			response:    []byte(`{"data":[]}`),
			expectedErr: "json: cannot unmarshal array into Go",
		},
		{
			name:        "should fail with invalid json response from proxy",
			entityType:  models.Proxy,
			address:     contractAddress,
			response:    []byte(`{"data":[]}`),
			expectedErr: "json: cannot unmarshal array into Go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := createMockClientRespondingBytes(tt.response)
			args := createMockArgsProxy(httpClient, tt.entityType)
			ep, _ := NewProxy(args)

			response, err := ep.ExecuteVMQuery(context.Background(), &models.VmValueRequest{
				Address:    tt.address,
				FuncName:   tt.funcName,
				CallerAddr: tt.callerAddr,
			})

			if tt.expectedErr != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, response)
			} else {
				require.Nil(t, err)
				require.Equal(t, tt.expectedRes, string(response.Data.ReturnData[0]))
			}
		})
	}
}
