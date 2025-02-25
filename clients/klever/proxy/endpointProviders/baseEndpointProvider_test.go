package endpointProviders

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseEndpointProvider(t *testing.T) {
	t.Parallel()

	base := &baseEndpointProvider{}
	assert.Equal(t, networkConfig, base.GetNetworkConfig())
	assert.Equal(t, "address/addressAsBech32", base.GetAccount("addressAsBech32"))
	assert.Equal(t, estimateTransactionFees, base.GetEstimateTransactionFees())
	assert.Equal(t, sendTransaction, base.GetSendTransaction())
	assert.Equal(t, sendMultipleTransactions, base.GetSendMultipleTransactions())
	assert.Equal(t, "transaction/hex/status", base.GetTransactionStatus("hex"))
	assert.Equal(t, "transaction/hex", base.GetTransactionInfo("hex"))
	assert.Equal(t, vmValues, base.GetVmValues())
	assert.Equal(t, "address/klv1address/kda/TKN-001122", base.GetKDATokenData("klv1address", "TKN-001122"))
}
