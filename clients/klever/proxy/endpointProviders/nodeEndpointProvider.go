package endpointProviders

import (
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
)

// nodeEndpointProvider is suitable to work with a MultiversX node (observer)
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

// IsInterfaceNil returns true if there is no value under the interface
func (node *nodeEndpointProvider) IsInterfaceNil() bool {
	return node == nil
}
