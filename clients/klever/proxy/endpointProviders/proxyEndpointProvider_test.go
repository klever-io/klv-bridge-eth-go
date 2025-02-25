package endpointProviders

import (
	"testing"

	"github.com/klever-io/klever-go/tools/check"
	"github.com/stretchr/testify/assert"
)

func TestNewProxyEndpointProvider(t *testing.T) {
	t.Parallel()

	provider := NewProxyEndpointProvider()
	assert.False(t, check.IfNil(provider))
}

func TestProxyEndpointProvider_GetNodeStatus(t *testing.T) {
	t.Parallel()

	provider := NewProxyEndpointProvider()
	assert.Equal(t, "network/status/", provider.GetNodeStatus())
}
