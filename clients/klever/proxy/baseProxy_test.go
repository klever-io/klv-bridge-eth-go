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
	t.Run("response error - node endpoint provider", func(t *testing.T) {
		t.Parallel()

		resp := &models.NodeOverviewApiResponse{
			Data:  models.NodeOverviewResponseData{},
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
	t.Run("GetNodeStatus returns nil network status - node endpoint provider", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBaseProxy()
		args.httpClientWrapper = &testsCommon.HTTPClientWrapperStub{
			GetHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
				return getNodeStatusBytes(nil), http.StatusOK, nil
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
