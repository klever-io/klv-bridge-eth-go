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

func TestBaseProxy_GetNetworkStatus(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	tests := []struct {
		name           string
		httpClientStub *testsCommon.HTTPClientWrapperStub
		expectedResult *models.NodeOverview
		expectedErr    error
	}{
		{
			name: "should receive expected error",
			httpClientStub: &testsCommon.HTTPClientWrapperStub{
				GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
					return nil, http.StatusBadRequest, expectedErr
				},
			},
			expectedResult: nil,
			expectedErr:    expectedErr,
		},
		{
			name: "malformed response - node endpoint provider",
			httpClientStub: &testsCommon.HTTPClientWrapperStub{
				GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
					return []byte("malformed response"), http.StatusOK, nil
				},
			},
			expectedResult: nil,
			expectedErr:    errors.New("invalid character 'm'"),
		},
		{
			name: "response error - node endpoint provider",
			httpClientStub: &testsCommon.HTTPClientWrapperStub{
				GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
					resp := &models.NodeOverviewApiResponse{
						Data:  models.NodeOverviewResponseData{},
						Error: expectedErr.Error(),
						Code:  "",
					}
					respBytes, _ := json.Marshal(resp)
					return respBytes, http.StatusOK, nil
				},
			},
			expectedResult: nil,
			expectedErr:    expectedErr,
		},
		{
			name: "GetNodeStatus returns nil network status - node endpoint provider",
			httpClientStub: &testsCommon.HTTPClientWrapperStub{
				GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
					return getNodeStatusBytes(nil), http.StatusOK, nil
				},
			},
			expectedResult: nil,
			expectedErr:    ErrNilNetworkStatus,
		},
		{
			name: "should work",
			httpClientStub: &testsCommon.HTTPClientWrapperStub{
				GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
					providedNetworkStatus := &models.NodeOverview{
						EpochNumber:       2,
						Nonce:             3,
						NonceAtEpochStart: 4,
					}
					return getNodeStatusBytes(providedNetworkStatus), http.StatusOK, nil
				},
			},
			expectedResult: &models.NodeOverview{
				EpochNumber:       2,
				Nonce:             3,
				NonceAtEpochStart: 4,
			},
			expectedErr: nil,
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
