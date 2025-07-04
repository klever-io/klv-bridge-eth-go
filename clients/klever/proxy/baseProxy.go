package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/klever-io/klever-go/tools/check"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/factory"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("klever-go-sdk/blockchain")

const (
	minimumCachingInterval = time.Second
)

type argsBaseProxy struct {
	expirationTime    time.Duration
	httpClientWrapper httpClientWrapper
	endpointProvider  factory.EndpointProvider
}

type baseProxy struct {
	httpClientWrapper
	mut                 sync.RWMutex
	fetchedConfigs      *models.NetworkConfig
	lastFetchedTime     time.Time
	cacheExpiryDuration time.Duration
	sinceTimeHandler    func(t time.Time) time.Duration
	endpointProvider    factory.EndpointProvider
}

// newBaseProxy will create a base kc proxy with cache instance
func newBaseProxy(args argsBaseProxy) (*baseProxy, error) {
	err := checkArgsBaseProxy(args)
	if err != nil {
		return nil, err
	}

	return &baseProxy{
		httpClientWrapper:   args.httpClientWrapper,
		cacheExpiryDuration: args.expirationTime,
		endpointProvider:    args.endpointProvider,
		sinceTimeHandler:    since,
	}, nil
}

func checkArgsBaseProxy(args argsBaseProxy) error {
	if args.expirationTime < minimumCachingInterval {
		return fmt.Errorf("%w, provided: %v, minimum: %v", ErrInvalidCacherDuration, args.expirationTime, minimumCachingInterval)
	}
	if check.IfNil(args.httpClientWrapper) {
		return ErrNilHTTPClientWrapper
	}
	if check.IfNil(args.endpointProvider) {
		return ErrNilEndpointProvider
	}

	return nil
}

func since(t time.Time) time.Duration {
	return time.Since(t)
}

// GetNetworkConfig will return the cached network configs fetching new values and saving them if necessary
func (proxy *baseProxy) GetNetworkConfig(ctx context.Context) (*models.NetworkConfig, error) {
	proxy.mut.RLock()
	cachedConfigs := proxy.getCachedConfigs()
	proxy.mut.RUnlock()

	if cachedConfigs != nil {
		return cachedConfigs, nil
	}

	return proxy.cacheConfigs(ctx)
}

func (proxy *baseProxy) getCachedConfigs() *models.NetworkConfig {
	if proxy.sinceTimeHandler(proxy.lastFetchedTime) > proxy.cacheExpiryDuration {
		return nil
	}

	return proxy.fetchedConfigs
}

func (proxy *baseProxy) cacheConfigs(ctx context.Context) (*models.NetworkConfig, error) {
	proxy.mut.Lock()
	defer proxy.mut.Unlock()

	// maybe another parallel running go routine already did the fetching
	cachedConfig := proxy.getCachedConfigs()
	if cachedConfig != nil {
		return cachedConfig, nil
	}

	log.Debug("Network config not cached. caching...")
	configs, err := proxy.getNetworkConfigFromSource(ctx)
	if err != nil {
		return nil, err
	}

	proxy.lastFetchedTime = time.Now()
	proxy.fetchedConfigs = configs

	return configs, nil
}

// getNetworkConfigFromSource retrieves the network configuration from the proxy
func (proxy *baseProxy) getNetworkConfigFromSource(ctx context.Context) (*models.NetworkConfig, error) {
	buff, code, err := proxy.GetHTTP(ctx, proxy.endpointProvider.GetNetworkConfig())
	if err != nil || code != http.StatusOK {
		return nil, createHTTPStatusError(code, err)
	}

	response := &models.NetworkConfigResponse{}
	err = json.Unmarshal(buff, response)
	if err != nil {
		return nil, err
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return response.Data.NetworkConfig, nil
}

// GetNetworkStatus will return the network status of a provided shard
func (proxy *baseProxy) GetNetworkStatus(ctx context.Context) (*models.NodeOverview, error) {
	endpoint := proxy.endpointProvider.GetNodeStatus()
	buff, code, err := proxy.GetHTTP(ctx, endpoint)
	if err != nil || code != http.StatusOK {
		return nil, createHTTPStatusError(code, err)
	}

	response := &models.NodeOverviewApiResponse{}
	err = json.Unmarshal(buff, response)
	if err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	if response.Data.NodeOverview == nil {
		return nil, ErrNilNetworkStatus
	}

	return response.Data.NodeOverview, nil
}

// GetRestAPIEntityType returns the REST API entity type that this implementation works with
func (proxy *baseProxy) GetRestAPIEntityType() models.RestAPIEntityType {
	return proxy.endpointProvider.GetRestAPIEntityType()
}

// IsInterfaceNil returns true if there is no value under the interface
func (proxy *baseProxy) IsInterfaceNil() bool {
	return proxy == nil
}
