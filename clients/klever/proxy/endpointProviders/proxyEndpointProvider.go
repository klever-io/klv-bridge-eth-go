package endpointProviders

import (
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
)

const (
	proxyVmQuery = "sc/query"
)

// proxyEndpointProvider is suitable to work with a MultiversX Proxy
type proxyEndpointProvider struct {
	*baseEndpointProvider
}

// NewProxyEndpointProvider returns a new instance of a proxyEndpointProvider
func NewProxyEndpointProvider() *proxyEndpointProvider {
	return &proxyEndpointProvider{
		baseEndpointProvider: &baseEndpointProvider{},
	}
}

// GetRestAPIEntityType returns the proxy constant
func (proxy *proxyEndpointProvider) GetRestAPIEntityType() models.RestAPIEntityType {
	return models.Proxy
}

// GetVmQuery returns the proxy path constant for VM query endpoint
func (proxy *proxyEndpointProvider) GetVmQuery() string {
	return proxyVmQuery
}

// IsInterfaceNil returns true if there is no value under the interface
func (proxy *proxyEndpointProvider) IsInterfaceNil() bool {
	return proxy == nil
}
