package endpointProviders

import (
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
)

const (
	nodeVmQuery = "vm/query"
)

// nodeEndpointProvider is suitable to work with a Klever Blockchain node (observer)
type nodeEndpointProvider struct {
	*baseEndpointProvider
}

// NewNodeEndpointProvider returns a new instance of a nodeEndpointProvider
func NewNodeEndpointProvider() *nodeEndpointProvider {
	return &nodeEndpointProvider{
		baseEndpointProvider: &baseEndpointProvider{},
	}
}

// GetRestAPIEntityType returns the observer node constant
func (node *nodeEndpointProvider) GetRestAPIEntityType() models.RestAPIEntityType {
	return models.ObserverNode
}

// GetVmQuery returns the proxy path constant for VM values endpoint
func (proxy *nodeEndpointProvider) GetVmQuery() string {
	return nodeVmQuery
}

// IsInterfaceNil returns true if there is no value under the interface
func (node *nodeEndpointProvider) IsInterfaceNil() bool {
	return node == nil
}
