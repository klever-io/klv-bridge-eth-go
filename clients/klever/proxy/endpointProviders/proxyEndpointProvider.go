package endpointProviders

import (
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
)

const (
	proxyGetNodeStatus = "network/status/"
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

// GetNodeStatus returns the node status endpoint
func (proxy *proxyEndpointProvider) GetNodeStatus() string {
	return proxyGetNodeStatus
}

// GetRestAPIEntityType returns the proxy constant
func (proxy *proxyEndpointProvider) GetRestAPIEntityType() models.RestAPIEntityType {
	return models.Proxy
}

// IsInterfaceNil returns true if there is no value under the interface
func (proxy *proxyEndpointProvider) IsInterfaceNil() bool {
	return proxy == nil
}
