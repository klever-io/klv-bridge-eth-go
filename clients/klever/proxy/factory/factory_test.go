package factory

import (
	"errors"
	"fmt"
	"testing"

	"github.com/klever-io/klever-go/tools/check"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateEndpointProvider(t *testing.T) {
	t.Parallel()

	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()

		provider, err := CreateEndpointProvider("unknown")
		assert.True(t, check.IfNil(provider))
		assert.True(t, errors.Is(err, ErrUnknownRestAPIEntityType))
	})
	t.Run("node type", func(t *testing.T) {
		t.Parallel()

		provider, err := CreateEndpointProvider(models.ObserverNode)
		assert.False(t, check.IfNil(provider))
		assert.Nil(t, err)
		assert.Equal(t, "*endpointProviders.nodeEndpointProvider", fmt.Sprintf("%T", provider))
	})
	t.Run("proxy type", func(t *testing.T) {
		t.Parallel()

		provider, err := CreateEndpointProvider(models.Proxy)
		assert.False(t, check.IfNil(provider))
		assert.Nil(t, err)
		assert.Equal(t, "*endpointProviders.proxyEndpointProvider", fmt.Sprintf("%T", provider))
	})
}
