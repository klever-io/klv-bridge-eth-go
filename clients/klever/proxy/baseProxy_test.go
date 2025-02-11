package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klever-go/tools/check"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/endpointProviders"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsBaseProxy() argsBaseProxy {
	return argsBaseProxy{
		httpClientWrapper: &testsCommon.HTTPClientWrapperStub{},
		expirationTime:    time.Second,
		endpointProvider:  endpointProviders.NewNodeEndpointProvider(),
	}
}

func TestNewBaseProxy(t *testing.T) {
	t.Parallel()

	t.Run("nil http client wrapper", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = nil
		baseProxyInstance, err := newBaseProxy(args)

		assert.True(t, check.IfNil(baseProxyInstance))
		assert.True(t, errors.Is(err, ErrNilHTTPClientWrapper))
	})
	t.Run("invalid caching duration", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.expirationTime = time.Second - time.Nanosecond
		baseProxyInstance, err := newBaseProxy(args)

		assert.True(t, check.IfNil(baseProxyInstance))
		assert.True(t, errors.Is(err, ErrInvalidCacherDuration))
	})
	t.Run("nil endpoint provider", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.endpointProvider = nil
		baseProxyInstance, err := newBaseProxy(args)

		assert.True(t, check.IfNil(baseProxyInstance))
		assert.True(t, errors.Is(err, ErrNilEndpointProvider))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		baseProxyInstance, err := newBaseProxy(args)

		assert.False(t, check.IfNil(baseProxyInstance))
		assert.Nil(t, err)
	})
}

func TestBaseProxy_GetNetworkConfig(t *testing.T) {
	t.Parallel()

	expectedReturnedNetworkConfig := &models.NetworkConfig{
		ChainID:            "test",
		ConsensusGroupSize: 1,
		NumMetachainNodes:  7,
		SlotDuration:       4000,
		SlotsPerEpoch:      20,
		StartTime:          12,
	}
	response := &models.NetworkConfigResponse{
		Data:  expectedReturnedNetworkConfig,
		Error: "",
		Code:  "",
	}
	networkConfigBytes, _ := json.Marshal(response)

	t.Run("cache time expired", func(t *testing.T) {
		t.Parallel()

		mockWrapper := &testsCommon.HTTPClientWrapperStub{}
		wasCalled := false
		mockWrapper.GetHTTPCalled = func(ctx context.Context, endpoint string) ([]byte, int, error) {
			wasCalled = true
			return networkConfigBytes, http.StatusOK, nil
		}

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = mockWrapper
		args.expirationTime = minimumCachingInterval * 2
		baseProxyInstance, _ := newBaseProxy(args)
		baseProxyInstance.sinceTimeHandler = func(t time.Time) time.Duration {
			return minimumCachingInterval
		}

		configs, err := baseProxyInstance.GetNetworkConfig(context.Background())

		require.Nil(t, err)
		require.True(t, wasCalled)
		assert.Equal(t, expectedReturnedNetworkConfig, configs)
	})
	t.Run("fetchedConfigs is nil", func(t *testing.T) {
		t.Parallel()

		mockWrapper := &testsCommon.HTTPClientWrapperStub{}
		wasCalled := false
		mockWrapper.GetHTTPCalled = func(ctx context.Context, endpoint string) ([]byte, int, error) {
			wasCalled = true
			return networkConfigBytes, http.StatusOK, nil
		}

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = mockWrapper
		args.expirationTime = minimumCachingInterval * 2
		baseProxyInstance, _ := newBaseProxy(args)
		baseProxyInstance.sinceTimeHandler = func(t time.Time) time.Duration {
			return minimumCachingInterval*2 + time.Millisecond
		}

		configs, err := baseProxyInstance.GetNetworkConfig(context.Background())

		require.Nil(t, err)
		require.True(t, wasCalled)
		assert.Equal(t, expectedReturnedNetworkConfig, configs)
	})
	t.Run("Proxy.GetNetworkConfig returns error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		mockWrapper := &testsCommon.HTTPClientWrapperStub{}
		wasCalled := false
		mockWrapper.GetHTTPCalled = func(ctx context.Context, endpoint string) ([]byte, int, error) {
			wasCalled = true
			return nil, http.StatusBadRequest, expectedErr
		}

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = mockWrapper
		baseProxyInstance, _ := newBaseProxy(args)

		configs, err := baseProxyInstance.GetNetworkConfig(context.Background())

		require.Nil(t, configs)
		require.True(t, wasCalled)
		assert.True(t, errors.Is(err, expectedErr))
		assert.True(t, strings.Contains(err.Error(), http.StatusText(http.StatusBadRequest)))
	})
	t.Run("and Proxy.GetNetworkConfig returns malformed data", func(t *testing.T) {
		t.Parallel()

		mockWrapper := &testsCommon.HTTPClientWrapperStub{}
		wasCalled := false
		mockWrapper.GetHTTPCalled = func(ctx context.Context, endpoint string) ([]byte, int, error) {
			wasCalled = true
			return []byte("malformed data"), http.StatusOK, nil
		}

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = mockWrapper
		baseProxyInstance, _ := newBaseProxy(args)

		configs, err := baseProxyInstance.GetNetworkConfig(context.Background())

		require.Nil(t, configs)
		require.True(t, wasCalled)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "invalid character"))
	})
	t.Run("and Proxy.GetNetworkConfig returns a response error", func(t *testing.T) {
		t.Parallel()

		errMessage := "error message"
		erroredResponse := &models.NetworkConfigResponse{
			Data:  &models.NetworkConfig{},
			Error: errMessage,
			Code:  "",
		}
		erroredNetworkConfigBytes, _ := json.Marshal(erroredResponse)

		mockWrapper := &testsCommon.HTTPClientWrapperStub{}
		wasCalled := false
		mockWrapper.GetHTTPCalled = func(ctx context.Context, endpoint string) ([]byte, int, error) {
			wasCalled = true
			return erroredNetworkConfigBytes, http.StatusOK, nil
		}

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = mockWrapper
		baseProxyInstance, _ := newBaseProxy(args)

		configs, err := baseProxyInstance.GetNetworkConfig(context.Background())

		require.Nil(t, configs)
		require.True(t, wasCalled)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), errMessage))
	})
	t.Run("getCachedConfigs returns valid fetchedConfigs", func(t *testing.T) {
		t.Parallel()

		mockWrapper := &testsCommon.HTTPClientWrapperStub{}
		wasCalled := false
		mockWrapper.GetHTTPCalled = func(ctx context.Context, endpoint string) ([]byte, int, error) {
			wasCalled = true
			return nil, http.StatusOK, nil
		}

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = mockWrapper
		args.expirationTime = minimumCachingInterval * 2
		baseProxyInstance, _ := newBaseProxy(args)
		baseProxyInstance.fetchedConfigs = expectedReturnedNetworkConfig
		baseProxyInstance.sinceTimeHandler = func(t time.Time) time.Duration {
			return minimumCachingInterval
		}

		configs, err := baseProxyInstance.GetNetworkConfig(context.Background())

		require.Nil(t, err)
		assert.False(t, wasCalled)
		assert.Equal(t, expectedReturnedNetworkConfig, configs)
	})
}

func TestBaseProxy_GetNetworkStatus(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("get errors", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return nil, http.StatusBadRequest, expectedErr
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		result, err := baseProxyInstance.GetNetworkStatus(context.Background())
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, expectedErr))
		assert.True(t, strings.Contains(err.Error(), http.StatusText(http.StatusBadRequest)))
	})
	t.Run("malformed response - node endpoint provider", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return []byte("malformed response"), http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		result, err := baseProxyInstance.GetNetworkStatus(context.Background())
		assert.Nil(t, result)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "invalid character 'm'"))
	})
	t.Run("malformed response - proxy endpoint provider", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.endpointProvider = endpointProviders.NewProxyEndpointProvider()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return []byte("malformed response"), http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		result, err := baseProxyInstance.GetNetworkStatus(context.Background())
		assert.Nil(t, result)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "invalid character 'm'"))
	})
	t.Run("response error - node endpoint provider", func(t *testing.T) {
		t.Parallel()

		resp := &data.NodeStatusResponse{
			Data: struct {
				Status *data.NetworkStatus `json:"metrics"`
			}{},
			Error: expectedErr.Error(),
			Code:  "",
		}
		respBytes, _ := json.Marshal(resp)

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return respBytes, http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		result, err := baseProxyInstance.GetNetworkStatus(context.Background())
		assert.Nil(t, result)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
	})
	t.Run("response error - proxy endpoint provider", func(t *testing.T) {
		t.Parallel()

		resp := &data.NetworkStatusResponse{
			Data: struct {
				Status *data.NetworkStatus `json:"status"`
			}{},
			Error: expectedErr.Error(),
			Code:  "",
		}
		respBytes, _ := json.Marshal(resp)

		args := createMockArgsBaseProxy()
		args.endpointProvider = endpointProviders.NewProxyEndpointProvider()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return respBytes, http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		result, err := baseProxyInstance.GetNetworkStatus(context.Background())
		assert.Nil(t, result)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
	})
	t.Run("GetNodeStatus returns nil network status - node endpoint provider", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return getNetworkStatusBytes(nil), http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		result, err := baseProxyInstance.GetNetworkStatus(context.Background())
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, ErrNilNetworkStatus))
	})
	t.Run("GetNodeStatus returns nil network status - proxy endpoint provider", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.endpointProvider = endpointProviders.NewProxyEndpointProvider()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return getNetworkStatusBytes(nil), http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		result, err := baseProxyInstance.GetNetworkStatus(context.Background())
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, ErrNilNetworkStatus))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedNetworkStatus := &models.NodeOverview{
			EpochNumber:       2,
			Nonce:             3,
			NonceAtEpochStart: 4,
		}

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return getNodeStatusBytes(providedNetworkStatus), http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		result, err := baseProxyInstance.GetNetworkStatus(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, providedNetworkStatus, result)
	})
	t.Run("should work with proxy endpoint provider", func(t *testing.T) {
		t.Parallel()

		providedNetworkStatus := &models.NodeOverview{
			EpochNumber:       2,
			Nonce:             3,
			NonceAtEpochStart: 4,
		}

		args := createMockArgsBaseProxy()
		args.endpointProvider = endpointProviders.NewProxyEndpointProvider()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return getNetworkStatusBytes(providedNetworkStatus), http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		result, err := baseProxyInstance.GetNetworkStatus(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, providedNetworkStatus, result)
	})
}

func getNetworkStatusBytes(status *models.NodeOverview) []byte {
	resp := &models.NodeOverviewApiResponse{
		Data: models.NodeOverviewResponse{
			NodeOverview: status,
		},
	}
	respBytes, _ := json.Marshal(resp)

	return respBytes
}

func getNodeStatusBytes(status *models.NodeOverview) []byte {
	resp := &models.NodeOverviewApiResponse{
		Data: models.NodeOverviewResponse{
			NodeOverview: status,
		},
	}
	respBytes, _ := json.Marshal(resp)

	return respBytes
}

func TestBaseProxy_GetRestAPIEntityType(t *testing.T) {
	t.Parallel()

	args := createMockArgsBaseProxy()
	baseProxyInstance, _ := newBaseProxy(args)

	assert.Equal(t, args.endpointProvider.GetRestAPIEntityType(), baseProxyInstance.GetRestAPIEntityType())
}

func TestBaseProxyInstance_ProcessTransactionStatus(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("proxy errors when calling the API endpoint - StatusNotFound", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return nil, http.StatusNotFound, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		txStatus, err := baseProxyInstance.ProcessTransactionStatus(context.Background(), "tx hash")
		assert.Equal(t, transaction.Transaction_Fail, txStatus)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "returned http status: 404")
		assert.Contains(t, err.Error(), "please make sure you run the proxy version v1.1.38 or higher")
	})
	t.Run("proxy errors when calling the API endpoint, internal error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return nil, http.StatusOK, expectedErr
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		txStatus, err := baseProxyInstance.ProcessTransactionStatus(context.Background(), "tx hash")
		assert.Equal(t, transaction.Transaction_Fail, txStatus)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("proxy errors when calling the API endpoint, internal error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return nil, http.StatusOK, expectedErr
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		txStatus, err := baseProxyInstance.ProcessTransactionStatus(context.Background(), "tx hash")
		assert.Equal(t, transaction.Transaction_Fail, txStatus)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("proxy returns a malformed response", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return []byte("not a correct buffer"), http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		txStatus, err := baseProxyInstance.ProcessTransactionStatus(context.Background(), "tx hash")
		assert.Equal(t, transaction.Transaction_Fail, txStatus)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid character")
	})
	t.Run("proxy returns a valid response but with an error", func(t *testing.T) {
		t.Parallel()

		response := &data.ProcessedTransactionStatus{
			Data: struct {
				ProcessedStatus string `json:"status"`
			}{},
			Error: expectedErr.Error(),
			Code:  "",
		}

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				buff, _ := json.Marshal(response)
				return buff, http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		txStatus, err := baseProxyInstance.ProcessTransactionStatus(context.Background(), "tx hash")
		assert.Equal(t, transaction.Transaction_Fail, txStatus)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		response := &data.ProcessedTransactionStatus{
			Data: struct {
				ProcessedStatus string `json:"status"`
			}{
				ProcessedStatus: transaction.Transaction_SUCCESS.String(),
			},
			Error: "",
			Code:  "",
		}

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				buff, _ := json.Marshal(response)
				return buff, http.StatusOK, nil
			},
		}
		baseProxyInstance, _ := newBaseProxy(args)

		txStatus, err := baseProxyInstance.ProcessTransactionStatus(context.Background(), "tx hash")
		assert.Equal(t, transaction.Transaction_SUCCESS, txStatus)
		assert.Nil(t, err)
	})
}
