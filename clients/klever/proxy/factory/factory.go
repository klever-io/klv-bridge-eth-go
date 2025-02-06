package factory

import (
	"fmt"

	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/endpointProviders"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
)

// CreateEndpointProvider creates a new instance of EndpointProvider
func CreateEndpointProvider(entityType models.RestAPIEntityType) (EndpointProvider, error) {
	switch entityType {
	case models.ObserverNode:
		return endpointProviders.NewNodeEndpointProvider(), nil
	case models.Proxy:
		return endpointProviders.NewProxyEndpointProvider(), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownRestAPIEntityType, entityType)
	}
}
