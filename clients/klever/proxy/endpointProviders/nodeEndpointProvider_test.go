package endpointProviders

import (
	"testing"

	"github.com/klever-io/klever-go/tools/check"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/stretchr/testify/assert"
)

func TestNewNodeEndpointProvider(t *testing.T) {
	t.Parallel()

	provider := NewNodeEndpointProvider()
	assert.False(t, check.IfNil(provider))
}

func TestNodeEndpointProvider_Getters(t *testing.T) {
	t.Parallel()

	provider := NewNodeEndpointProvider()
	assert.Equal(t, models.ObserverNode, provider.GetRestAPIEntityType())
}
