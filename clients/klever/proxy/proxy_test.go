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

	idata "github.com/klever-io/klever-go/indexer/data"
	"github.com/klever-io/klever-go/tools/check"
	kleverAddress "github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	sdkHttp "github.com/multiversx/mx-sdk-go/core/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testHttpURL = "https://test.org"
const networkConfigEndpoint = "network/config"
const getNodeStatusEndpoint = "node/status"
const chainID = "chainID"

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
	currentNonce uint64,
) (*http.Response, bool, error) {

	handled := false
	url := req.URL.String()
	var response interface{}
	switch url {
	case fmt.Sprintf("%s/%s", testHttpURL, networkConfigEndpoint):
		handled = true
		response = models.NetworkConfigResponse{
			Data: &models.NetworkConfig{
				ChainID: chainID,
			},
		}

	case fmt.Sprintf("%s/%s", testHttpURL, getNodeStatusEndpoint):
		handled = true
		response = models.NodeOverviewResponse{
			Data: struct {
				NodeOverview *models.NodeOverview `json:"overview"`
			}{
				NodeOverview: &models.NodeOverview{
					Nonce: currentNonce,
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
			response, handled, err := handleRequestNetworkConfigAndStatus(req, 9170526)
			if handled {
				return response, err
			}

			account := models.AccountApiResponse{
				Data: models.ResponseAccount{
					AccountData: models.Account{
						AccountInfo: &idata.AccountInfo{
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
	require.Equal(t, uint64(61), tx.Data.Transaction.BlockNum)
	require.Equal(t, "klv12e0kqcvqsrayj8j0c4dqjyvnv4ep253m5anx4rfj4jeq34lxsg8s84ec9j", tx.Data.Transaction.Sender)
	require.Equal(t, uint64(1), tx.Data.Transaction.Nonce)
	require.Equal(t, int64(500000), tx.Data.Transaction.KAppFee)
	require.Equal(t, int64(1000000), tx.Data.Transaction.BandwidthFee)
	require.Equal(t, "success", tx.Data.Transaction.Status)
	require.Equal(t, "Ok", tx.Data.Transaction.ResultCode)
	require.Equal(t, uint32(1), tx.Data.Transaction.Version)
	require.Equal(t, "420420", tx.Data.Transaction.ChainID)
	require.Len(t, tx.Data.Transaction.Signature, 1)
}
