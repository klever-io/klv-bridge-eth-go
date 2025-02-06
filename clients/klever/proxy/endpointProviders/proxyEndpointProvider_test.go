package endpointProviders

import (
	"testing"

	"github.com/klever-io/klever-go/tools/check"
	"github.com/multiversx/mx-chain-core-go/core"
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
	assert.Equal(t, "network/status/0", provider.GetNodeStatus(0))
	assert.Equal(t, "network/status/4294967295", provider.GetNodeStatus(core.MetachainShardId))
}
