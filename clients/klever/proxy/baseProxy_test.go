package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/klever-io/klever-go/tools/check"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/endpointProviders"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
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

func createGetHttpStub(expectedResponse []byte, expectedStatusCode int, expectedErr error) *testsCommon.HTTPClientWrapperStub {
	return &testsCommon.HTTPClientWrapperStub{
		GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
			return expectedResponse, expectedStatusCode, expectedErr
		},
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

	expectedReturnedNetworkConfig := models.NetworkConfigResponseData{
		NetworkConfig: &models.NetworkConfig{
			ChainID:            "test",
			ConsensusGroupSize: 1,
			NumMetachainNodes:  7,
			SlotDuration:       4000,
			SlotsPerEpoch:      20,
			StartTime:          12,
		},
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
		assert.Equal(t, expectedReturnedNetworkConfig.NetworkConfig, configs)
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
		assert.Equal(t, expectedReturnedNetworkConfig.NetworkConfig, configs)
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
			Data:  models.NetworkConfigResponseData{},
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
		baseProxyInstance.fetchedConfigs = expectedReturnedNetworkConfig.NetworkConfig
		baseProxyInstance.sinceTimeHandler = func(t time.Time) time.Duration {
			return minimumCachingInterval
		}

		configs, err := baseProxyInstance.GetNetworkConfig(context.Background())

		require.Nil(t, err)
		assert.False(t, wasCalled)
		assert.Equal(t, expectedReturnedNetworkConfig.NetworkConfig, configs)
	})
}

func TestBaseProxy_GetNetworkStatus(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	tests := []struct {
		name               string
		httpClientStub     *testsCommon.HTTPClientWrapperStub
		expectedResult     *models.NodeOverview
		expectedStatusCode int
		expectedErr        error
	}{
		{
			name:               "should receive expected error",
			httpClientStub:     createGetHttpStub(nil, http.StatusBadRequest, expectedErr),
			expectedStatusCode: http.StatusBadRequest,
			expectedResult:     nil,
			expectedErr:        expectedErr,
		},
		{
			name:               "malformed response - node endpoint provider",
			httpClientStub:     createGetHttpStub([]byte("malformed response"), http.StatusOK, nil),
			expectedStatusCode: http.StatusOK,
			expectedResult:     nil,
			expectedErr:        errors.New("invalid character 'm'"),
		},
		{
			name:               "response error - node endpoint provider",
			httpClientStub:     createGetHttpStub(getGenericResponseWithErrorMessage(expectedErr.Error()), http.StatusOK, nil),
			expectedStatusCode: http.StatusOK,
			expectedResult:     nil,
			expectedErr:        expectedErr,
		},
		{
			name:               "GetNodeStatus returns nil network status - node endpoint provider",
			httpClientStub:     createGetHttpStub(getNodeStatusBytes(nil), http.StatusOK, nil),
			expectedStatusCode: http.StatusOK,
			expectedResult:     nil,
			expectedErr:        ErrNilNetworkStatus,
		},
		{
			name: "should work",
			httpClientStub: createGetHttpStub(
				getNodeStatusBytes(&models.NodeOverview{
					EpochNumber:       2,
					Nonce:             3,
					NonceAtEpochStart: 4,
				}),
				http.StatusOK, nil),
			expectedResult: &models.NodeOverview{
				EpochNumber:       2,
				Nonce:             3,
				NonceAtEpochStart: 4,
			},
			expectedStatusCode: http.StatusOK,
			expectedErr:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := createMockArgsBaseProxy()
			args.httpClientWrapper = tt.httpClientStub
			baseProxyInstance, _ := newBaseProxy(args)

			result, err := baseProxyInstance.GetNetworkStatus(context.Background())
			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedErr != nil {
				assert.NotNil(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.expectedErr.Error()))
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func getGenericResponseWithErrorMessage(errorMessage string) []byte {
	resp := &models.GenericApiResponse{
		Error: errorMessage,
	}
	respBytes, _ := json.Marshal(resp)

	return respBytes
}

func getNodeStatusBytes(status *models.NodeOverview) []byte {
	resp := &models.NodeOverviewApiResponse{
		Data: models.NodeOverviewResponseData{
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
